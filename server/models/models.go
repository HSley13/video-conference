package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	RefreshToken string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

type Room struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OwnerID         uuid.UUID `gorm:"type:uuid;not null"`
	Name            string    `gorm:"not null"`
	MaxParticipants int       `gorm:"not null;default:10"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	IsActive        bool `gorm:"not null;default:true"`
}

type Participant struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoomID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	JoinedAt  time.Time `gorm:"not null"`
	LeftAt    *time.Time
	SessionID uuid.UUID `gorm:"type:uuid;not null"`
}
