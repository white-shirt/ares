package worker_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_WorkerFlow(t *testing.T) {
	Convey("建立workerflow", t, func() {
		So("1", ShouldEqual, "1")
	})
}
