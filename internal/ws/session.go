package ws

import (
	"context"
	"sync"
	"time"

	"log/slog"

	"github.com/coder/websocket"
	"github.com/its-ernest/RascalRAT/internal/protocol"
)

type AgentSession struct {
	ID          string
	Conn        *websocket.Conn
	Connected   time.Time
	Send        chan protocol.TaskRequest
	ActiveTasks map[string]chan protocol.TaskResponse
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	closeOnce   sync.Once
}

func NewAgentSession(id string, conn *websocket.Conn) *AgentSession {
	ctx, cancel := context.WithCancel(context.Background())

	return &AgentSession{
		ID:          id,
		Conn:        conn,
		Connected:   time.Now(),
		Send:        make(chan protocol.TaskRequest, 64),
		ActiveTasks: make(map[string]chan protocol.TaskResponse),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *AgentSession) RegisterTaskChan(taskID string, ch chan protocol.TaskResponse) {
	s.mu.Lock()
	s.ActiveTasks[taskID] = ch
	s.mu.Unlock()
}

func (s *AgentSession) DeregisterTaskChan(taskID string) {
	s.mu.Lock()
	delete(s.ActiveTasks, taskID)
	s.mu.Unlock()
}

func (s *AgentSession) RouteResponse(resp protocol.TaskResponse) {
	s.mu.RLock()
	ch, ok := s.ActiveTasks[resp.TaskID]
	s.mu.RUnlock()

	if !ok {
		slog.Warn("received unmatched task response", "node_id", s.ID, "task_id", resp.TaskID)
		return
	}

	select {
	case ch <- resp:
	default:
		slog.Warn("dropping task response because handler is not ready", "node_id", s.ID, "task_id", resp.TaskID)
	}
}

func (s *AgentSession) Start(hub *Hub) {
	go s.readPump(hub)
	go s.writePump()
}

func (s *AgentSession) Close(reason string) {
	s.closeOnce.Do(func() {
		s.cancel()
		s.failPendingTasks(reason)
		close(s.Send)
		if err := s.Conn.Close(websocket.StatusNormalClosure, reason); err != nil {
			slog.Debug("websocket close returned error", "node_id", s.ID, "err", err)
		}
	})
}

func (s *AgentSession) failPendingTasks(reason string) {
	s.mu.Lock()
	pending := s.ActiveTasks
	s.ActiveTasks = make(map[string]chan protocol.TaskResponse)
	s.mu.Unlock()

	for taskID, ch := range pending {
		select {
		case ch <- protocol.TaskResponse{TaskID: taskID, Success: false, ErrorMessage: reason}:
		default:
		}
	}
}
