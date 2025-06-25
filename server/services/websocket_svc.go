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
	peerCount := len(s.connections[roomID])
	s.mutex.Unlock()

	log.Printf("[ROOM %s] new socket %s (peers=%d)", roomID, userID, peerCount)
	defer s.cleanupConnection(ctx, roomID, userID)

	if peerCount > s.maxConnections {
		_ = conn.WriteJSON(fiberMap("error", "room full"))
		conn.Close()
		return
	}

	_ = s.roomRepo.AddParticipant(ctx, roomID, userID)
	defer s.roomRepo.RemoveParticipant(ctx, roomID, userID)

	_ = conn.WriteJSON(fiberMap("type", "iceServers", "iceServers", s.iceServers))

	user := s.ensureUser(ctx, userID)

	if peerCount == 1 {
		join := fiberMap(
			"type", "user-joined",
			"userID", user.ID, "userName", user.Name, "userPhoto", user.ImgUrl,
			"sender", userID,
		)
		_ = s.roomRepo.PublishMessage(ctx, roomID, join)
		log.Printf("[ROOM %s] broadcast user-joined: %+v", roomID, join)
	}

	sub, err := s.roomRepo.SubscribeToRoom(ctx, roomID)
	if err != nil {
		log.Printf("subscribe error: %v", err)
		return
	}
	defer s.roomRepo.UnsubscribeFromRoom(ctx, sub)

	go s.readFromClient(ctx, conn, roomID, userID, user)

	for msg := range sub.Channel {
		var pl map[string]any
		if json.Unmarshal([]byte(msg.Payload), &pl) != nil {
			continue
		}
		if snd, ok := pl["sender"].(string); ok && snd == userID {
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
		mt, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error %s: %v", userID, err)
			}
			break
		}
		if mt != websocket.TextMessage {
			continue
		}

		var pl map[string]any
		if json.Unmarshal(raw, &pl) != nil {
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

func (s *WebSocketService) handleChat(ctx context.Context, roomID, userID string, in map[string]any) {
	msg := in
	if inner, ok := in["message"].(map[string]any); ok {
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

func (s *WebSocketService) forwardSDP(roomID, from string, pl map[string]any) {
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
	pl["sender"] = from
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

func fiberMap(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}
