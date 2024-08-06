//go:build !windows && !darwin
// +build !windows,!darwin

package pidfile

import (
	"os"
	"path/filepath"
	"strconv"
)

func pidIsExist(pid int) bool {
	_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(pid)))
	return err == nil
}
