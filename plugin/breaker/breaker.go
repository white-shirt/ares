package breaker

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/sevenNt/ares/server/echo"
	"github.com/sevenNt/ares/server/yell"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type TripFunc func(*Breaker) bool
type Breaker struct {
	Options
}

func New(opts ...Option) *Breaker {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &Breaker{
		Options: options,
	}
}

// Func implements Middleware interface.
func (b *Breaker) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			hystrix.Do(b.label, func() (err error) {
				return next(c)
			}, func(err error) error {
				return c.Status(echo.StatusTooManyRequests)
			})
			return nil
		}
	}
}

func (b *Breaker) UnaryClientIntercept() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return hystrix.Do(b.label, func() (err error) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}, func(err error) error {
			return status.Errorf(yell.CodeCircuitBreak, "circuit break, %s", err)
		})
	}
}

func (b *Breaker) StreamClientIntercept() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
		err = hystrix.Do(b.label, func() (err error) {
			cs, err = streamer(ctx, desc, cc, method, opts...)
			return
		}, func(err error) error {
			return status.Errorf(yell.CodeCircuitBreak, "circuit break, %s", err)
		})
		return
	}
}

// UnaryServerIntercept implements gRPC unary server interceptor interface
//func (r *Breaker) UnaryServerIntercept() grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
//		return handler(ctx, req)
//	}
//}

// StreamServerIntercept implements gRPC stream server interceptor interface
//func (r *Breaker) StreamServerIntercept() grpc.StreamServerInterceptor {
//	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
//		return handler(srv, ss)
//	}
//}

func (b *Breaker) Label() string {
	return "breaker"
}
