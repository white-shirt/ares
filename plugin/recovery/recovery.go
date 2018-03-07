package recovery

import (
	"fmt"
	"runtime"

	"github.com/sevenNt/wzap"
	"github.com/sevenNt/ares/server/echo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Recovery wraps sets of recovery.
type Recovery struct {
	Options
}

// New constructs a new recovery interceptor.
func New(opts ...Option) Recovery {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return Recovery{
		Options: options,
	}
}

// Default gets default metric interceptor.
func Default() Recovery {
	return New(StackSize(4096))
}

// UnaryServerIntercept implements UnaryIntercept function of Interceptor.
func (r Recovery) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if recover := recover(); recover != nil {
				stack := make([]byte, r.stacksize)
				stack = stack[:runtime.Stack(stack, false)]

				// TODO 怎么样优美的处理错误？
				err = grpc.Errorf(codes.Internal, "panic error: %v", recover)
			}
		}()
		return handler(ctx, req)
	}
}

// StreamServerIntercept implements StreamIntercept function of Interceptor.
func (r Recovery) StreamServerIntercept() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if recover := recover(); recover != nil {
				stack := make([]byte, r.stacksize)
				stack = stack[:runtime.Stack(stack, false)]

				// TODO 怎么样优美的处理错误？
				err = grpc.Errorf(codes.Internal, "panic error: %v", recover)
			}
		}()
		return handler(srv, stream)
	}
}

// Func implements Middleware interface.
func (r Recovery) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			defer func() {
				if rec := recover(); rec != nil {
					wzap.Error("panic: ", rec)
					var err error
					switch rec := rec.(type) {
					case error:
						err = rec
					default:
						err = fmt.Errorf("%v", rec)

					}
					stack := make([]byte, r.stacksize)
					length := runtime.Stack(stack, !r.disableStackAll)
					if !r.disablePrintStack {
						// TODO logger
						//c.Logger().Printf("[%s] %s %s", color.Red("PANIC RECOVER"), err, stack[:length])
						fmt.Printf("%s %s", err, stack[:length])
					}
					// TODO Error
					//c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

// Label implements Middleware interface.
func (r Recovery) Label() string {
	return "recovery"
}
