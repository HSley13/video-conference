package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
	"video-conference/models"
	"video-conference/repositories"

	"github.com/gofiber/websocket/v2"
)

type WebSocketService struct {
	roomRepo       *repositories.RoomRepository
	userRepo       *repositories.UserRepository
	connections    map[string]map[string]*websocket.Conn
	mutex          sync.RWMutex
	iceServers     []string
	maxConnections int
}

func NewWebSocketService(roomRepo *repositories.RoomRepository, userRepo *repositories.UserRepository, iceServers []string, maxConnections int) *WebSocketService {
	return &WebSocketService{
		roomRepo:       roomRepo,
		userRepo:       userRepo,
		connections:    make(map[string]map[string]*websocket.Conn),
		iceServers:     iceServers,
		maxConnections: maxConnections,
	}
}

func (s *WebSocketService) HandleConnection(ctx context.Context, conn *websocket.Conn, roomID string, userID string) {
	log.Printf("User %s is connecting to room %s\n", userID, roomID)

	s.mutex.Lock()
	if _, exists := s.connections[roomID]; !exists {
		s.connections[roomID] = make(map[string]*websocket.Conn)
	}
	s.connections[roomID][userID] = conn
	connectionCount := len(s.connections[roomID])
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.connections[roomID], userID)
		if len(s.connections[roomID]) == 0 {
			delete(s.connections, roomID)
		}
		s.mutex.Unlock()
		log.Printf("User %s has disconnected from room %s\n", userID, roomID)
	}()

	if connectionCount > s.maxConnections {
		log.Printf("Room %s has reached maximum connections (%d)\n", roomID, s.maxConnections)
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "Room is full",
		})
		conn.Close()
		return
	}

	if err := s.roomRepo.AddParticipant(ctx, roomID, userID); err != nil {
		log.Printf("Error adding user %s to room %s: %v\n", userID, roomID, err)
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "Failed to join room",
		})
		conn.Close()
		return
	}
	defer func() {
		if err := s.roomRepo.RemoveParticipant(ctx, roomID, userID); err != nil {
			log.Printf("Error removing user %s from room %s: %v\n", userID, roomID, err)
		}
	}()

	iceConfig := map[string]interface{}{
		"type":       "iceServers",
		"iceServers": s.iceServers,
	}
	if err := conn.WriteJSON(iceConfig); err != nil {
		log.Printf("Error sending ICE config to user %s: %v\n", userID, err)
		return
	}

	subscription, err := s.roomRepo.SubscribeToRoom(ctx, roomID)
	if err != nil {
		log.Printf("Error subscribing to room %s: %v\n", roomID, err)
		return
	}
	defer func() {
		if err := s.roomRepo.UnsubscribeFromRoom(ctx, subscription); err != nil {
			log.Printf("Error unsubscribing from room %s: %v\n", roomID, err)
		}
	}()

	userIDUuid, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid userID %s: %v\n", userID, err)
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "Invalid user ID",
		})
		return
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error fetching user %s, creating temporary user: %v\n", userID, err)

		user = &models.User{
			ID:     userIDUuid,
			Name:   "Anonymous",
			ImgUrl: "https://via.placeholder.com/150",
		}
		if err := s.userRepo.CreateUser(ctx, user); err != nil {
			log.Printf("Error creating temporary user %s: %v\n", userID, err)
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"message": "Failed to create user profile",
			})
			return
		}
	}

	log.Printf("User %s (%s) has joined room %s\n", userID, user.Name, roomID)

	joinMessage := map[string]interface{}{
		"type": "user-joined",
		"user": map[string]interface{}{
			"id":     user.ID,
			"name":   user.Name,
			"imgUrl": user.ImgUrl,
		},
	}
	if err := s.roomRepo.PublishMessage(ctx, roomID, joinMessage); err != nil {
		log.Printf("Error publishing join message for user %s: %v\n", userID, err)
	}

	go s.handleIncomingMessages(ctx, conn, roomID, userID, user)

	for msg := range subscription.Channel {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Printf("Error unmarshaling message in room %s: %v\n", roomID, err)
			continue
		}

		if sender, ok := payload["sender"].(string); ok && sender == userID {
			continue
		}

		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("Error sending message to user %s: %v\n", userID, err)
			break
		}
	}

	leaveMessage := map[string]interface{}{
		"type": "user-left",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"photo": user.ImgUrl,
		},
	}
	if err := s.roomRepo.PublishMessage(ctx, roomID, leaveMessage); err != nil {
		log.Printf("Error publishing leave message for user %s: %v\n", userID, err)
	}

	log.Printf("User %s (%s) has left room %s\n", userID, user.Name, roomID)
}

func (s *WebSocketService) handleIncomingMessages(ctx context.Context, conn *websocket.Conn, roomID, userID string, user *models.User) {
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for user %s: %v\n", userID, err)
			}
			break
		}

		if mt == websocket.TextMessage {
			var payload map[string]interface{}
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Printf("Error unmarshaling message from user %s: %v\n", userID, err)
				continue
			}

			switch payload["type"] {
			case "chat-message":
				log.Printf("User %s sent a chat message in room %s\n", userID, roomID)

				chatMessage := map[string]interface{}{
					"id":   time.Now().UnixNano(),
					"text": payload["text"],
					"time": time.Now().Format(time.RFC3339),
					"user": map[string]interface{}{
						"id":    user.ID,
						"name":  user.Name,
						"photo": user.ImgUrl,
					},
				}

				if err := s.roomRepo.PublishMessage(ctx, roomID, map[string]interface{}{
					"type":    "chat-message",
					"message": chatMessage,
					"sender":  userID,
				}); err != nil {
					log.Printf("Error publishing chat message: %v\n", err)
				}

			case "offer", "answer", "ice-candidate":
				if to, ok := payload["to"].(string); ok {
					s.mutex.RLock()
					recipientConn, exists := s.connections[roomID][to]
					s.mutex.RUnlock()

					if exists {
						log.Printf("Forwarding %s message from user %s to user %s\n", payload["type"], userID, to)
						if err := recipientConn.WriteJSON(payload); err != nil {
							log.Printf("Error forwarding message to user %s: %v\n", to, err)
						}
					} else {
						log.Printf("Recipient %s not found in room %s\n", to, roomID)
					}
				}
			}
		}
	}
}

func (s *WebSocketService) CanJoinRoom(ctx context.Context, roomID string) (bool, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return false, fmt.Errorf("error fetching room: %w", err)
	}

	if !room.IsActive {
		return false, errors.New("room is not active")
	}

	participants, err := s.roomRepo.GetParticipants(ctx, roomID)
	if err != nil {
		return false, fmt.Errorf("error fetching participants: %w", err)
	}

	if len(participants) >= room.MaxParticipants {
		return false, errors.New("room is full")
	}

	return true, nil
}

func (s *WebSocketService) NotifyUserJoined(ctx context.Context, roomID string, userID string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error fetching user: %w", err)
	}

	payload := map[string]interface{}{
		"type": "user-joined",
		"user": map[string]interface{}{
			"id":     user.ID,
			"name":   user.Name,
			"imgUrl": user.ImgUrl,
		},
	}
	return s.roomRepo.PublishMessage(ctx, roomID, payload)
}

func (s *WebSocketService) NotifyUserLeft(ctx context.Context, roomID string, userID string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error fetching user: %w", err)
	}

	payload := map[string]interface{}{
		"type": "user-left",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"photo": user.ImgUrl,
		},
	}
	return s.roomRepo.PublishMessage(ctx, roomID, payload)
}
