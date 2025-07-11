package services

import (
	"context"
	"errors"
	"time"

	"video-conference/db_aws"
	"video-conference/models"
	"video-conference/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo *repositories.UserRepository, secret string) *AuthService {
	return &AuthService{userRepo: repo, jwtSecret: []byte(secret)}
}

func (s *AuthService) Register(ctx context.Context, username string, email string, password string) (access string, refresh string, userID string, err error) {

	hash, err := db_aws.HashPassword(password)
	if err != nil {
		return "", "", "", err
	}

	user := &models.User{
		UserName:     username,
		ImgUrl:       "https://via.placeholder.com/150",
		Email:        email,
		HashPassword: hash,
	}
	if err = s.userRepo.CreateUser(ctx, user); err != nil {
		return "", "", "", err
	}

	access, err = s.generateAccessToken(user.ID.String())
	if err != nil {
		return "", "", "", err
	}
	refresh, err = s.generateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", "", err
	}

	_ = s.storeSession(ctx, user.ID, refresh)

	return access, refresh, user.ID.String(), nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (access string, refresh string, userID string, err error) {

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || db_aws.VerifyPassword(password, user.HashPassword) != nil {
		return "", "", "", errors.New("invalid credentials")
	}

	access, err = s.generateAccessToken(user.ID.String())
	if err != nil {
		return "", "", "", err
	}
	refresh, err = s.generateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", "", err
	}

	_ = s.storeSession(ctx, user.ID, refresh)

	return access, refresh, user.ID.String(), nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (newAccess string, err error) {

	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	uid := claims["sub"].(string)

	sess, err := s.userRepo.GetSessionByUserID(ctx, uid)
	if err != nil || sess.RefreshToken != refreshToken || time.Now().After(sess.ExpiresAt) {
		return "", errors.New("expired session")
	}
	return s.generateAccessToken(uid)
}

func (s *AuthService) generateAccessToken(uid string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString(s.jwtSecret)
}

func (s *AuthService) generateRefreshToken(uid string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tok string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return s.jwtSecret, nil })
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func (s *AuthService) storeSession(ctx context.Context, uid uuid.UUID, refresh string) error {
	sess := &models.Session{
		UserID:       uid,
		RefreshToken: refresh,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	return s.userRepo.CreateSession(ctx, sess)
}

func (s *AuthService) SetAuthCookies(c *fiber.Ctx, access string, refresh string, userID string) {
	if access != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    access,
			Expires:  time.Now().Add(15 * time.Minute),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		})
	}
	if refresh != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/video-conference/auth/refresh",
		})
	}
	if userID != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "videoConferenceUserId",
			Value:    userID,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HTTPOnly: false,
			Secure:   true,
			SameSite: "Lax",
		})
	}
}
