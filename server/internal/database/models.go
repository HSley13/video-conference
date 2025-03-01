package database

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	FirstName    string    `gorm:"not null" type:"varchar(100)" json:"first_name"`
	LastName     string    `gorm:"not null" type:"varchar(100)" json:"last_name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	HashPassword string    `gorm:"not null;size:100" json:"hash_password"`
	RefreshToken string    `gorm:"size:100" json:"refresh_token"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
}

type Room struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Description  string    `json:"description"`
	OwnerID      uuid.UUID `gorm:"not null;type:uuid;index" json:"owner_id"`
	Owner        User      `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"owner"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	Participants []Participant
}

type Participant struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID   uuid.UUID `gorm:"not null;type:uuid;index" json:"user_id"`
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"user"`
	RoomID   uuid.UUID `gorm:"not null;type:uuid;index" json:"room_id"`
	Room     Room      `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"room"`
	JoinedAt time.Time `gorm:"not null;default:now()" json:"joined_at"`
	LeftAt   time.Time `gorm:"not null;default:now()" json:"left_at"`
}
