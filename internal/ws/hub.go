package ws

import (
	"errors"
	"sync"

	"github.com/coder/websocket"
)

type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*AgentSession
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*AgentSession),
	}
}

func (h *Hub) Register(id string, conn *websocket.Conn) *AgentSession {
	session := NewAgentSession(id, conn)

	h.mu.Lock()
	h.sessions[id] = session
	h.mu.Unlock()

	return session
}

func (h *Hub) Deregister(id string) {
	h.mu.Lock()
	session, exists := h.sessions[id]
	if exists {
		delete(h.sessions, id)
	}
	h.mu.Unlock()

	if exists {
		session.Close("session terminated by server")
	}
}

func (h *Hub) GetSession(id string) (*AgentSession, error) {
	h.mu.RLock()
	session, exists := h.sessions[id]
	h.mu.RUnlock()

	if !exists {
		return nil, errors.New("requested endpoint agent is currently offline")
	}
	return session, nil
}

func (h *Hub) CloseAll(reason string) {
	h.mu.RLock()
	sessions := make([]*AgentSession, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, session)
	}
	h.mu.RUnlock()

	for _, session := range sessions {
		session.Close(reason)
	}
}
