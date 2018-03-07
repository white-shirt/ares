package worker

// Worker is an interface for application worker.
type Worker interface {
	Run()
	Stop()
	GracefulStop()
}

// Option is used to set options for the application work.
type Option func(*Options)

// Options wraps application service work.
type Options struct {
	label string
}

// Label sets work label.
func Label(label string) Option {
	return func(o *Options) {
		o.label = label
	}
}
