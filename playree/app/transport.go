package app

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e.POST("/create_playlist", app.HandleCreatePlaylist, app.IfNotLogined)

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

	token, err := app.Authenticator.Token(c.Request().Context(), state, c.Request())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	client := spotify.New(app.Authenticator.Client(c.Request().Context(), token))

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

	if err := app.TokenStore.Delete(c.Request().Context(), userID); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (app *Application) HandleCreatePlaylist(c echo.Context) error {
	playlistID := path.Base(c.FormValue("playlist_link"))

	userID, err := getContext(c, "user_id")
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	token, err := app.TokenStore.Get(c.Request().Context(), userID)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	client := spotify.New(app.Authenticator.Client(c.Request().Context(), token))

	defer func() {
		updatedToken, err := client.Token()
		if err != nil {
			c.Logger().Error(err)
		}

		if updatedToken.AccessToken != token.AccessToken {
			if err := app.TokenStore.Update(c.Request().Context(), userID, updatedToken); err != nil {
				c.Logger().Error(err)
			}
		}
	}()

	tracks, err := getTracksFromPlaylist(client, playlistID)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	for _, t := range tracks {
		fmt.Printf("%+v\n", t)
	}

	return nil
}
