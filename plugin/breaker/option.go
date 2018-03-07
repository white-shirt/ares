package breaker

type Options struct {
	//backoff       backoff.BackOff
	//clock         clock.Clock
	//shouldTrip    TripFunc
	//windowTime    time.Duration
	//windowBuckets int
	label string
}

type Option func(*Options)

//func BackOff(backoff backoff.BackOff) *Option {
//return func(opts *Options) {
//opts.backoff = backoff
//}
//}

//func ShouldTrip(shouldTrip TripFunc) *Option {
//return func(opts *Options) {
//opts.shouldTrip = shouldTrip
//}
//}

//func Window(dura time.Duration, buckets int) *Option {
//return func(opts *Options) {
//opts.windowTime = dura
//opts.windowBuckets = buckets
//}
//}
func Label(label string) Option {
	return func(opts *Options) {
		opts.label = label
	}
}
