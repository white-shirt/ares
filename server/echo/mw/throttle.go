package mw

import (
	"time"

	"github.com/juju/ratelimit"
	"github.com/sevenNt/ares/server/echo"
)

// DefaultThrottle is the default instance of Throttle.
var DefaultThrottle = Throttle{}

// Throttle defines the config for Throttle middleware.
type Throttle struct {
	Base
	FillInterval    time.Duration
	Capacity        int64
	Rate            float64
	WaitMaxDuration time.Duration
}

// Func implements Middleware interface.
func (t Throttle) Func() echo.MiddlewareFunc {
	rt := ratelimit.NewBucket(t.FillInterval, t.Capacity)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if _, wait := rt.TakeMaxDuration(1, t.WaitMaxDuration); !wait {
				return c.String(echo.StatusTooManyRequests, "too many requests")
			}

			return next(c)
		}
	}
}
