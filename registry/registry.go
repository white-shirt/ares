package registry

import (
	"encoding/json"
	"errors"
)

// default registry
var (
	//registry             Registry
	ErrUninitialRegistry = errors.New("uninitial registry")
)

type AppInfo struct {
	UpAt      string            `json:"ut"`            // 实例的启动时间
	Hostname  string            `json:"hostname"`      // 机器hostname
	AppID     string            `json:"aid"`           // 项目应用ID
	AppName   string            `json:"-"`             // 应用名称
	VcsInfo   string            `json:"vcs"`           // version control info
	BuildTime string            `json:"bt"`            // 编译时间信息
	PID       int               `json:"pid"`           // 进程id
	Services  map[string]string `json:"srv,omitempty"` // 服务信息
	UUID      string            `json:"-"`
}

func (info AppInfo) Desc() string {
	raw, _ := json.Marshal(info)
	return string(raw)
}

// Registry provides an interface for service discovery
type Registry interface {
	Register(string, string) error
	Unregister(string, string) error
	UnregisterAll()
	Watch() (Watcher, error)
	String() string

	RegisterApp(AppInfo) error
	UnregisterApp() error
}

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() (*Result, error)
	Stop()
}

// Result is returned by a call to Next on
// the watcher. Actions can be create, update, delete
type Result struct {
	Action  string
	Service *Service
}
