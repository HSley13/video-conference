package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OwnerID         uuid.UUID `gorm:"type:uuid;not null;index"             json:"owner_id"`
	Title           string    `gorm:"size:100;not null"                   json:"title"`
	Description     string    `gorm:"size:255;not null"                   json:"description"`
	MaxParticipants int       `gorm:"not null;default:10"                json:"max_participants"`
	IsActive        bool      `gorm:"not null;default:true"              json:"is_active"`
	CreatedAt       time.Time `gorm:"not null;default:now()"             json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"             json:"updated_at"`
}

func (*Room) TableName() string { return "rooms" }

type Participant struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"room_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user"`
	SessionID uuid.UUID  `gorm:"type:uuid;not null"                             json:"session_id"`
	JoinedAt  time.Time  `gorm:"not null;default:now()"                         json:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
}

func (*Participant) TableName() string { return "participants" }

type Code struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user"`
	Code     string    `gorm:"type:text;not null"                             json:"code"`
	ExpireAt time.Time `gorm:"not null;index"                                 json:"expire_at"`
}

func (*Code) TableName() string { return "codes" }
