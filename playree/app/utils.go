package app

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/NikhilSharmaWe/rabbitmq"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Application struct {
	CookieStore *sessions.CookieStore
	Upgrader    websocket.Upgrader

	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectPath string
	Authenticator       *spotifyauth.Authenticator

	MinioClient     *minio.Client
	MinioBucketName string

	UserStore     store.UserStore
	PlaylistStore store.PlaylistStore
	TrackStore    store.TrackStore
	TokenStore    store.TokenStore

	CreatePlaylistResponseClient  *rabbitmq.RabbitClient
	CreatePlaylistResponseChannel map[string]chan models.RabbitMQCreatePlaylistResponse
	PublishingConn                *amqp.Connection
	RabbitMQInstanceID            string
}

func NewApplication() (*Application, error) {
	db := createSQLDB()

	rc := createRedisClient()

	minioServerAddr := os.Getenv("MINIO_SERVER_ADDR")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")

	client, err := minio.New(minioServerAddr, &minio.Options{
		Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
	})
	if err != nil {
		return nil, err
	}

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
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},

		SpotifyClientID:     os.Getenv("CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectPath: os.Getenv("REDIRECT_PATH"),
		Authenticator: spotifyauth.New(
			spotifyauth.WithRedirectURL(fmt.Sprintf("http://%s%s", os.Getenv("ADDR"), os.Getenv("REDIRECT_PATH"))),
			spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
			spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
			spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
		),

		MinioClient:     client,
		MinioBucketName: minioBucketName,

		UserStore:     store.NewUserStore(db),
		PlaylistStore: store.NewPlaylistStore(db),
		TrackStore:    store.NewTrackStore(db),
		TokenStore:    store.NewTokenStore(rc, "oauth_tokens"),

		CreatePlaylistResponseClient:  createPlaylistResponseClient,
		CreatePlaylistResponseChannel: make(map[string]chan models.RabbitMQCreatePlaylistResponse),
		PublishingConn:                publishingConnection,
		RabbitMQInstanceID:            instanceID,
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

func (app *Application) handleAfterPlaylistCreated(playlist *models.PlaylistsDBModel) error {
	db := app.TrackStore.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		playlistStore := store.NewPlaylistStore(db)
		tokenStore := store.NewTrackStore(db)

		if err := playlistStore.Create(*playlist); err != nil {
			return err
		}

		data, err := app.generatePresignedURIsForPlaylistTracks(playlist.PlaylistID)
		if err != nil {
			return err
		}

		tracks := []models.TrackDBModel{}
		for key, uri := range data {
			tracks = append(tracks, models.TrackDBModel{
				PlaylistID: playlist.PlaylistID,
				TrackKey:   key,
				TrackURI:   uri,
			})
		}

		return tokenStore.CreateInBatches(tracks)
	})
}

func (app *Application) generatePresignedURIsForPlaylistTracks(playreePlaylistID string) (map[string]string, error) {
	data := make(map[string]string)

	keys, err := app.getListOfAllFiles(playreePlaylistID)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		url, err := app.MinioClient.PresignedGetObject(context.Background(), "playree-playlists", key, 7*24*time.Hour, nil)
		if err != nil {
			return nil, err
		}

		data[key] = url.String()
	}

	return data, nil
}

func (app *Application) shouldUpdatePresignedURIs(playreePlaylistID string) (bool, error) {
	now := time.Now()
	fiveDaysAgo := now.AddDate(0, 0, -5)

	exists, err := app.TrackStore.IsExists("playlist_id = ? AND inserted_at <= ?", playreePlaylistID, fiveDaysAgo)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (app *Application) getListOfAllFiles(playreePlaylistID string) ([]string, error) {
	objKeys := []string{}

	objectCh := app.MinioClient.ListObjects(context.Background(), app.MinioBucketName, minio.ListObjectsOptions{
		Prefix:    playreePlaylistID,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return objKeys, object.Err
		}

		objKeys = append(objKeys, object.Key)
	}

	return objKeys, nil
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

func getNameAndTracksFromPlaylist(client *spotify.Client, playlistID string) ([]*models.Track, string, error) {
	data := []*models.Track{}

	playlist, err := client.GetPlaylist(context.Background(), spotify.ID(playlistID))
	if err != nil {
		return nil, "", err
	}

	for _, track := range playlist.Tracks.Tracks {
		artists := "$"

		for _, artist := range track.Track.SimpleTrack.Artists {
			artists = artists + artist.Name + ", "
		}
		artists = artists[:len(artists)-2] + "$"

		data = append(data, &models.Track{
			Name:    track.Track.Name,
			Artists: artists,
		})
	}

	return data, playlist.Name, nil
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

func sendMessageToFrontend(conn *websocket.Conn, msg string) {
	conn.WriteMessage(1, []byte(msg))
}

func sendFailStatusToFrontend(conn *websocket.Conn) {
	conn.WriteMessage(1, []byte("Error: creating process failed"))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
