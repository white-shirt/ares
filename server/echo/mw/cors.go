package mw

import (
	"strconv"
	"strings"

	"github.com/sevenNt/ares/server/echo"
)

// DefaultCORS default cors
var DefaultCORS = CORS{
	AllowCredentials: false,
	AllowHeaders:     []string{},
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{echo.GET, echo.POST, echo.HEAD, echo.PUT, echo.PATCH, echo.DELETE},
}

// CORS cors
type CORS struct {
	Base
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// Func impletements Middleware interface
func (m CORS) Func() echo.MiddlewareFunc {
	if len(m.AllowOrigins) == 0 {
		m.AllowOrigins = DefaultCORS.AllowOrigins
	}

	if len(m.AllowMethods) == 0 {
		m.AllowMethods = DefaultCORS.AllowMethods
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method == echo.OPTIONS {
				c.Response().Header().Add(echo.HeaderVary, echo.HeaderOrigin)
				c.Response().Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
				c.Response().Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
				c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, strings.Join(m.AllowOrigins, ","))
				c.Response().Header().Set(echo.HeaderAccessControlAllowMethods, strings.Join(m.AllowMethods, ","))
				if m.AllowCredentials {
					c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}
				if len(m.AllowHeaders) != 0 {
					if h := c.Request().Header().Get(echo.HeaderAccessControlRequestHeaders); h != "" {
						c.Response().Header().Set(echo.HeaderAccessControlAllowHeaders, h)
					}
				}
				if m.MaxAge > 0 {
					c.Response().Header().Set(echo.HeaderAccessControlMaxAge, strconv.Itoa(m.MaxAge))
				}
				return nil
			}

			c.Response().Header().Add(echo.HeaderVary, echo.HeaderOrigin)
			c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, strings.Join(m.AllowOrigins, ","))
			if m.AllowCredentials {
				c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}
			if len(m.ExposeHeaders) != 0 {
				c.Response().Header().Set(echo.HeaderAccessControlExposeHeaders, strings.Join(m.ExposeHeaders, ","))
			}
			return next(c)
		}
	}
}
