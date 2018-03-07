// +build !windows
package ares

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

const (
	listenFdsStart = 3
)

// FDsFromSystemd return registered fd by systemd.
func FDsFromSystemd(unsetEnv bool) []*os.File {
	if unsetEnv {
		defer func() {
			if err := os.Unsetenv("LISTEN_PID"); err != nil {
				fmt.Println("os.unsetEnv LISTEN_PID failed, ", err)
			}
			if err := os.Unsetenv("LISTEN_FDS"); err != nil {
				fmt.Println("os.unsetEnv LISTEN_FDS failed, ", err)
			}
		}()
	}

	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return nil
	}

	nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil || nfds == 0 {
		return nil
	}

	files := make([]*os.File, 0, nfds)
	for fd := listenFdsStart; fd < listenFdsStart+nfds; fd++ {
		syscall.CloseOnExec(fd)
		files = append(files, os.NewFile(uintptr(fd), "LISTEN_FD_"+strconv.Itoa(fd)))
	}

	return files
}
