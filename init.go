package ares

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sevenNt/ares/application"
	"github.com/sevenNt/ares/flag"
	"github.com/sevenNt/hera"
	"github.com/sevenNt/wzap"
)

func init() {
	runtime.GOMAXPROCS(int(runtime.NumCPU()) / 2)

	wzap.SetDefaultFields(
		wzap.String("aid", application.ID()),
		wzap.String("iid", application.UUID()),
	)
	wzap.SetDefaultDir(application.LogDir())
	wzap.SetDefaultLogger(wzap.New(
		wzap.WithOutputKV(map[string]interface{}{
			"level": "Info",
			"path":  "default.json",
		}),
	))

	flag.Register(
		&flag.StringFlag{
			Name:    "host",
			Usage:   "specify host",
			Default: "127.0.0.1",
			Action: func(name string, fs *flag.FlagSet) {
				host := fs.String(name)
				if host == "" {
					host = application.EnvServerHost()
				}
				fmt.Printf(">HOST: %s\n", host)
			},
		},
		&flag.StringFlag{
			Name:    "c, config",
			Usage:   "config file path",
			Default: "config/config.toml",
			Action: func(name string, fs *flag.FlagSet) {
				fmt.Println("config: ", fs.String(name))
				hera.MustLoadFromFile(fs.String(name), false)
			},
		},
		&flag.BoolFlag{
			Name:    "v, version",
			Usage:   "print app version",
			Default: false,
			Action: func(name string, fs *flag.FlagSet) {
				fmt.Println(application.BuildFlags())
				os.Exit(0)
			},
		},
	)
}
