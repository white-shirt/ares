package tracer

import (
	"context"

	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/server/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Tracer struct {
	Options
}

// New constructs a new recovery interceptor.
func New(opts ...Option) *Tracer {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &Tracer{
		Options: options,
	}
}

// Default gets default metric interceptor.
func Default() *Tracer {
	return New()
}

// Func implements Middleware interface.
func (t *Tracer) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			return next(c)
		}
	}
}

func (t *Tracer) UnaryClientIntercept() grpc.UnaryClientInterceptor {
	return func(parentCtx context.Context, method string, req, rep interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx := metadata.NewOutgoingContext(parentCtx, metadata.Pairs(
			"DY-Client-ID", application.ID(),
			"DY-Client-UID", application.UUID(),
		))
		return invoker(ctx, method, req, rep, cc, opts...)
	}
}

func (t *Tracer) StreamClientIntercept() grpc.StreamClientInterceptor {
	return func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
		ctx := metadata.NewOutgoingContext(parentCtx, metadata.Pairs(
			"DY-Client-ID", application.ID(),
			"DY-Client-UID", application.UUID(),
		))
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (t *Tracer) Label() string {
	return "tracer"
}
