package main

import (
	"fmt"
	"time"

	"github.com/sevenNt/ares/client/chiwoo"
	"github.com/sevenNt/ares/plugin/breaker"
	"github.com/sevenNt/ares/plugin/tracer"
	"github.com/sevenNt/ares/server/yell"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/status"
)

const (
	address = "10.18.3.53:18091"
	//address     = "127.0.0.1:18091"
	defaultName = "world"
	AppID       = "190"
)

var client *chiwoo.Chiwoo

func init() {
	// 走服务发现
	client = chiwoo.New("grpc:main:1.0.0:lvchao",
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			chiwoo.UnaryClientInterceptorChain(
				tracer.New(),
				breaker.New(
					breaker.Label("chiyou"),
				),
			),
		),
	)

	client.SetDialTimeout(time.Second * 2)
}

func main() {
	if err := client.Do(func(conn *grpc.ClientConn) error {
		c := pb.NewGreeterClient(conn)
		r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: "hello"})
		if err != nil {
			return err
		}
		fmt.Printf("Greeting: %s", r.Message)
		return nil
	}); err != nil {
		if sts, ok := status.FromError(err); ok {
			switch sts.Code() {
			case codes.OK:
				fmt.Println("what ???")
			case yell.CodeTooManyRequest:
				fmt.Println("?==> ", sts.Code(), sts.Message())
			}
		}
	}
}
