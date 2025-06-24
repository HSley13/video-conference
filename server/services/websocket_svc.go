package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"video-conference/models"
	"video-conference/repositories"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type WebSocketService struct {
	roomRepo       *repositories.RoomRepository
	userRepo       *repositories.UserRepository
	connections    map[string]map[string]*websocket.Conn
	mutex          sync.RWMutex
	iceServers     []string
	maxConnections int
}

func NewWebSocketService(rRepo *repositories.RoomRepository, uRepo *repositories.UserRepository,
	ice []string, max int) *WebSocketService {
	return &WebSocketService{
		roomRepo:       rRepo,
		userRepo:       uRepo,
		connections:    make(map[string]map[string]*websocket.Conn),
		iceServers:     ice,
		maxConnections: max,
	}
}

func (s *WebSocketService) HandleConnection(
	ctx context.Context,
	conn *websocket.Conn,
	roomID, userID string,
) {
	s.mutex.Lock()
	if _, ok := s.connections[roomID]; !ok {
		s.connections[roomID] = make(map[string]*websocket.Conn)
	}
	s.connections[roomID][userID] = conn
	count := len(s.connections[roomID])
	s.mutex.Unlock()

	log.Printf("[ROOM %s] new socket %s (peers=%d)", roomID, userID, count)
	defer s.cleanupConnection(ctx, roomID, userID)

	if count > s.maxConnections {
		_ = conn.WriteJSON(fiberMap("error", "room full"))
		conn.Close()
		return
	}

	_ = s.roomRepo.AddParticipant(ctx, roomID, userID)
	defer s.roomRepo.RemoveParticipant(ctx, roomID, userID)

	_ = conn.WriteJSON(fiberMap("type", "iceServers", "iceServers", s.iceServers))

	user := s.ensureUser(ctx, userID)

	joinPayload := fiberMap(
		"type", "user-joined",
		"userID", user.ID, "userName", user.Name, "userPhoto", user.ImgUrl,
		"sender", userID,
	)
	_ = s.roomRepo.PublishMessage(ctx, roomID, joinPayload)
	log.Printf("[ROOM %s] broadcast user-joined: %+v", roomID, joinPayload)

	sub, err := s.roomRepo.SubscribeToRoom(ctx, roomID)
	if err != nil {
		log.Printf("subscribe error: %v", err)
		return
	}
	defer s.roomRepo.UnsubscribeFromRoom(ctx, sub)

	go s.readFromClient(ctx, conn, roomID, userID, user)

	for msg := range sub.Channel {
		var pl map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &pl); err != nil {
			continue
		}
		log.Printf("[ROOM %s] relay => %s : %+v", roomID, userID, pl)
		_ = conn.WriteJSON(pl)
	}

	leave := fiberMap(
		"type", "user-left",
		"userID", user.ID, "userName", user.Name, "userPhoto", user.ImgUrl,
		"sender", userID,
	)
	_ = s.roomRepo.PublishMessage(ctx, roomID, leave)
	log.Printf("[ROOM %s] broadcast user-left: %+v", roomID, leave)
}

func (s *WebSocketService) readFromClient(
	ctx context.Context,
	conn *websocket.Conn,
	roomID, userID string,
	user *models.User,
) {
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error %s: %v", userID, err)
			}
			break
		}
		if mt != websocket.TextMessage {
			continue
		}

		var pl map[string]interface{}
		if err := json.Unmarshal(data, &pl); err != nil {
			log.Printf("json parse from %s: %v", userID, err)
			continue
		}

		switch pl["type"] {
		case "chat-message":
			s.handleChat(ctx, roomID, userID, pl)
		case "offer", "answer", "ice-candidate":
			s.forwardSDP(roomID, userID, pl)
		}
	}
}

func (s *WebSocketService) handleChat(ctx context.Context, roomID, userID string, in map[string]interface{}) {
	msg := in
	if inner, ok := in["message"].(map[string]interface{}); ok {
		msg = inner
	}
	payload := fiberMap(
		"type", "chat-message",
		"id", msg["id"], "text", msg["text"], "time", msg["time"], "user", msg["user"],
		"sender", userID,
	)
	log.Printf("[ROOM %s] chat by %s -> %q", roomID, userID, payload["text"])
	_ = s.roomRepo.PublishMessage(ctx, roomID, payload)
}

func (s *WebSocketService) forwardSDP(roomID, from string, pl map[string]interface{}) {
	to, ok := pl["to"].(string)
	if !ok {
		return
	}
	s.mutex.RLock()
	target, exists := s.connections[roomID][to]
	s.mutex.RUnlock()
	if !exists {
		return
	}
	pl["from"] = from
	log.Printf("[ROOM %s] %s → %s (%s)", roomID, from, to, pl["type"])
	_ = target.WriteJSON(pl)
}

func (s *WebSocketService) ensureUser(ctx context.Context, uid string) *models.User {
	if u, err := s.userRepo.GetUserByID(ctx, uid); err == nil {
		return u
	}
	parsed, _ := uuid.Parse(uid)
	u := &models.User{ID: parsed, Name: "Anonymous", ImgUrl: "https://via.placeholder.com/150"}
	_ = s.userRepo.CreateUser(ctx, u)
	return u
}

func (s *WebSocketService) cleanupConnection(ctx context.Context, roomID, uid string) {
	s.mutex.Lock()
	delete(s.connections[roomID], uid)
	if len(s.connections[roomID]) == 0 {
		delete(s.connections, roomID)
	}
	s.mutex.Unlock()
	log.Printf("[ROOM %s] socket closed %s", roomID, uid)
	_ = s.roomRepo.RemoveParticipant(ctx, roomID, uid)
}

func fiberMap(kv ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
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
