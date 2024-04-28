package app

import (
	"context"
	"net/http"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zmb3/spotify/v2"
)

func (app *Application) Router() *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(app.createSessionMiddleware)
	e.Static("/assets", "./public")

	e.GET("/", ServeHTML("./public/login.html"))
	e.GET("/signup", ServeHTML("./public/signup.html"), app.IfAlreadyLogined)
	e.GET("/home", ServeHTML("./public/home.html"), app.IfAlreadyLogined)

	// e.GET("/auth", app.HandleAuth, app.IfAlreadyLogined)
	e.GET("/spotify-auth", app.HandleSpotifyAuth)
	e.GET(app.SpotifyRedirectPath, app.HandleSpotifyRedirect)

	return e
}

func ServeHTML(htmlPath string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(htmlPath)
	}
}

// func (app *Application) HandleAuth(c echo.Context) error {
// 	action := c.QueryParam("action")
// 	username := c.FormValue("username")

// 	switch action {
// 	case "signup":
// 		exists, err := app.UserStore.IsExists("username = ?", username)
// 		if err != nil {
// 			c.Logger().Error(err)
// 			return err
// 		}

// 		if exists {
// 			return echo.NewHTTPError(http.StatusBadRequest, models.ErrUserAlreadyExists)
// 		}

// 		setContext(c, map[string]any{})

// 		return c.Redirect(http.StatusSeeOther, "/spotify-auth")

// 	case "login":
// 		exists, err := app.UserStore.IsExists("username = ?", username)
// 		if err != nil {
// 			c.Logger().Error(err)
// 			return err
// 		}

// 		if !exists {
// 			return echo.NewHTTPError(http.StatusBadRequest, models.ErrUserNotExists)
// 		}

// 		return c.Redirect(http.StatusSeeOther, "/spotify-auth")

// 	default:
// 		return echo.NewHTTPError(http.StatusBadRequest, models.ErrInvalidAction)
// 	}
// }

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

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	// return c.Redirect(http.StatusSeeOther, url)
	if _, err = http.DefaultClient.Do(req); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleSpotifyRedirect(c echo.Context) error {
	defer func() {
		deleteContext(c, []string{"action", "state"})
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

	// return c.File("./public/home.html")
	return nil

	// fmt.Printf("USER: %+v\n", user.User)

	// playlist, err := client.GetPlaylist(context.Background(), "37i9dQZF1DZ06evO3gsacM")
	// if err != nil {
	// 	return err
	// }

	// for _, track := range playlist.Tracks.Tracks {
	// 	fmt.Printf("%+v\n", track)
	// }

	// track, err := client.GetTrack(context.Background(), "3zTcMQUVeOgNELB4rSATdG")

	// fmt.Printf("%+v\n", track)
	// fmt.Printf("%+v\n", track.Artists)

	// return nil
}
