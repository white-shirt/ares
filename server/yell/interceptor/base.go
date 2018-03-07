package interceptor

import (
	"github.com/sevenNt/ares/server/yell"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Base ...
type Base struct{}

// UnaryIntercept implements UnaryIntercept function of Interceptor.
func (b Base) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return handler(ctx, req)
	}
}

// StreamIntercept implements StreamIntercept function of Interceptor.
func (b Base) StreamServerIntercept() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		return handler(srv, stream)
	}
}

func (b Base) HookBeforeServe(*yell.Server) {}

func (b Base) Label() string { return "unknown" }

// ServerStream wraps grpc stream server.
type ServerStream struct {
	grpc.ServerStream
	context context.Context
}

// Context returns server stream context.
func (ss *ServerStream) Context() context.Context {
	return ss.context
}

// SetContext sets server stream context.
func (ss *ServerStream) SetContext(ctx context.Context) {
	ss.context = ctx
}

// WrapServerStream wraps server stream with provided stream.
func WrapServerStream(stream grpc.ServerStream) *ServerStream {
	if existing, ok := stream.(*ServerStream); ok {
		return existing
	}

	return &ServerStream{ServerStream: stream, context: stream.Context()}
}
