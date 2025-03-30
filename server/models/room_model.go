package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OwnerID         uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	Name            string    `gorm:"not null;size:100" json:"name"`
	MaxParticipants int       `gorm:"not null;default:10" json:"max_participants"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
	IsActive        bool      `gorm:"not null;default:true"`
}

type Participant struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID  `gorm:"type:uuid;not null" json:"room_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"user"`
	JoinedAt  time.Time  `gorm:"not null" json:"joined_at"`
	LeftAt    *time.Time `gorm:"null" json:"left_at"`
	SessionID uuid.UUID  `gorm:"type:uuid;not null"`
}

type Code struct {
	ID       string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID   string    `gorm:"not null;type:uuid;index" json:"user_id"`
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"user"`
	Code     string    `gorm:"not null;type:text" json:"code"`
	ExpireAt time.Time `gorm:"not null" json:"expire_at"`
}
