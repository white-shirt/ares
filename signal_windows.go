// +build windows
package ares

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func (app *App) hookSignals() {
	signal.Notify(
		app.sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		var sig os.Signal
		for {
			sig = <-app.sigChan
			log.Warnf("[ares] Receive Signal %v", sig)
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP:
				app.shutdown()
			}
		}
	}()
}
