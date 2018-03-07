// build windows
package ares

import (
	"os"
)

// FDsFromSystemd return registered fd by systemd.
func FDsFromSystemd(unsetEnv bool) []*os.File {
	files := make([]*os.File, 0)
	return files
}
