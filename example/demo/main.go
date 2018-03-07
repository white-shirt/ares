package main

import (
	"github.com/sevenNt/ares"
	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/example/demo/app/grpc"
	"github.com/sevenNt/ares/example/demo/app/http"
	"github.com/sevenNt/ares/server"
	"github.com/sevenNt/wzap"
)

func main() {
	wzap.Info("startup app")

	app := ares.NewAPP()
	//app.AddWorker("tick", worker.NewTickWorker(time.Second*10))

	wzap.WInfo("db", "hello", "uuid", application.UUID())
	wzap.WInfo("none", "bye")
	wzap.WWarn("default", "bye")
	wzap.WInfo("grpc", "grpc", "grpc", "grpc")

	app.Serve(
		new(grpc.ExampleGRPCServer),
		server.Port(18093),
		server.Alias("example-http-alias"),
		server.Name("example-grpc"),
	)

	app.Serve(
		new(http.ExampleHTTPServer),
		server.Port(18098),
		server.Name("example-http"),
	)

	app.Run()
}
