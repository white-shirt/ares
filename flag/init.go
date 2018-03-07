package flag

import (
	"os"
)

var defaultFlags = []Flag{
	// HelpFlag prints usage of application.
	&BoolFlag{
		Name:  "h, help",
		Usage: "show help",
		Action: func(name string, fs *FlagSet) {
			fs.PrintDefaults()
			os.Exit(0)
		},
	},
}
