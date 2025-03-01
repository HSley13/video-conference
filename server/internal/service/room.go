package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"video-conference/internal/database"
	"video-conference/internal/repository"
)

type RoomService struct {
	roomRepo *repository.RoomRepository
}

func NewRoomService(roomRepo *repository.RoomRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo}
}

func NewWebSocketService(redisRepo *repository.WSRepository, roomRepo *repository.RoomRepository, userRepo *repository.UserRepository) *WebSocketService {
	return &WebSocketService{redisRepo: redisRepo, roomRepo: roomRepo, userRepo: userRepo}
}

func (s *RoomService) CreateRoom(ctx context.Context, userID uuid.UUID, name, description string) (*database.Room, error) {
	if name == "" {
		return nil, errors.New("room name is required")
	}

	room := &database.Room{
		Name:        name,
		Description: description,
		OwnerID:     userID,
		IsActive:    true,
	}

	if err := s.roomRepo.CreateRoom(room); err != nil {
		return nil, errors.New("failed to create room")
	}

	return room, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, userID uuid.UUID, roomID uuid.UUID) (*database.Participant, error) {
	room, err := s.roomRepo.GetRoom(roomID)
	if err != nil || !room.IsActive {
		return nil, errors.New("room not available")
	}

	participant := &database.Participant{
		UserID:   userID,
		RoomID:   roomID,
		JoinedAt: time.Now(),
	}

	if err := s.roomRepo.AddParticipant(participant); err != nil {
		return nil, errors.New("failed to join room")
	}

	return participant, nil
}

func (s *RoomService) ListActiveRooms(ctx context.Context) ([]database.Room, error) {
	return s.roomRepo.GetActiveRooms()
}

func (s *RoomService) EndRoom(ctx context.Context, userID uuid.UUID, roomID uuid.UUID) error {
	room, err := s.roomRepo.GetRoom(roomID)
	if err != nil || room.OwnerID != userID {
		return errors.New("unauthorized operation")
	}

	return s.roomRepo.DeactivateRoom(roomID)
}
