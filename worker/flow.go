package worker

import (
	"sync"
)

type WorkerFlowNode interface {
	Worker

	Out() chan interface{}
	SetIn(chan interface{})
}

type WorkerFlow struct {
	workers []Worker
	last    WorkerFlowNode
}

func NewWorkerFlow() *WorkerFlow {
	return &WorkerFlow{
		workers: make([]Worker, 0),
	}
}

func (flow *WorkerFlow) Add(wrk WorkerFlowNode) *WorkerFlow {
	flow.workers = append(flow.workers, wrk)
	if flow.last != nil {
		wrk.SetIn(flow.last.Out())
	}
	flow.last = wrk

	return flow
}

func (flow *WorkerFlow) Stop() {
	for _, wrk := range flow.workers {
		wrk.Stop()
	}
}

func (flow *WorkerFlow) GracefulStop() {
	for _, wrk := range flow.workers {
		wrk.GracefulStop()
	}
}

func (flow *WorkerFlow) Run() {
	var wg sync.WaitGroup
	for _, wrk := range flow.workers {
		wg.Add(1)
		defer wg.Done()
		wrk.Run()
	}

	wg.Wait()
}
