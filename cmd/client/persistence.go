//go:build windows

package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const runRegistryKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// isAlreadyPersistent interrogates the HKCU Run key to see if the value matches expectations
func isAlreadyPersistent(name, expectedPath string) bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	currentValue, _, err := key.GetStringValue(name)
	if err != nil {
		return false
	}

	// Strict comparison prevents unnecessary mutations if the values match exactly
	return filepath.Clean(currentValue) == filepath.Clean(expectedPath)
}

// establishPersistence copies the file to the appdata directory and links it to startup
func establishPersistence(appName, targetExe string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return err
	}

	// Create target hosting container directory structure if missing
	targetDir := filepath.Dir(targetExe)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Only duplicate the file on the storage drive if we aren't already running from it
	if filepath.Clean(currentExe) != filepath.Clean(targetExe) {
		slog.Info("Copying binary to user profile space", "destination", targetExe)
		if err := copyFile(currentExe, targetExe); err != nil {
			return err
		}
	}

	// Commit registry configuration value changes
	slog.Info("Updating environment registry state keys", "key_path", runRegistryKeyPath)
	return writeToRegistry(appName, targetExe)
}

// writeToRegistry updates the context execution path target inside HKCU
func writeToRegistry(name, path string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(name, path)
}

// copyFile handles fundamental atomic filesystem duplication operations
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
