package yell

import "google.golang.org/grpc"

// UnaryInterceptor wraps groc unary interceptor.
type UnaryInterceptor interface {
	UnaryIntercept() grpc.UnaryServerInterceptor
}

// Streamceptor wraps groc stream interceptor.
type StreamInterceptor interface {
	StreamIntercept() grpc.StreamServerInterceptor
}

// Interceptor wraps groc interceptor.
type Interceptor interface {
	UnaryIntercept() grpc.UnaryServerInterceptor
	StreamIntercept() grpc.StreamServerInterceptor
	Label() string
	HookBeforeServe(*Server)
}

// Interceptor wraps groc interceptor.
type ServerInterceptor interface {
	UnaryServerIntercept() grpc.UnaryServerInterceptor
	StreamServerIntercept() grpc.StreamServerInterceptor
}

type ClientInterceptor interface {
	UnaryClientIntercept() grpc.UnaryClientInterceptor
	StreamClientIntercept() grpc.StreamClientInterceptor
}
