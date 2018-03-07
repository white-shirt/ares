package registry

import (
	"fmt"

	"github.com/sevenNt/ares/application"
)

var (
	defaultRegistry Registry
)

func InitRegistry(r Registry) {
	defaultRegistry = r
}

func RegisterApp(info AppInfo) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	return defaultRegistry.RegisterApp(info)
}

func UnregisterApp() error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	return defaultRegistry.UnregisterApp()
}

func RegisterResource(uri string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	key := fmt.Sprintf("deps:%s:%s", application.Label(), application.UUID())
	return defaultRegistry.Register(key, uri)
}

func UnregisterResource(uri string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	key := fmt.Sprintf("deps:%s:%s", application.Label(), application.UUID())
	return defaultRegistry.Unregister(key, uri)
}

func RegisterService(scheme string, addr string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	key := fmt.Sprintf("%s:%s", scheme, application.Label())
	return defaultRegistry.Register(key, addr)
}

func UnregisterService(scheme string, addr string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	key := fmt.Sprintf("%s:%s", scheme, application.Label())
	return defaultRegistry.Unregister(key, addr)
}

func Register(key string, addr string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	return defaultRegistry.Register(key, addr)
}

func Unregister(key string, addr string) error {
	if defaultRegistry == nil {
		return ErrUninitialRegistry
	}
	return defaultRegistry.Unregister(key, addr)
}

func UnregisterAll() {
	if defaultRegistry != nil {
		defaultRegistry.UnregisterAll()
	}
}
