//go:build !windows

package main

import "os/exec"

// applyHiddenWindow is a no-op on non-Windows hosts.
func applyHiddenWindow(cmd *exec.Cmd) {}
