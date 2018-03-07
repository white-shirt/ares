package yell

// Option is used to set options for the server.
type Option func(*Options)

// Options wraps server options.
type Options struct {
	name string
	addr string
}

// WithName adds name to options.
func WithName(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// WithAddr adds addr to options.
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}
