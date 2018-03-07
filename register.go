package ares

import (
	"os"
	"time"

	"log"

	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/registry"
)

func (app *App) register() {
	info := registry.AppInfo{
		UpAt:      time.Now().Format("2006-01-02 15:04:05"),
		Hostname:  application.Hostname(),
		AppID:     application.ID(),
		AppName:   app.options.Label(),
		BuildTime: application.BuildTime(),
		VcsInfo:   application.VcsInfo(),
		PID:       os.Getpid(),
		UUID:      application.UUID(),
		Services:  make(map[string]string),
	}

	for _, srv := range app.servers {
		info.Services[srv.Scheme()] = srv.Addr()
	}

	if err := registry.RegisterApp(info); err != nil {
		log.Printf("[APP] \x1b[33m%8s\x1b[0m %s %s", "REGISTER", "app failed", err.Error())
	}
}

func (app *App) unregister() {
	registry.UnregisterAll()
}
