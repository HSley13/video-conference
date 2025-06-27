package repositories

import (
	"context"
	"errors"
	"fmt"

	"video-conference/models"

	"gorm.io/gorm"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) CreateUser(ctx context.Context, u *models.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).
		First(&u, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) CreateSession(ctx context.Context, s *models.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *UserRepository) GetSession(ctx context.Context, sessID string) (*models.Session, error) {
	var s models.Session
	err := r.db.WithContext(ctx).
		First(&s, "id = ?", sessID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *UserRepository) GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error) {
	var s models.Session
	err := r.db.WithContext(ctx).
		First(&s, "user_id = ?", userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *UserRepository) DeleteSession(ctx context.Context, sessID string) error {
	return r.db.WithContext(ctx).
		Delete(&models.Session{}, "id = ?", sessID).Error
}
