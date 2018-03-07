package grpc

import (
	"time"

	"github.com/sevenNt/ares/plugin/logger"
	"github.com/sevenNt/ares/plugin/metric"
	"github.com/sevenNt/ares/plugin/ratelimit"
	"github.com/sevenNt/ares/plugin/recovery"
	"github.com/sevenNt/ares/server/yell"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// ExampleGRPCServer http server for example
type ExampleGRPCServer struct {
	*yell.Server
}

// Mux server route
func (s *ExampleGRPCServer) Mux() {
	s.Register(helloworld.RegisterGreeterServer, &Greeter{})
	s.WithUnaryInterceptors(
		recovery.Default(),
		logger.NewAccessLog(
			logger.ConsoleOutput(true), // console打印影响性能
		),
		metric.New(
			metric.Interval(time.Second*10),
			metric.Prefix("1910"),
			metric.PrometheusPusherWithAddr("10.1.41.51:9091"),
		),
		ratelimit.New(
			ratelimit.FillInterval(time.Second*3),
			ratelimit.Capacity(1),
			ratelimit.Label("grpc-example-ratelimit"),
		),
	)
}
