package chiwoo

import (
	"context"

	"google.golang.org/grpc"
)

type UnaryClientInterceptorFunc interface {
	UnaryClientIntercept() grpc.UnaryClientInterceptor
}

type SteamClientInterceptorFunc interface {
	StreamClientIntercept() grpc.StreamClientInterceptor
}

// UnaryClientInterceptorChain returns stream interceptors chain.
func UnaryClientInterceptorChain(interceptors ...UnaryClientInterceptorFunc) grpc.UnaryClientInterceptor {
	build := func(c grpc.UnaryClientInterceptor, invoker grpc.UnaryInvoker) grpc.UnaryInvoker {
		return func(ctx context.Context, method string, req, rep interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return c(ctx, method, req, rep, cc, invoker, opts...)
		}
	}
	//type UnaryInvoker func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) error
	return func(ctx context.Context, method string, req, rep interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		chain := invoker
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i].UnaryClientIntercept(), chain)
		}
		return chain(ctx, method, req, rep, cc, opts...)
	}
}

// StreamClientInterceptorChain returns stream interceptors chain.
func StreamClientInterceptorChain(interceptors ...SteamClientInterceptorFunc) grpc.StreamClientInterceptor {
	build := func(c grpc.StreamClientInterceptor, streamer grpc.Streamer) grpc.Streamer {
		return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return c(ctx, desc, cc, method, streamer, opts...)
		}
	}
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		chain := streamer
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i].StreamClientIntercept(), chain)
		}
		return chain(ctx, desc, cc, method, opts...)
	}
}
