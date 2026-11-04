//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// GUI exe에서 ffmpeg를 띄울 때 콘솔 창이 깜빡이지 않게 한다.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
