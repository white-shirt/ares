package server

import "fmt"

// Option is used to set options for the application service.
type Option func(*Options)

// Options wraps application service options.
type Options struct {
	addr  string
	host  string
	port  int
	name  string
	alias []string
}

// Addr gets service startup address.
func (opts Options) Addr() string {
	return fmt.Sprintf("%s:%d", opts.host, opts.port)
}

func (opts Options) Name() string {
	return opts.name
}

// Alias gets service alias name in server-naming
func (opts Options) Alias() []string {
	return opts.alias
}

// Host sets service startup host.
func Host(host string) Option {
	return func(o *Options) {
		o.host = host
	}
}

// Port sets service startup port.
func Port(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

// PortStr sets service startup string port.
func PortStr(port string) Option {
	return func(o *Options) {
		o.port = 9090
	}
}

// Addr sets service startup address.
func Addr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}

// Name sets service startup name.
func Name(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// Alias set service startup naming alias
func Alias(alias ...string) Option {
	return func(o *Options) {
		o.alias = alias
	}
}
