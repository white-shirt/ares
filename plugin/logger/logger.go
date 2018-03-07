package logger

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sevenNt/wzap"
	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/server/echo"
	"google.golang.org/grpc"
)

const (
	defaultAccessLogName = "access.json"
)

// Logger logger plugin.
type Logger struct {
	Options
	logger *wzap.Logger
}

// Default gets default logger interceptor.
func Default() *Logger {
	return New(
		Directory(application.EnvAccessLogDir()),
	)
}

// NewAccessLog returns new access log instance.
func NewAccessLog(opts ...Option) *Logger {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	var logOpts = make([]wzap.Option, 0)
	path := options.path
	if path == "" {
		if options.directory != "" {
			if strings.HasSuffix(options.directory, "/") {
				options.directory = options.directory[:len(options.directory)-1]
			}
			path = fmt.Sprintf("%s/%s", options.directory, defaultAccessLogName)
		} else {
			path = defaultAccessLogName
		}
	}

	logOpts = append(logOpts, wzap.WithPath(path))
	logOpts = append(logOpts, wzap.WithLevelString("INFO"))
	if options.consoleOutput {
		logOpts = append(logOpts, wzap.WithOutput(
			wzap.WithLevelString("INFO"),
			wzap.WithColorful(true),
			wzap.WithAsync(false),
			wzap.WithPrefix("APP]>"),
		))
	}

	return &Logger{
		logger: wzap.New(logOpts...),
	}
}

// New returns new logger instance
func New(opts ...Option) *Logger {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return &Logger{
		Options: options,
		logger: wzap.New(
			wzap.WithPath(options.path),
			wzap.WithLevelString("INFO"),
		),
	}
}

// Func implements HTTP Middleware interface.
func (r *Logger) Func() echo.MiddlewareFunc {
	pool := sync.Pool{
		New: func() interface{} {
			return &HTTPAccessMessage{}
		},
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			beg := time.Now().Round(time.Microsecond)
			err = next(c)
			req := c.Request()
			rep := c.Response()

			msg := pool.Get().(*HTTPAccessMessage)
			msg.Beg = beg.Unix()
			msg.Cost = float64(time.Since(beg).Round(time.Microsecond)) / float64(time.Millisecond)
			msg.Method = req.Method
			msg.Client = req.RemoteAddr
			msg.Status = rep.Status()
			msg.Path = req.URL.Path()
			r.logger.Info("http", "access", msg)
			pool.Put(msg)
			return
		}
	}
}

// UnaryServerIntercept implements gRPC unary server interceptor interface
func (r *Logger) UnaryServerIntercept() grpc.UnaryServerInterceptor {
	pool := sync.Pool{
		New: func() interface{} {
			return &GRPCAccessMessage{}
		},
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		beg := time.Now()
		resp, err = handler(ctx, req)

		msg := pool.Get().(*GRPCAccessMessage)
		msg.Beg = beg.Unix()
		msg.Cost = float64(time.Since(beg).Round(time.Microsecond)) / float64(time.Millisecond)
		msg.Method = info.FullMethod
		msg.Error = err
		r.logger.Info("unary", "access", msg)
		pool.Put(msg)
		return
	}
}

// StreamServerIntercept implements gRPC stream server interceptor interface
func (r *Logger) StreamServerIntercept() grpc.StreamServerInterceptor {
	pool := sync.Pool{
		New: func() interface{} {
			return &GRPCAccessMessage{}
		},
	}
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		beg := time.Now()
		err = handler(srv, ss)

		msg := pool.Get().(*GRPCAccessMessage)
		msg.Beg = beg.Unix()
		msg.Cost = float64(time.Since(beg).Round(time.Microsecond)) / float64(time.Millisecond)
		msg.Method = info.FullMethod
		msg.Error = err
		r.logger.Info("stream", "access", msg)
		pool.Put(msg)
		return
	}
}

// Label implements plugin interface
func (r *Logger) Label() string {
	return "logger"
}
