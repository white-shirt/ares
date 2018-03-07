package echo

// Option is used to set options for the server.
type Option func(*Options)

// Options wraps server options.
type Options struct {
	name string
	addr string
}

// WithName injects name.
func WithName(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// WithAddr injects address.
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}
