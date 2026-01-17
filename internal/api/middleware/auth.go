package middleware

import (
	"crypto/subtle"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/jonosize/affiliate-platform/internal/config"
)

// BasicAuth creates a BasicAuth middleware using credentials from config
func BasicAuth() echo.MiddlewareFunc {
	return echomw.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		cfg := config.Get()
		expectedUsername := cfg.GetBasicAuthUsername()
		expectedPassword := cfg.GetBasicAuthPassword()

		// Use constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare(
			[]byte(username),
			[]byte(expectedUsername),
		) == 1

		passwordMatch := subtle.ConstantTimeCompare(
			[]byte(password),
			[]byte(expectedPassword),
		) == 1

		return usernameMatch && passwordMatch, nil
	})
}
