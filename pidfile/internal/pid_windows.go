//go:build windows
// +build windows

package internal

import (
	"syscall"
)

func isProcessExist(pid int) bool {
	p, err := syscall.OpenProcess(0x1000, false, uint32(pid))
	if err != nil {
		return false
	}
	var code uint32
	err = syscall.GetExitCodeProcess(p, &code)
	syscall.Close(p)
	if err != nil {
		return code == 259
	}
	return true
}
