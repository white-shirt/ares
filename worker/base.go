package worker

type Base struct{}

func (worker Base) Run() {
	panic("must implements")
}

func (worker Base) Stop() {}

func (worker Base) GracefulStop() {}

func (worker Base) SetIn(in chan interface{}) {}

func (worker Base) Out() chan interface{} { return nil }
