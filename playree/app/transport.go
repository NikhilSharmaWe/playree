package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/rabbitmq"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zmb3/spotify/v2"
)

func (app *Application) Router() *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(app.CreateSessionMiddleware)
	e.Static("/assets", "./public")

	e.GET("/", ServeFile("./public/login"), app.IfAlreadyLogined)
	e.GET("/signup", ServeFile("./public/signup"), app.IfAlreadyLogined)
	e.GET("/home", ServeFile("./public/home"), app.IfNotLogined)
	e.GET("/create_playlist", ServeFile("./public/create_playlist"), app.IfNotLogined)

	e.GET("/spotify-auth", app.HandleSpotifyAuth)
	e.GET(app.SpotifyRedirectPath, app.HandleSpotifyRedirect)
	e.GET("/logout", app.HandleLogout, app.IfNotLogined)
	e.GET("/playlist/:playlist_id", app.HandlePlaylist, app.IfNotLogined)

	e.GET("/start-processing", app.HandleCreatePlaylistProcess, app.IfNotLogined)
	e.GET("/send-playlist-data", app.HandlePlaylistData, app.IfNotLogined)

	e.POST("/create_playlist", app.HandleCreatePlaylist, app.IfNotLogined, app.UpdateSpotifyTokenIfExpired)

	return e
}

func ServeFile(path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(path)
	}
}

func (app *Application) HandleSpotifyAuth(c echo.Context) error {
	action := c.QueryParam("action")
	if action != "signup" && action != "login" {
		return echo.NewHTTPError(http.StatusBadRequest, models.ErrInvalidAction)
	}

	if err := setSession(c, map[string]any{"action": action}); err != nil {
		c.Logger().Error(err)
		return err
	}

	state := uuid.NewString()
	url := app.Authenticator.AuthURL(state)

	if err := setSession(c, map[string]any{"state": state}); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.Redirect(http.StatusSeeOther, url)
}

func (app *Application) HandleSpotifyRedirect(c echo.Context) error {
	defer func() {
		deleteFromSession(c, []string{"action", "state"})
	}()

	action, err := getContext(c, "action")
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	if action != "signup" && action != "login" {
		return echo.NewHTTPError(http.StatusBadRequest, models.ErrInvalidAction)
	}

	state, err := getContext(c, "state")
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	if c.FormValue("state") != state {
		return echo.NewHTTPError(http.StatusNotFound, models.ErrStateMismatch)
	}

	token, err := app.Authenticator.Token(context.Background(), state, c.Request())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	client := spotify.New(app.Authenticator.Client(context.Background(), token))

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	switch action {
	case "signup":
		exists, err := app.UserStore.IsExists("user_id = ?", user.ID)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		if exists {
			return echo.NewHTTPError(http.StatusBadRequest, models.ErrUserAlreadyExists)
		}

		if err := app.UserStore.Create(models.UserDBModel{
			UserID:   user.ID,
			Username: user.DisplayName,
		}); err != nil {
			c.Logger().Error(err)
			return err
		}

	case "login":
		exists, err := app.UserStore.IsExists("user_id = ?", user.ID)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		if !exists {
			return echo.NewHTTPError(http.StatusBadRequest, models.ErrUserNotExists)
		}
	default:
		return echo.NewHTTPError(http.StatusBadRequest, models.ErrInvalidAction)
	}

	if err := app.TokenStore.Save(context.Background(), user.ID, token); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := setSession(c,
		map[string]any{"user_id": user.ID, "authenticated": true},
	); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/home")
}

func (app *Application) HandleLogout(c echo.Context) error {
	userID, err := getContext(c, "user_id")

	if err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := clearSessionHandler(c); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := app.TokenStore.Delete(context.Background(), userID); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (app *Application) HandleCreatePlaylist(c echo.Context) error {
	playlistID := path.Base(c.FormValue("playlist_link"))
	if err := setSession(c, map[string]any{"playlist_id": playlistID}); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := c.File("./public/processing"); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleCreatePlaylistProcess(c echo.Context) error {
	var (
		upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		errCh           = make(chan error)
		playlistCreated = make(chan string)
	)

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	defer conn.Close()

	sendMessageToFrontend(conn, "setting up creating process")

	playlistID, err := getContext(c, "playlist_id")
	if err != nil {
		sendFailStatusToFrontend(conn)
		c.Logger().Error(err)
	}
	defer deleteFromSession(c, []string{"playlist_id"})

	userID, err := getContext(c, "user_id")
	if err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	token, err := app.TokenStore.Get(context.Background(), userID)
	if err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	spotifyClient := spotify.New(app.Authenticator.Client(context.Background(), token))

	defer func() {
		if err := app.updateTokenFromClientIfNeeded(token, spotifyClient, userID); err != nil {
			c.Logger().Error(err)
		}
	}()

	tracksData, playlistName, err := getNameAndTracksFromPlaylist(spotifyClient, playlistID)
	if err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	playreePlaylistID := uuid.NewString()

	createPlaylistReq := models.CreatePlaylistRequest{
		PlayreePlaylistID: playreePlaylistID,
		Tracks:            tracksData,
	}

	app.CreatePlaylistResponseChannel[playreePlaylistID] = make(chan models.RabbitMQCreatePlaylistResponse)
	defer delete(app.CreatePlaylistResponseChannel, playreePlaylistID)

	go func() {
		resp := <-app.CreatePlaylistResponseChannel[playreePlaylistID]
		if !resp.Success {
			errCh <- errors.New(resp.Error)
		} else {
			playlistCreated <- "playlist-created"
		}
	}()

	rabbitMQClient, err := rabbitmq.NewRabbitMQClient(app.PublishingConn)
	if err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	defer rabbitMQClient.Close()

	body, err := json.Marshal(createPlaylistReq)
	if err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	if err := rabbitMQClient.Send(context.Background(), "create-playlist", "create-playlist-request", amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		ReplyTo:      "create-playlist-response-" + app.RabbitMQInstanceID,
		DeliveryMode: amqp.Persistent,
	}); err != nil {
		c.Logger().Error(err)
		sendFailStatusToFrontend(conn)
		return err
	}

	sendMessageToFrontend(conn, "creating playlist")

	ticker := time.NewTicker(5 * time.Minute)

	select {
	case <-playlistCreated:
		if err := app.PlaylistStore.Create(models.PlaylistsDBModel{
			UserID:       userID,
			PlaylistID:   playreePlaylistID,
			PlaylistName: playlistName,
		}); err != nil {
			c.Logger().Error(err)
			sendFailStatusToFrontend(conn)
			return err
		}

		sendMessageToFrontend(conn, "playlist created")
		sendMessageToFrontend(conn, fmt.Sprintf("PLAYLIST URL:http://%s/playlist/%s", os.Getenv("ADDR"), playreePlaylistID))

		return nil

	case err := <-errCh:
		c.Logger().Error(err)
		return err
	case <-ticker.C:
		c.Logger().Error(models.ErrCreatePlaylistServiceTimeout)
		sendMessageToFrontend(conn, "TIMEOUT: creating playlist timeout")
		return echo.NewHTTPError(http.StatusRequestTimeout, models.ErrCreatePlaylistServiceTimeout)
	}
}

func (app *Application) HandlePlaylist(c echo.Context) error {
	playlistID := c.Param("playlist_id")
	if err := setSession(c, map[string]any{"playlist_id": playlistID}); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.File("./public/playlist")
}

func (app *Application) HandlePlaylistData(c echo.Context) error {
	_, err := getContext(c, "playlist_id")
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}
