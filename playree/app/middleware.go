package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (app *Application) createSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
