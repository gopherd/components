//go:build darwin
// +build darwin

package pidfile

import (
	"syscall"
)

func pidIsExist(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
