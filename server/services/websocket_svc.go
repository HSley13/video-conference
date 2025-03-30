package services

import (
	"context"
	"encoding/json"
	"errors"
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

	if err := s.roomRepo.AddParticipant(ctx, roomID, userID); err != nil {
		log.Printf("Error adding user %s to room %s: %v\n", userID, roomID, err)
		return
	}
	defer s.roomRepo.RemoveParticipant(ctx, roomID, userID)

	iceConfig := map[string]interface{}{
		"type":       "iceServers",
		"iceServers": s.iceServers,
	}
	if err := conn.WriteJSON(iceConfig); err != nil {
		log.Printf("Error sending ICE config to user %s: %v\n", userID, err)
		return
	}

	messages := s.roomRepo.SubscribeToRoom(ctx, roomID)

	userIDUuid, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid userID %s: %v\n", userID, err)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error fetching user %s: %v\n", userID, err)

		newUser := &models.User{
			ID:     userIDUuid,
			Name:   "Test",
			ImgUrl: "https://via.placeholder.com/150",
		}
		s.userRepo.CreateUser(ctx, newUser)

		// TODO:
		// return later when u add the signin form
		// return
	}

	log.Printf("User %s (%s) has joined room %s\n", userID, user.Name, roomID)

	s.roomRepo.PublishMessage(ctx, roomID, map[string]interface{}{
		"type": "user-joined",
		"user": map[string]interface{}{
			"id":     user.ID,
			"name":   user.Name,
			"imgUrl": user.ImgUrl,
		},
	})

	log.Printf("Message user-joined has been sent %s\n", user)

	go func() {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil || mt == websocket.CloseMessage {
				log.Printf("User %s has left the WebSocket connection: %v\n", userID, err)
				break
			}

			if mt == websocket.TextMessage {
				var payload map[string]interface{}
				if err := json.Unmarshal(msg, &payload); err == nil {
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

						s.roomRepo.PublishMessage(ctx, roomID, map[string]interface{}{
							"type":    "chat-message",
							"message": chatMessage,
						})

					case "offer", "answer", "ice-candidate":
						log.Println("offer, answer, ice-candidate message received")
						if to, ok := payload["to"].(string); ok {
							s.mutex.RLock()
							if recipientConn, exists := s.connections[roomID][to]; exists {
								log.Printf("Forwarding %s message from user %s to user %s\n", payload["type"], userID, to)
								recipientConn.WriteJSON(payload)
							} else {
								log.Printf("Recipient %s not found in room %s\n", to, roomID)
							}
							s.mutex.RUnlock()
						}
					}
				}
			}
		}
	}()

	for msg := range messages {
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

	log.Printf("User %s (%s) has left room %s\n", userID, user.Name, roomID)

	s.roomRepo.PublishMessage(ctx, roomID, map[string]interface{}{
		"type": "user-left",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"photo": user.ImgUrl,
		},
	})
}

func (s *WebSocketService) CanJoinRoom(ctx context.Context, roomID string) (bool, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room %s: %v\n", roomID, err)
		return false, err
	}

	if !room.IsActive {
		log.Printf("Room %s is not active\n", roomID)
		return false, errors.New("room is not active")
	}

	participants, err := s.roomRepo.GetParticipants(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching participants for room %s: %v\n", roomID, err)
		return false, err
	}

	if len(participants) >= room.MaxParticipants {
		log.Printf("Room %s is full\n", roomID)
		return false, errors.New("room is full")
	}

	log.Printf("User can join room %s\n", roomID)
	return true, nil
}

func (s *WebSocketService) NotifyUserJoined(ctx context.Context, roomID string, userID string) error {
	log.Printf("Notifying room %s that user %s has joined\n", roomID, userID)
	payload := map[string]interface{}{
		"type":   "user-joined",
		"userID": userID,
	}
	return s.roomRepo.PublishMessage(ctx, roomID, payload)
}

func (s *WebSocketService) NotifyUserLeft(ctx context.Context, roomID string, userID string) error {
	log.Printf("Notifying room %s that user %s has left\n", roomID, userID)
	payload := map[string]interface{}{
		"type":   "user-left",
		"userID": userID,
	}
	return s.roomRepo.PublishMessage(ctx, roomID, payload)
}
