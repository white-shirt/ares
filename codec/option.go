package codec

type (
	Options struct {
		indent    bool
		useNumber bool
		format    string
	}

	Option func(*Options)
)

func Indent(indent bool) Option {
	return func(o *Options) {
		o.indent = indent
	}
}

func UseNumber(useNumber bool) Option {
	return func(o *Options) {
		o.useNumber = useNumber
	}
}

func Format(format string) Option {
	return func(o *Options) {
		o.format = format
	}
}
