package app

import (
	"net/http"
	"time"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (app *Application) CreateSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := app.CookieStore.Get(c.Request(), "session")
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		c.Set("session", session)

		return next(c)
	}
}

func (app *Application) IfAlreadyLogined(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if app.alreadyLoggedIn(c) {
			return c.Redirect(http.StatusFound, "/home")
		}
		return next(c)
	}
}

func (app *Application) IfNotLogined(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !app.alreadyLoggedIn(c) {
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

func (app *Application) UpdateSpotifyTokenIfExpired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		if token == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, models.ErrTokenNotExists)
		}

		checkedToken, err := app.Authenticator.RefreshToken(c.Request().Context(), token)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		if checkedToken.AccessToken != token.AccessToken {
			if err := app.TokenStore.Update(c.Request().Context(), userID, checkedToken); err != nil {
				c.Logger().Error(err)
				return err
			}
		}

		return next(c)
	}
}

func (app *Application) UpdateTrackURIsIfAboutToExpire(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		playlistID := c.Param("playlist_id")
		update, err := app.shouldUpdatePresignedURIs(playlistID)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		if update {
			data, err := app.generatePresignedURIsForPlaylistTracks(playlistID)
			if err != nil {
				c.Logger().Error(err)
				return err
			}

			db := app.TrackStore.DB()
			return db.Transaction(func(tx *gorm.DB) error {
				trackStore := store.NewTrackStore(db)

				for key, uri := range data {
					if err := trackStore.Update(map[string]any{"track_uri": uri, "inserted_at": time.Now()}, "track_key = ?", key); err != nil {
						c.Logger().Error(err)
						return err
					}
				}

				return nil
			})
		}

		return next(c)
	}
}
