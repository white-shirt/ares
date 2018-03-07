package yell_test

import (
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc/examples/helloworld/helloworld"

	"github.com/sevenNt/ares/server/yell"
	. "github.com/smartystreets/goconvey/convey"
)

type Greeter struct {
	yell.Handler
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello, " + in.Name}, nil
}

func TestRegister(t *testing.T) {
	Convey("注册路由", t, func() {
		server := yell.New()
		server.Register(helloworld.RegisterGreeterServer, new(Greeter))
		server.Start()
		So(len(server.GetServiceInfo()), ShouldEqual, 1)
	})
}
