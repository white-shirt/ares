package ratelimit

import "time"

type Option func(*Options)
type Options struct {
	rate            float64
	interval        time.Duration
	label           string
	capacity        int64
	waitMaxDuration time.Duration
	trigger         bool
}

func Rate(rate float64) Option {
	return func(opts *Options) {
		opts.rate = rate
	}
}

func Label(label string) Option {
	return func(opts *Options) {
		opts.label = label
	}
}

func FillInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.interval = interval
	}
}

func Capacity(cap int64) Option {
	return func(opts *Options) {
		opts.capacity = cap
	}
}

func WaitMaxDuration(max time.Duration) Option {
	return func(opts *Options) {
		opts.waitMaxDuration = max
	}
}

func Trigger(trig bool) Option {
	return func(opts *Options) {
		opts.trigger = trig
	}
}
