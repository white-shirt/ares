package worker

import (
	"runtime"
	"time"

	"github.com/sevenNt/wzap"
)

type Singleton struct {
	Base
	stop     chan struct{}
	process  chan struct{}
	interval time.Duration
	Do       func() error
}

// NewSingleton 返回一个单例模式的worker， 该worker只会有一个实例执行，即使单次执行超过执行间隔
// NOTE (lvchao) 单次执行时间超过interval定义的时间，将导致下次执行延后
func NewSingleton() *Singleton {
	return &Singleton{
		stop:    make(chan struct{}),
		process: make(chan struct{}),
	}
}

func (singleton Singleton) Run() {
	go singleton.run()

	ticker := time.NewTicker(singleton.interval)
	for {
		select {
		case <-ticker.C:
			go singleton.run()
		case <-singleton.stop:
			return
		}
	}
}

func (singleton Singleton) Stop() {
	if !singleton.isRunning() {
		return
	}

	singleton.stop <- struct{}{}
}

func (singleton Singleton) GracefulStop() {
	if !singleton.isRunning() {
		return
	}

	singleton.process <- struct{}{}
	singleton.stop <- struct{}{}
}

func (singleton Singleton) run() {
	if singleton.isRunning() {
		return
	}

	singleton.lock()
	defer singleton.unlock()
	singleton.work()
}

func (singleton Singleton) lock() {
	singleton.process <- struct{}{}
}
func (singleton Singleton) unlock() {
	<-singleton.process
}

func (singleton Singleton) isRunning() bool {
	return len(singleton.process) > 0
}

func (singleton Singleton) work() {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 1024)
			length := runtime.Stack(stack, true)
			wzap.Warn("[singleton]", "err", string(stack[:length]))
		}
	}()

	singleton.Do()
}
