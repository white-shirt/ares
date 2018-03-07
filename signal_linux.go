// +build !windows

package ares

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sevenNt/wzap"
)

func (app *App) hookSignals() {
	signal.Notify(
		app.sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGSTOP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGKILL,
	)

	go func() {
		var sig os.Signal
		for {
			sig = <-app.sigChan
			wzap.Warnf("[ares] Receive Signal %v", sig)
			time.Sleep(time.Second)
			switch sig {
			case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGUSR1:
				app.shutdown() // graceful stop
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGUSR2, syscall.SIGTERM:
				app.terminate() // terminalte now
			}
		}
	}()
}
