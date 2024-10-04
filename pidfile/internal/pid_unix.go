//go:build !windows && !darwin
// +build !windows,!darwin

package internal

import (
	"os"
	"path/filepath"
	"strconv"
)

func pidExist(pid int) bool {
	_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(pid)))
	return err == nil
}
