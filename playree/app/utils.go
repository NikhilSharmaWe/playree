package app

import (
	"context"
	"fmt"
	"os"

	"github.com/NikhilSharmaWe/rabbitmq"
	"github.com/google/uuid"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Application struct {
	CookieStore *sessions.CookieStore

	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectPath string
	Authenticator       *spotifyauth.Authenticator

	UserStore     store.UserStore
	PlaylistStore store.PlaylistStore
	TokenStore    store.TokenStore

	CreatePlaylistResponseClient  *rabbitmq.RabbitClient
	CreatePlaylistResponseChannel map[string]chan models.RabbitMQCreatePlaylistResponse
	PublishingConn                *amqp.Connection
	RabbitMQInstanceID            string
}

func NewApplication() (*Application, error) {
	db := createSQLDB()
	rc := createRedisClient()

	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQVhost := os.Getenv("RABBITMQ_VHOST")
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")

	// each concurrent task should be done with new channel
	// different connections should be used for publishing and consuming

	instanceID := uuid.NewString()

	consumingConnection, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	publishingConnection, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	_, err = rabbitmq.CreateNewQueueReturnClient(publishingConnection, "create-playlist-request", true, true)
	if err != nil {
		return nil, err
	}

	createPlaylistResponseClient, err := rabbitmq.CreateNewQueueReturnClient(consumingConnection, "create-playlist-response-"+instanceID, true, true)
	if err != nil {
		return nil, err
	}

	return &Application{
		CookieStore: sessions.NewCookieStore([]byte(os.Getenv("SECRET"))),

		SpotifyClientID:     os.Getenv("CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectPath: os.Getenv("REDIRECT_PATH"),
		Authenticator: spotifyauth.New(
			spotifyauth.WithRedirectURL(fmt.Sprintf("http://localhost%s%s", os.Getenv("ADDR"), os.Getenv("REDIRECT_PATH"))),
			spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
			spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
			spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
		),

		UserStore:                    store.NewUserStore(db),
		PlaylistStore:                store.NewPlaylistStore(db),
		TokenStore:                   store.NewTokenStore(rc, "oauth_tokens"),
		CreatePlaylistResponseClient: createPlaylistResponseClient,
		PublishingConn:               publishingConnection,
	}, nil
}

func (app *Application) updateTokenFromClientIfNeeded(token *oauth2.Token, client *spotify.Client, userID string) error {
	updatedToken, err := client.Token()
	if err != nil {
		return err
	}

	if updatedToken.AccessToken != token.AccessToken {
		if err := app.TokenStore.Update(context.Background(), userID, updatedToken); err != nil {
			return err
		}
	}

	return nil
}

func getTracksFromPlaylist(client *spotify.Client, playlistID string) ([]*models.Track, error) {
	data := []*models.Track{}

	playlist, err := client.GetPlaylist(context.Background(), spotify.ID(playlistID))
	if err != nil {
		return nil, err
	}

	for _, track := range playlist.Tracks.Tracks {
		var artists string

		for _, artist := range track.Track.SimpleTrack.Artists {
			artists = artists + " " + artist.Name
		}

		data = append(data, &models.Track{
			Name:    track.Track.Name,
			Artists: artists,
		})
	}

	return data, nil
}

func (app *Application) alreadyLoggedIn(c echo.Context) bool {
	session := c.Get("session").(*sessions.Session)

	userID, ok := session.Values["user_id"].(string)
	if !ok {
		return false
	}

	if exists, err := app.UserStore.IsExists("user_id = ?", userID); err != nil || !exists {
		return false
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	if ok && authenticated {
		return true
	}

	return false
}

func setSession(c echo.Context, keyValues map[string]any) error {
	session := c.Get("session").(*sessions.Session)
	for k, v := range keyValues {
		session.Values[k] = v
	}

	return session.Save(c.Request(), c.Response())
}

func getContext(c echo.Context, key string) (string, error) {
	session := c.Get("session").(*sessions.Session)
	v, ok := session.Values[key]
	if !ok {
		return "", models.ErrInvalidRequest
	}

	return v.(string), nil
}

func deleteFromSession(c echo.Context, keys []string) error {
	session := c.Get("session").(*sessions.Session)

	for _, k := range keys {
		delete(session.Values, k)
	}

	return session.Save(c.Request(), c.Response())
}

func clearSessionHandler(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.Options.MaxAge = -1
	return session.Save(c.Request(), c.Response())
}
