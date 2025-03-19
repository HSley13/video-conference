package seed

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM rooms")
	db.Exec("DELETE FROM sessions")
	db.Exec("DELETE FROM participants")
}
