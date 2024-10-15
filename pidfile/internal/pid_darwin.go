//go:build darwin
// +build darwin

package internal

import (
	"syscall"
)

func isProcessExist(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
