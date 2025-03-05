package services

import (
	"context"
	"errors"
	"time"
	"video-conference/models"
	"video-conference/repositories"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := s.generateAccessToken(user.ID.String())
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", err
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) generateAccessToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	userID := claims["sub"].(string)
	session, err := s.userRepo.GetSession(ctx, userID)
	if err != nil || session.RefreshToken != refreshToken {
		return "", errors.New("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}

	return s.generateAccessToken(userID)
}
