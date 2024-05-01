package app

import (
	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

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
