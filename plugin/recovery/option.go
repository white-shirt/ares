package recovery

type Option func(*Options)
type Options struct {
	stacksize         int
	disableStackAll   bool
	disablePrintStack bool
}

func StackSize(size int) Option {
	return func(opts *Options) {
		opts.stacksize = size
	}
}

func DisablePrintStack(disable bool) Option {
	return func(opts *Options) {
		opts.disablePrintStack = disable
	}
}

func DisableStackAll(disable bool) Option {
	return func(opts *Options) {
		opts.disableStackAll = disable
	}
}
