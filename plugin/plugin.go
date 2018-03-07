package plugin

import (
	"github.com/sevenNt/ares/server/echo"
	"google.golang.org/grpc"
)

type Plugin interface {
	Func() echo.MiddlewareFunc
	UnaryServerIntercept() grpc.UnaryServerInterceptor
	UnaryClientIntercept() grpc.UnaryClientInterceptor
	StreamServerIntercept() grpc.StreamServerInterceptor
	StreamClientIntercept() grpc.StreamClientInterceptor
	Label() string
}

//type Deprecated struct{}
//func (Deprecated) HookRoute(method string, path string) {}
//func (Deprecated) Clone() (echo.Middleware, bool)       {}
