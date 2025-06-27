package services

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"video-conference/models"
	"video-conference/repositories"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type WebSocketService struct {
	roomRepo *repositories.RoomRepository
	userRepo *repositories.UserRepository

	connections map[string]map[string]*websocket.Conn
	mutex       sync.RWMutex

	iceServers     []string
	maxConnections int
}

func NewWebSocketService(
	roomRepo *repositories.RoomRepository,
	userRepo *repositories.UserRepository,
	iceServers []string,
	maxConns int,
) *WebSocketService {
	return &WebSocketService{
		roomRepo:       roomRepo,
		userRepo:       userRepo,
		connections:    make(map[string]map[string]*websocket.Conn),
		iceServers:     iceServers,
		maxConnections: maxConns,
	}
}

func (s *WebSocketService) HandleConnection(
	ctx context.Context,
	conn *websocket.Conn,
	roomID, userID string,
) {
	s.mutex.Lock()
	roomMap, ok := s.connections[roomID]
	if !ok {
		roomMap = make(map[string]*websocket.Conn)
		s.connections[roomID] = roomMap
	}
	roomMap[userID] = conn
	peerCount := len(roomMap)
	s.mutex.Unlock()

	log.Printf("[ROOM %s] socket open → %s (peers=%d)", roomID, userID, peerCount)
	defer s.cleanupConnection(ctx, roomID, userID)

	if peerCount > s.maxConnections {
		_ = conn.WriteJSON(fiberMap("error", "room full"))
		return
	}

	if err := s.roomRepo.AddParticipant(ctx, roomID, userID); err != nil {
		log.Printf("AddParticipant: %v", err)
		return
	}
	defer s.roomRepo.RemoveParticipant(ctx, roomID, userID)

	_ = conn.WriteJSON(fiberMap("type", "iceServers", "iceServers", s.iceServers))

	user := s.ensureUser(ctx, userID)

	join := fiberMap(
		"type", "user-joined",
		"userID", user.ID,
		"userName", user.Name,
		"userPhoto", user.ImgUrl,
		"sender", userID,
	)
	_ = s.roomRepo.PublishMessage(ctx, roomID, join)

	sub, err := s.roomRepo.SubscribeToRoom(ctx, roomID)
	if err != nil {
		log.Printf("SubscribeToRoom: %v", err)
		return
	}
	defer s.roomRepo.UnsubscribeFromRoom(ctx, sub)

	go s.readFromClient(ctx, conn, roomID, userID, user)

	for msg := range sub.Channel {
		var payload map[string]any
		if json.Unmarshal([]byte(msg.Payload), &payload) != nil {
			continue
		}
		if payload["sender"] == userID {
			continue
		}
		_ = conn.WriteJSON(payload)
	}

	leave := fiberMap(
		"type", "user-left",
		"userID", user.ID,
		"userName", user.Name,
		"userPhoto", user.ImgUrl,
		"sender", userID,
	)
	_ = s.roomRepo.PublishMessage(ctx, roomID, leave)
}

func (s *WebSocketService) readFromClient(
	ctx context.Context,
	conn *websocket.Conn,
	roomID, userID string,
	user *models.User,
) {
	for {
		mt, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("read error %s: %v", userID, err)
			}
			return
		}
		if mt != websocket.TextMessage {
			continue
		}

		var payload map[string]any
		if json.Unmarshal(raw, &payload) != nil {
			continue
		}

		switch payload["type"] {
		case "chat-message":
			s.handleChat(ctx, roomID, userID, payload)
		case "offer", "answer", "ice-candidate":
			s.forwardSDP(roomID, userID, payload)
		}
	}
}

func (s *WebSocketService) handleChat(
	ctx context.Context,
	roomID, userID string,
	raw map[string]any,
) {
	msg := raw
	if inner, ok := raw["message"].(map[string]any); ok {
		msg = inner
	}

	payload := fiberMap(
		"type", "chat-message",
		"id", msg["id"],
		"text", msg["text"],
		"time", msg["time"],
		"user", msg["user"],
		"sender", userID,
	)

	_ = s.roomRepo.PublishMessage(ctx, roomID, payload)
}

func (s *WebSocketService) forwardSDP(roomID, from string, payload map[string]any) {
	to, ok := payload["to"].(string)
	if !ok || to == "" {
		return
	}

	s.mutex.RLock()
	target, exists := s.connections[roomID][to]
	s.mutex.RUnlock()
	if !exists {
		return
	}

	payload["from"] = from
	payload["sender"] = from

	_ = target.WriteJSON(payload)
}

func (s *WebSocketService) ensureUser(ctx context.Context, uid string) *models.User {
	if u, err := s.userRepo.GetUserByID(ctx, uid); err == nil {
		return u
	}

	parsed, _ := uuid.Parse(uid)
	user := &models.User{
		ID:     parsed,
		Name:   "Anonymous",
		ImgUrl: "https://via.placeholder.com/150",
	}
	_ = s.userRepo.CreateUser(ctx, user)
	return user
}

func (s *WebSocketService) cleanupConnection(ctx context.Context, roomID, uid string) {
	s.mutex.Lock()
	if roomMap, ok := s.connections[roomID]; ok {
		delete(roomMap, uid)
		if len(roomMap) == 0 {
			delete(s.connections, roomID)
		}
	}
	s.mutex.Unlock()

	_ = s.roomRepo.RemoveParticipant(ctx, roomID, uid)
	log.Printf("[ROOM %s] socket closed ← %s", roomID, uid)
}

func fiberMap(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}
