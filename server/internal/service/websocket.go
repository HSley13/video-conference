package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"video-conference/internal/database"
	"video-conference/internal/repository"
)

const (
	messageTypeSignal      = "signal"
	messageTypeChat        = "chat"
	messageTypeFileMeta    = "file_meta"
	messageTypeFileChunk   = "file_chunk"
	messageTypeError       = "error"
	messageTypeJoin        = "join"
	messageTypeLeave       = "leave"
	websocketWriteDeadline = 10 * time.Second
)

type WebSocketService struct {
	redisRepo repository.WSRepository
	roomRepo  repository.RoomRepository
	userRepo  repository.UserRepository
}

func NewWebSocketService(redisRepo repository.WSRepository, roomRepo repository.RoomRepository, userRepo repository.UserRepository) *WebSocketService {
	return &WebSocketService{
		redisRepo: redisRepo,
		roomRepo:  roomRepo,
		userRepo:  userRepo,
	}
}

func (s *WebSocketService) HandleConnection(conn *websocket.Conn, userID uuid.UUID, roomID uuid.UUID) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Validate room existence and activity
	room, err := s.roomRepo.GetRoom(roomID)
	if err != nil || !room.IsActive {
		sendError(conn, "invalid room", nil)
		return
	}

	// Add participant to room
	participant, err := s.roomRepo.AddParticipant(&database.Participant{
		UserID:   userID,
		RoomID:   roomID,
		JoinedAt: time.Now(),
	}, roomID)
	if err != nil {
		sendError(conn, "failed to join room", err)
		return
	}
	defer s.roomRepo.RemoveParticipant(participant.ID)

	// Get user details for messaging
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		sendError(conn, "user not found", err)
		return
	}

	// Subscribe to room and personal channels
	messages := s.redisRepo.Subscribe(ctx, userID)
	defer s.redisRepo.Unsubscribe(ctx, userID)

	// Notify room about new participant
	s.broadcastSystemMessage(roomID, messageTypeJoin, user.Username)

	// Handle message processing
	go s.handleIncomingMessages(ctx, conn, roomID, userID, user.Username)
	s.handleOutgoingMessages(ctx, conn, messages)

	// Notify room about participant leaving
	s.broadcastSystemMessage(roomID, messageTypeLeave, user.Username)
}

func (s *WebSocketService) handleIncomingMessages(
	ctx context.Context,
	conn *websocket.Conn,
	roomID uuid.UUID,
	userID uuid.UUID,
	username string,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					log.Printf("WebSocket closed unexpectedly: %v", err)
				}
				return
			}

			if msgType != websocket.TextMessage {
				sendError(conn, "binary messages not supported", nil)
				continue
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg, &payload); err != nil {
				sendError(conn, "invalid message format", err)
				continue
			}

			if err := s.validateMessagePayload(payload); err != nil {
				sendError(conn, err.Error(), nil)
				continue
			}

			switch payload["type"] {
			case messageTypeSignal:
				s.handleWebRTCSignal(roomID, userID, payload)
			case messageTypeChat:
				s.handleChatMessage(roomID, userID, username, payload)
			case messageTypeFileMeta, messageTypeFileChunk:
				s.handleFileTransfer(roomID, userID, payload)
			default:
				sendError(conn, "unknown message type", nil)
			}
		}
	}
}

func (s *WebSocketService) handleWebRTCSignal(roomID uuid.UUID, senderID uuid.UUID, payload map[string]interface{}) {
	targetID, err := uuid.Parse(payload["target"].(string))
	if err != nil {
		log.Printf("Invalid target ID: %v", err)
		return
	}

	// // Validate target user exists in room
	// if exists, _ := s.roomRepo.ParticipantExists(roomID, targetID); !exists {
	// 	log.Printf("Target user %s not in room %s", targetID, roomID)
	// 	return
	// }

	// Forward signal to target user's private channel
	targetChannel := fmt.Sprintf("%s:%s", roomID.String(), targetID.String())
	message := map[string]interface{}{
		"type":   messageTypeSignal,
		"sender": senderID.String(),
		"data":   payload["data"],
	}

	if err := s.redisRepo.PublishMessage(context.Background(), targetChannel, message); err != nil {
		log.Printf("Failed to publish signal: %v", err)
	}
}

func (s *WebSocketService) handleChatMessage(roomID uuid.UUID, userID uuid.UUID, username string, payload map[string]interface{}) {
	message := map[string]interface{}{
		"type":      messageTypeChat,
		"sender":    userID.String(),
		"username":  username,
		"message":   payload["message"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.redisRepo.PublishMessage(context.Background(), roomID.String(), message); err != nil {
		log.Printf("Failed to publish chat message: %v", err)
	}
}

func (s *WebSocketService) handleFileTransfer(roomID uuid.UUID, userID uuid.UUID, payload map[string]interface{}) {
	// Validate file metadata
	if payload["type"] == messageTypeFileMeta {
		if payload["name"] == nil || payload["size"] == nil || payload["type"] == nil {
			log.Println("Invalid file metadata")
			return
		}
	}

	// Add sender info and forward to room
	payload["sender"] = userID.String()
	if err := s.redisRepo.PublishMessage(context.Background(), roomID.String(), payload); err != nil {
		log.Printf("Failed to publish file transfer: %v", err)
	}
}

func (s *WebSocketService) handleOutgoingMessages(
	ctx context.Context,
	conn *websocket.Conn,
	messages <-chan *redis.Message,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			if err := conn.SetWriteDeadline(time.Now().Add(websocketWriteDeadline)); err != nil {
				log.Printf("SetWriteDeadline failed: %v", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					log.Printf("WriteMessage error: %v", err)
				}
				return
			}
		}
	}
}

func (s *WebSocketService) broadcastSystemMessage(roomID uuid.UUID, messageType string, username string) {
	message := map[string]interface{}{
		"type":      messageType,
		"username":  username,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.redisRepo.PublishMessage(context.Background(), roomID.String(), message); err != nil {
		log.Printf("Failed to broadcast system message: %v", err)
	}
}

func (s *WebSocketService) validateMessagePayload(payload map[string]interface{}) error {
	if payload["type"] == nil {
		return fmt.Errorf("message type is required")
	}

	switch payload["type"] {
	case messageTypeSignal:
		if payload["target"] == nil || payload["data"] == nil {
			return fmt.Errorf("signal messages require target and data fields")
		}
	case messageTypeChat:
		if payload["message"] == nil {
			return fmt.Errorf("chat messages require a message field")
		}
	case messageTypeFileMeta:
		if payload["name"] == nil || payload["size"] == nil || payload["type"] == nil {
			return fmt.Errorf("file metadata requires name, size, and type fields")
		}
	case messageTypeFileChunk:
		if payload["chunk"] == nil || payload["sequence"] == nil {
			return fmt.Errorf("file chunks require chunk and sequence fields")
		}
	}

	return nil
}

func sendError(conn *websocket.Conn, message string, err error) {
	log.Printf("WebSocket error: %s (%v)", message, err)
	conn.WriteJSON(map[string]interface{}{
		"type":    messageTypeError,
		"message": message,
		"error":   err.Error(),
	})
	conn.Close()
}
