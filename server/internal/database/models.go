package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	Password     string `gorm:"not null"`
	RefreshToken string
	IsActive     bool `gorm:"default:true"`
}

type Room struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	OwnerID     uint `gorm:"not null"`
	IsActive    bool `gorm:"default:true"`
}

type Participant struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint `gorm:"not null"`
	RoomID   uint `gorm:"not null"`
	JoinedAt time.Time
	LeftAt   *time.Time
	User     User `gorm:"foreignKey:UserID"`
	Room     Room `gorm:"foreignKey:RoomID"`
}
