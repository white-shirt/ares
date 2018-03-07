package application

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
)

var vcsInfo string

var buildTime string

var name string

var id string

// label application label.
var label = fmt.Sprintf("%s", name)

// uuid application unique instance id.
//var uuid = gouuid.Must(gouuid.NewV4()).String()
var uuid = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", Hostname(), id))))

// Name gets application name.
func Name() string {
	if name == "" {
		name = filepath.Base(os.Args[0])
	}
	return name
}

// ID gets application id.
func ID() string {
	return id
}

// Label gets label.
func Label() string {
	return label
}

// BuildTime gets building time.
func BuildTime() string {
	return buildTime
}

// VcsInfo gets vcs revision.
func VcsInfo() string {
	return vcsInfo
}

// Hostname gets hostname.
func Hostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hn
}

// BuildFlags print build flags.
func BuildFlags() string {
	return "\n" +
		fmt.Sprintf("[APP] \x1b[34m%12s\x1b[0m %s\n", "name", Name()) +
		fmt.Sprintf("[APP] \x1b[34m%12s\x1b[0m %s\n", "id", ID()) +
		fmt.Sprintf("[APP] \x1b[34m%12s\x1b[0m %s\n", "vcsInfo", vcsInfo) +
		fmt.Sprintf("[APP] \x1b[34m%12s\x1b[0m %s\n", "buildTime", buildTime)
}

// UUID gets unique instance id.
func UUID() string {
	return uuid
}

// SetLabel sets label.
func SetLabel(ver, mode string) {
	if ver == "" {
		label = fmt.Sprintf("%s:%s", Name(), mode)
	} else {
		label = fmt.Sprintf("%s:%s:%s", Name(), ver, mode)
	}
}

// LogDir gets application log directory.
func LogDir() string {
	logDir := os.Getenv("ARES_BASE_LOG_DIR")
	if logDir == "" {
		logDir = "/home/www/logs/applogs"
	}
	return fmt.Sprintf("%s/%s/%s/", logDir, Name(), UUID())
}
