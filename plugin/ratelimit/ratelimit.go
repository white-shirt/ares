package ratelimit

import (
	"context"
	"errors"

	"github.com/juju/ratelimit"
	"github.com/sevenNt/ares/server/echo"
	"github.com/sevenNt/ares/server/yell"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// ErrRateLimitTooMany too many request error.
	ErrRateLimitTooMany = errors.New("[ratelimit] too many request")
)

// RateLimit rate limit
type RateLimit struct {
	Options
}

// New constructs a new rateLimit interceptor.
func New(opts ...Option) *RateLimit {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &RateLimit{
		Options: options,
	}
}

// Func implements HTTP Middleware interface.
func (r *RateLimit) Func() echo.MiddlewareFunc {
	rt := ratelimit.NewBucket(r.interval, r.capacity)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			if _, wait := rt.TakeMaxDuration(1, r.waitMaxDuration); !wait {
				return c.String(echo.StatusTooManyRequests, "too many requests")
			}

			return next(c)
		}
	}
}

// UnaryServerIntercept implements gRPC unary server interceptor interface
func (r *RateLimit) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	rt := ratelimit.NewBucket(r.interval, r.capacity)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, wait := rt.TakeMaxDuration(1, r.waitMaxDuration); !wait {
			return nil, status.Errorf(yell.CodeTooManyRequest, "too many request")
		}
		return handler(ctx, req)
	}
}

// StreamServerIntercept implements gRPC stream server interceptor interface
func (r *RateLimit) StreamServerIntercept() grpc.StreamServerInterceptor {
	rt := ratelimit.NewBucket(r.interval, r.capacity)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, wait := rt.TakeMaxDuration(1, r.waitMaxDuration); !wait {
			return status.Errorf(yell.CodeTooManyRequest, "too many request")

		}
		return handler(srv, ss)
	}
}

// Label implements HTTP Middleware interface.
func (r *RateLimit) Label() string {
	return "ratelimit"
}
