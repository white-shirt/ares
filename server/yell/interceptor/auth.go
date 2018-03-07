package interceptor

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// DefaultAuth is default auth interceptor.
var DefaultAuth = Auth{
	AuthFunc: func(ctx context.Context) (context.Context, error) {
		//metadata.MD[]
		return ctx, nil
	},
}

// Auth implements Interceptor.
type Auth struct {
	AuthFunc   func(ctx context.Context) (context.Context, error)
	AuthScheme string // bearer by default
	Token      string
}

// UnaryIntercept implements UnaryIntercept function of Interceptor.
func (t Auth) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		c, e := t.AuthFunc(ctx)
		if e != nil {
			return nil, e
		}
		return handler(c, req)
	}
}

// StreamIntercept implements StreamIntercept function of Interceptor.
func (t Auth) StreamServerIntercept() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		c, e := t.AuthFunc(stream.Context())
		if e != nil {
			return e
		}
		wrappedStream := WrapServerStream(stream)
		wrappedStream.SetContext(c)
		return handler(srv, wrappedStream)
	}
}

func (t Auth) Label() string {
	return "yell_auth"
}
