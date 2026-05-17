//go:build windows

package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	// Initialize structured logging format
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Resolve current executable path
	exePath, err := os.Executable()
	if err != nil {
		slog.Error("failed to resolve current executable path", slog.Any("error", err))
		os.Exit(1)
	}
	appName := filepath.Base(exePath)

	// 2. Compute the expected persistent destination path
	localAppData, err := os.UserConfigDir()
	if err != nil {
		slog.Error("failed to resolve user configuration directory", slog.Any("error", err))
		os.Exit(1)
	}
	expectedTargetExe := filepath.Join(localAppData, appName, appName)
	if filepath.Ext(expectedTargetExe) != ".exe" {
		expectedTargetExe += ".exe"
	}

	// 3. Check if configuration matches the target system status
	if isAlreadyPersistent(appName, expectedTargetExe) {
		slog.Info("startup persistence verified; skipping configuration steps",
			slog.String("app_name", appName),
			slog.String("registered_path", expectedTargetExe),
		)
		runMainApplication()
		return
	}

	// 4. Fallback: Establish persistence if missing or pointing to the wrong path
	slog.Info("persistence missing or configuration drift detected; installing...",
		slog.String("app_name", appName),
	)

	if err := establishPersistence(appName, expectedTargetExe); err != nil {
		slog.Error("failed to establish startup persistence",
			slog.String("app_name", appName),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	slog.Info("successfully established startup entry", slog.String("app_name", appName))

	runMainApplication()
}
