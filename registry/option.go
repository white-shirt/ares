package registry

import (
	"crypto/tls"
	"time"
)

// Option is used to set options for the registry.
type Option func(*Options)

// Options wraps registry options.
type Options struct {
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
	TLSConfig *tls.Config
}

// WithAddrs injects addresses.
func WithAddrs(addrs []string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// WithTimeout injects timeout.
func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// WithSecure injects secure flag.
func WithSecure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// WithTLSConfig injects TLS configuration.
func WithTLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}
