package job

import (
	"fmt"
	"time"

	"github.com/sevenNt/ares/worker"
)

type DemoWorker struct {
	worker.Base
}

func (DemoWorker) Run() {
	for {
		fmt.Println("demo worker run...")
		time.Sleep(time.Second)
	}
}

func (DemoWorker) Stop() {
	fmt.Println("demo job stop")
}

func (DemoWorker) GracefulStop() {
	time.Sleep(time.Second * 10)
	fmt.Println("demo job garaceful stop")
}
