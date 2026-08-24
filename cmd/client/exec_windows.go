//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// applyHiddenWindow ensures the spawned child process (cmd.exe /
// powershell.exe) runs detached with no visible console window so the
// daemon stays 100% background and silent on the target host.
func applyHiddenWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
