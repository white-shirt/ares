package yell

import (
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/sevenNt/ares/server"
	"google.golang.org/grpc"
)

// Server wraps grpc server.
type Server struct {
	*grpc.Server
	server.Options
	addr   string
	name   string
	alias  []string
	opts   []grpc.ServerOption
	inters []ServerInterceptor

	listener net.Listener

	Func func(*Server)

	registers          map[reflect.Value]interface{}
	hookersBeforeServe []func(*Server)

	running chan struct{}
}

// NewServer constructs a grpc server.
func NewServer(lis net.Listener, opts ...server.Option) *Server {
	s := &Server{
		inters:    make([]ServerInterceptor, 0),
		registers: make(map[reflect.Value]interface{}),
		listener:  lis,
		running:   make(chan struct{}, 1),
	}

	return s
}

// IsRunning checks which server is running.
func (s *Server) IsRunning() bool {
	return len(s.running) > 0
}

func (s *Server) register() {
	for register, receiver := range s.registers {
		rv := register
		rt := rv.Type()
		if rt.Kind() != reflect.Func {
			panic("register must be func")
		}

		cv := reflect.ValueOf(receiver)
		ct := cv.Type()
		if ct.Kind() == reflect.Ptr {
			ct = ct.Elem()
		}
		if ct.Kind() != reflect.Struct {
			panic("receiver must be struct")
		}

		params := make([]reflect.Value, 2)
		params[0] = reflect.ValueOf(s.Server)
		params[1] = cv

		rv.Call(params)
	}
}

// Register register
func (s *Server) Register(register interface{}, receiver interface{}) {
	s.registers[reflect.ValueOf(register)] = receiver
}

// DumpServiceInfo dumps service information.
func (s *Server) DumpServiceInfo() {
	for fm, info := range s.GetServiceInfo() {
		for _, method := range info.Methods {
			log.Printf("[YELL] \x1b[33m%8s\x1b[0m %s", fm, method.Name)
		}
	}
	log.Printf("[YELL] \x1b[33m%8s\x1b[0m %s", "Listen On", s.Addr())
}

// HookBeforeServe injects hooks executed before server run.
func (s *Server) HookBeforeServe(f func(*Server)) {
	s.hookersBeforeServe = append(s.hookersBeforeServe, f)
}

// WithGRPCServerOption adds grpc server options.
func (s *Server) WithGRPCServerOption(opts ...grpc.ServerOption) *Server {
	s.opts = append(s.opts, opts...)
	return s
}

// WithUnaryInterceptors adds unary interceptors.
func (s *Server) WithUnaryInterceptors(intes ...ServerInterceptor) *Server {
	var interceptors = make([]grpc.UnaryServerInterceptor, 0)
	for _, inte := range intes {
		interceptors = append(interceptors, inte.UnaryServerIntercept())
	}
	opt := grpc.UnaryInterceptor(UnaryInterceptorChain(interceptors...))
	s.opts = append(s.opts, opt)
	s.inters = append(s.inters, intes...)
	return s
}

// WithStreamInterceptors adds stream interceptors.
func (s *Server) WithStreamInterceptors(intes ...ServerInterceptor) *Server {
	var interceptors = make([]grpc.StreamServerInterceptor, 0)
	for _, inte := range intes {
		interceptors = append(interceptors, inte.StreamServerIntercept())
	}
	s.opts = append(s.opts, grpc.StreamInterceptor(StreamInterceptorChain(interceptors...)))
	s.inters = append(s.inters, intes...)
	return s
}

// Addr returns stream server address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
	//return s.addr
}

// Name returns stream server name.
func (s *Server) Name() string {
	return s.name
}

// Alias returns server aliasname
func (s *Server) Alias() []string {
	return s.alias
}

// Scheme returns server's scheme.
func (s *Server) Scheme() string {
	return "grpc"
}

// Serve returns stream server name.
func (s *Server) Serve() error {
	s.Server = grpc.NewServer(s.opts...)
	//s.Func(s)
	s.register()

	s.DumpServiceInfo()
	for _, hooker := range s.hookersBeforeServe {
		hooker(s)
	}

	s.running <- struct{}{}
	defer func() {
		<-s.running
	}()

	return s.Server.Serve(s.listener)
}

// Stop stops server.
func (s *Server) Stop() {
	log.Printf("grpc server %s-%s is stopping...", s.name, s.addr)
	s.Server.Stop()
}

// CloneHTTP11 ... TODO
func (s *Server) CloneHTTP11() {
	for fm, info := range s.GetServiceInfo() {
		for _, method := range info.Methods {
			fmt.Printf("[YELL] \x1b[33m%8s\x1b[0m %s\n", fm, method.Name)
		}
	}
}
