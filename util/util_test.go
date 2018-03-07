package util_test

import (
	"fmt"
	"testing"

	"github.com/sevenNt/ares/util"
)

func TestGetCurrentPath(t *testing.T) {
	path, err := util.GetCurrentPath()
	if err != nil {
		t.Fail()
	}

	fmt.Println("path: ", path)
}

func TestGetCurrentExecutable(t *testing.T) {
	file := util.GetCurrentExecutable()
	fmt.Print("file", file)
}
