package worker

import "time"

type TickWorker struct {
	interval time.Duration
}

func (worker TickWorker) Work() error {
	return nil
}

func (worker TickWorker) Stop() {

}

func (worker TickWorker) GracefulStop() {
}
