package repository

import (
	"github.com/google/uuid"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"time"
	"video-conference/internal/database"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func NewWSRepository(client *redis.Client) *WSRepository {
	return &WSRepository{client: client}
}

func (r *RoomRepository) CreateRoom(room *database.Room) error {
	return r.db.Create(room).Error
}

func (r *RoomRepository) GetRoom(id uuid.UUID) (*database.Room, error) {
	var room database.Room
	err := r.db.First(&room, id).Error
	return &room, err
}

func (r *RoomRepository) GetActiveRooms() ([]database.Room, error) {
	var rooms []database.Room
	err := r.db.Where("is_active = ?", true).Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepository) DeactivateRoom(id uuid.UUID) error {
	return r.db.Model(&database.Room{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *RoomRepository) AddParticipant(participant *database.Participant, roomID uuid.UUID) (*database.Participant, error) {
	room := database.Room{}
	r.db.First(&room, roomID)
	participant.RoomID = room.ID

	return participant, r.db.Create(participant).Error
}

func (r *RoomRepository) RemoveParticipant(participantID uuid.UUID) error {
	return r.db.Model(&database.Participant{}).Where("id = ?", participantID).Update("left_at", time.Now()).Error
}
