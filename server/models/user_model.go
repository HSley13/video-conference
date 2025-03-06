package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"not null;size:100;unique" json:"email"`
	HashPassword string    `gorm:"not null;size:100" json:"hash_password"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

type Session struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"not null;type:uuid;index" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE" json:"user"`
	RefreshToken string    `gorm:"not null" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
}
