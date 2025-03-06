package repositories

import (
	"context"
	"video-conference/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := models.User{}
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	return &user, result.Error
}

func (r *UserRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *UserRepository) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session := models.Session{}
	result := r.db.WithContext(ctx).First(&session, "id = ?", sessionID)
	return &session, result.Error
}

func (r *UserRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", sessionID).Error
}
