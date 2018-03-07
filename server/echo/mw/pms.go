package mw

import (
	"github.com/sevenNt/ares/server/echo"
)

// DefaultPms is the default instance of Pms.
var DefaultPms = Pms{}

// Pms defines the config for Pms middleware.
type Pms struct {
	Base
	IsPermit    func(*echo.Context) error
	UnPermitted func(*echo.Context) error
}

// Func implements Middleware interface.
func (p Pms) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if e := p.IsPermit(c); e != nil {
				return p.UnPermitted(c)
			}
			return next(c)
		}
	}
}
