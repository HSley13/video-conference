package services

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"video-conference/repositories"

	"github.com/gofiber/websocket/v2"
)

type WebSocketService struct {
	roomRepo       *repositories.RoomRepository
	connections    map[string]map[string]*websocket.Conn
	mutex          sync.RWMutex
	iceServers     []string
	maxConnections int
}

func NewWebSocketService(roomRepo *repositories.RoomRepository, iceServers []string, maxConnections int) *WebSocketService {
	return &WebSocketService{
		roomRepo:       roomRepo,
		connections:    make(map[string]map[string]*websocket.Conn),
		iceServers:     iceServers,
		maxConnections: maxConnections,
	}
}

func (s *WebSocketService) HandleConnection(ctx context.Context, conn *websocket.Conn, roomID, userID string) {
	s.mutex.Lock()
	if _, exists := s.connections[roomID]; !exists {
		s.connections[roomID] = make(map[string]*websocket.Conn)
	}
	s.connections[roomID][userID] = conn
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.connections[roomID], userID)
		if len(s.connections[roomID]) == 0 {
			delete(s.connections, roomID)
		}
		s.mutex.Unlock()
	}()

	if err := s.roomRepo.AddParticipant(ctx, roomID, userID); err != nil {
		return
	}
	defer s.roomRepo.RemoveParticipant(ctx, roomID, userID)

	iceConfig := map[string]interface{}{
		"type":       "iceServers",
		"iceServers": s.iceServers,
	}
	if err := conn.WriteJSON(iceConfig); err != nil {
		return
	}

	messages := s.roomRepo.SubscribeToRoom(ctx, roomID)

	go func() {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil || mt == websocket.CloseMessage {
				break
			}

			if mt == websocket.TextMessage {
				var payload map[string]interface{}
				if err := json.Unmarshal(msg, &payload); err == nil {
					payload["sender"] = userID
					if modifiedMsg, err := json.Marshal(payload); err == nil {
						s.roomRepo.PublishMessage(ctx, roomID, string(modifiedMsg))
					}
				}
			}
		}
	}()

	for msg := range messages {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}

		if sender, ok := payload["sender"].(string); ok && sender == userID {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			break
		}
	}
}

func (s *WebSocketService) CanJoinRoom(ctx context.Context, roomID string) (bool, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return false, err
	}

	if !room.IsActive {
		return false, errors.New("room is not active")
	}

	participants, err := s.roomRepo.GetParticipants(ctx, roomID)
	if err != nil {
		return false, err
	}

	if len(participants) >= room.MaxParticipants {
		return false, errors.New("room is full")
	}

	return true, nil
}
