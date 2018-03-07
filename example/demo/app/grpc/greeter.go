package grpc

import (
	"strings"

	"github.com/sevenNt/ares/server/echo/client"
	"github.com/sevenNt/ares/server/yell"
	"github.com/sevenNt/wzap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
)

// Greeter grpc handler
type Greeter struct {
	yell.Handler
}

var (
	c = client.SetDebug(true).SetHostURL("http://127.0.0.1:18090")
)

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, vals := range md {
			wzap.Infof("[Greeter] SayHello's metadata: %s:[%s]", key, strings.Join(vals, ","))
		}
	}

	reply := &helloworld.HelloReply{Message: "hello"}
	//resp, err := c.R().Get("/ping")
	//if err != nil {
	//reply.Message = err.Error()
	//return reply, nil
	//}

	//if err := resp.UnmarshalJSON(reply); err != nil {
	//reply.Message = err.Error()
	//}

	// 在做grpc proxy的时候，无法通过下面的时候设置response header
	// 因为无法取得stream transport
	//grpc.SetHeader(ctx, metadata.Pairs("X-RESP-ID", "123"))

	return reply, nil
}
