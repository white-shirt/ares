package metric

import (
	"time"
)

// Option metric option func
type Option func(*Options)

// Options metric options
type Options struct {
	pusher   Pusher
	prefix   string
	name     string
	interval time.Duration
}

// PrometheusPusherWithAddr create PrometheusPusher
func PrometheusPusherWithAddr(addr string) Option {
	pusher := &PrometheusPusher{
		address: addr,
		metrics: make(chan *EchoCollector, MaxMetricNum),
	}
	return func(opts *Options) {
		opts.pusher = pusher
	}
}

// Interval set metric  interval
func Interval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.interval = interval
	}
}

// Prefix set metric prefix
func Prefix(prefix string) Option {
	return func(opts *Options) {
		opts.prefix = prefix
	}
}
