package repositories

import (
	"context"
	"video-conference/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) CreateUser(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := models.User{}
	res := r.db.WithContext(ctx).Where("email = ?", email).First(&u)
	return &u, res.Error
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	u := models.User{}
	res := r.db.WithContext(ctx).First(&u, "id = ?", id)
	return &u, res.Error
}

func (r *UserRepository) CreateSession(ctx context.Context, s *models.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *UserRepository) GetSession(ctx context.Context, sessID string) (*models.Session, error) {
	s := models.Session{}
	res := r.db.WithContext(ctx).First(&s, "id = ?", sessID)
	return &s, res.Error
}

func (r *UserRepository) GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error) {
	s := models.Session{}
	res := r.db.WithContext(ctx).First(&s, "user_id = ?", userID)
	return &s, res.Error
}

func (r *UserRepository) DeleteSession(ctx context.Context, sessID string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", sessID).Error
}
