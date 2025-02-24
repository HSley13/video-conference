package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"conferencing-app/internal/database"
	"conferencing-app/internal/repository"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/websocket/v2"
)

type WebSocketService struct {
	redisRepo repository.WSRepository
	roomRepo  repository.RoomRepository
	userRepo  repository.UserRepository
}

func (s *WebSocketService) HandleConnection(conn *websocket.Conn, userID uint, roomID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify room exists
	room, err := s.roomRepo.GetRoomByID(roomID)
	if err != nil || !room.IsActive {
		sendError(conn, "Invalid room")
		return
	}

	// Add participant
	participant, err := s.roomRepo.AddParticipant(userID, room.ID)
	if err != nil {
		sendError(conn, "Failed to join room")
		return
	}
	defer s.roomRepo.RemoveParticipant(participant.ID)

	// Setup Redis subscription
	messages := s.redisRepo.Subscribe(ctx, roomID)
	defer s.redisRepo.Unsubscribe(ctx, roomID)

	// Message handling loop
	go s.handleIncomingMessages(ctx, conn, roomID, userID)
	s.handleOutgoingMessages(ctx, conn, messages)
}

func (s *WebSocketService) handleIncomingMessages(ctx context.Context, conn *websocket.Conn, roomID string, userID uint) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg, &payload); err != nil {
				continue
			}

			switch payload["type"] {
			case "signal":
				s.handleWebRTCSignal(roomID, payload)
			case "chat":
				s.handleChatMessage(roomID, userID, payload)
			case "file":
				s.handleFileTransfer(roomID, userID, payload)
			}
		}
	}
}

func (s *WebSocketService) handleOutgoingMessages(ctx context.Context, conn *websocket.Conn, messages <-chan *redis.Message) {
	for {
		select {
		case msg := <-messages:
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func sendError(conn *websocket.Conn, message string) {
	conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"message": message,
	})
	conn.Close()
}
