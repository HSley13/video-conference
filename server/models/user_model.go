package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"size:100;not null"                           json:"name"`
	Email        string    `gorm:"size:100;unique;not null"                    json:"email"`
	ImgUrl       string    `gorm:"size:255;not null"                           json:"img_url"`
	HashPassword string    `gorm:"size:255;not null"                           json:"hash_password"`
	CreatedAt    time.Time `gorm:"not null;default:now()"                      json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"                      json:"updated_at"`
}

func (*User) TableName() string { return "users" }

type Session struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user"`
	RefreshToken string    `gorm:"type:char(255);not null"                        json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"not null;index"                                 json:"expires_at"`
	CreatedAt    time.Time `gorm:"not null;default:now()"                         json:"created_at"`
}

func (*Session) TableName() string { return "sessions" }
