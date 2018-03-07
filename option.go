package ares

import (
	"fmt"
	"log"

	"github.com/sevenNt/ares/application"
)

var defaultOptions = Options{
	version: "v1",
	systemd: false,
	debug:   false,
}

// Option is used to set options for the application.
type Option func(*Options)

// Options wraps application options.
type Options struct {
	version string
	systemd bool
	mode    string
	debug   bool
}

// Label return app label
func (opts Options) Label() string {
	anyEmpty := func(opt ...string) (empty bool) {
		for _, o := range opt {
			if o == "" {
				return true
			}
		}
		return false
	}

	if anyEmpty(application.Name(), opts.mode) {
		log.Panicf("empty ares app options: %v", opts)
	}

	if opts.version != "" {
		return fmt.Sprintf("%s:%s:%s", application.Name(), opts.version, opts.mode)
	} else {
		return fmt.Sprintf("%s:%s", application.Name(), opts.mode)
	}
}

// WithVersion injects version.
func WithVersion(ver string) Option {
	return func(o *Options) {
		if ver != "" {
			o.version = ver
		}
	}
}

// WithSystemd injects systemd enable flag.
func WithSystemd(systemd bool) Option {
	return func(o *Options) {
		o.systemd = systemd
	}
}

// WithDebug injects application debug trigger
func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.debug = debug
	}
}

// WithMode injects application mode
func WithMode(mode string) Option {
	return func(o *Options) {
		o.mode = mode
	}
}
