package mw

import (
	"github.com/sevenNt/ares/server/echo"
)

// BasicAuth defines the config for BasicAuth middleware.
type BasicAuth struct{}

// Func implements Middleware interface.
func (a BasicAuth) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return next(c)
		}
	}
}

// Auth defines custom auth.
type Auth func(*echo.Context) error

// Func implements Middleware interface.
func (a Auth) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if a(c) != nil {
				return c.Status(echo.StatusUnauthorized)
			}
			return next(c)
		}
	}
}
