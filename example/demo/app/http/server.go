package http

import (
	"time"

	"github.com/sevenNt/ares/example/demo/app/grpc"
	"github.com/sevenNt/ares/plugin/logger"
	"github.com/sevenNt/ares/plugin/metric"
	"github.com/sevenNt/ares/plugin/ratelimit"
	"github.com/sevenNt/ares/plugin/recovery"
	"github.com/sevenNt/ares/server/echo"
)

// ExampleHTTPServer a example http server
type ExampleHTTPServer struct {
	*echo.Server
}

// Mux example http server router
func (s *ExampleHTTPServer) Mux() {
	s.Use(
		recovery.Default(),
		logger.NewAccessLog(
			logger.ConsoleOutput(true), // console打印影响性能
		),
		ratelimit.New(
			ratelimit.FillInterval(time.Second),
			ratelimit.Capacity(1),
			ratelimit.Label("http-example/ping-ratelimit"),
		),
		metric.New(
			metric.Prefix("1909"),
			metric.Interval(time.Second*10),
			metric.PrometheusPusherWithAddr("10.1.41.51:9091"),
		),
	)
	s.GET("/ping", pong)

	greeter := new(grpc.Greeter)
	s.GRPCProxyGet("/greet/hello", greeter.SayHello)

	s.HookBeforeServe(hooker)
}

func pong(c *echo.Context) error {
	return c.JSON(200, map[string]string{
		"message": "pong",
	})
}

func hooker(s *echo.Server) {
	s.DumpRouteInfo()
}
