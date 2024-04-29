package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Application struct {
	CookieStore         *sessions.CookieStore
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectPath string
	Authenticator       *spotifyauth.Authenticator
	UserStore           store.UserStore
	PlaylistStore       store.PlaylistStore
	TokenStore          store.TokenStore
}

func NewApplication() *Application {
	db := createSQLDB()
	rc := createRedisClient()

	return &Application{
		CookieStore:         sessions.NewCookieStore([]byte(os.Getenv("SECRET"))),
		SpotifyClientID:     os.Getenv("CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectPath: os.Getenv("REDIRECT_PATH"),
		Authenticator: spotifyauth.New(
			spotifyauth.WithRedirectURL(fmt.Sprintf("http://localhost%s%s", os.Getenv("ADDR"), os.Getenv("REDIRECT_PATH"))),
			spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
			spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
			spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
		),
		UserStore:     store.NewUserStore(db),
		PlaylistStore: store.NewPlaylistStore(db),
		TokenStore:    store.NewTokenStore(rc, "oauth_tokens"),
	}
}

func createSQLDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("SQL_DB_ADDRESS")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func createRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		PoolSize: 10,
	})
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

func getTracksFromPlaylist(client *spotify.Client, playlistID string) ([]models.Track, error) {
	playlist, err := client.GetPlaylist(context.Background(), spotify.ID(playlistID))
	if err != nil {
		return nil, err
	}

	var tracks []models.Track
	for _, track := range playlist.Tracks.Tracks {
		var artists string

		for _, artist := range track.Track.SimpleTrack.Artists {
			artists = artists + " " + artist.Name
		}

		tracks = append(tracks, models.Track{
			Name:    track.Track.Name,
			Artists: artists,
		})
	}

	return tracks, nil
}
