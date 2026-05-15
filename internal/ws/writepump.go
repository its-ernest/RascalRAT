package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (s *AgentSession) writePump() {
	for request := range s.Send {
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		if err := wsjson.Write(ctx, s.Conn, request); err != nil {
			status := websocket.CloseStatus(err)
			if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
				slog.Info("agent websocket writer exiting", "node_id", s.ID, "close_status", status)
			} else {
				slog.Error("failed to write task request to agent", "node_id", s.ID, "err", err, "close_status", status)
			}
			cancel()
			return
		}
		cancel()
	}
}
