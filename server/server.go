package server

// Server is an interface for http or grpc services.
type Server interface {
	Serve() error
	Stop()
	GracefulStop()

	IsRunning() bool

	Scheme() string
	Name() string
	Addr() string
	Alias() []string

	// 多路复用
	// Init()
	Mux()
}
