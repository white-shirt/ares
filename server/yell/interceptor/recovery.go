package interceptor

import (
	"fmt"
	"runtime"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// DefaultRecovery is default metric interceptor.
var DefaultRecovery = Recovery{
	StackSize: 200,
}

// Recovery wraps sets of recovery.
type Recovery struct {
	Base
	StackSize int
}

// NewRecovery constructs a new recovery interceptor.
func NewRecovery(stacksize int) Recovery {
	return Recovery{
		StackSize: stacksize,
	}
}

// UnaryIntercept implements UnaryIntercept function of Interceptor.
func (recovery Recovery) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := make([]byte, recovery.StackSize)
				stack = stack[:runtime.Stack(stack, false)]

				// TODO 怎么样优美的处理错误？
				err = grpc.Errorf(codes.Internal, "panic error: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// StreamIntercept implements StreamIntercept function of Interceptor.
func (recovery Recovery) StreamServerIntercept() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := make([]byte, recovery.StackSize)
				stack = stack[:runtime.Stack(stack, false)]

				// TODO 怎么样优美的处理错误？
				err = grpc.Errorf(codes.Internal, "panic error: %v", r)
			}
		}()
		fmt.Println("recovery: ")
		return handler(srv, stream)
	}
}
