package repository

import (
	"conferencing-app/internal/database"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoom(room *database.Room) error {
	return r.db.Create(room).Error
}

func (r *RoomRepository) GetRoom(id uint) (*database.Room, error) {
	var room database.Room
	err := r.db.First(&room, id).Error
	return &room, err
}

func (r *RoomRepository) GetActiveRooms() ([]database.Room, error) {
	var rooms []database.Room
	err := r.db.Where("is_active = ?", true).Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepository) DeactivateRoom(id uint) error {
	return r.db.Model(&database.Room{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *RoomRepository) AddParticipant(participant *database.Participant) error {
	return r.db.Create(participant).Error
}

func (r *RoomRepository) RemoveParticipant(participantID uint) error {
	return r.db.Model(&database.Participant{}).Where("id = ?", participantID).Update("left_at", time.Now()).Error
}
