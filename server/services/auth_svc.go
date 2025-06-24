package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"video-conference/models"
	"video-conference/repositories"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/argon2"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo *repositories.UserRepository, secret string) *AuthService {
	return &AuthService{userRepo: repo, jwtSecret: []byte(secret)}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	name := strings.Split(email, "@")[0]
	user := &models.User{
		Name:         name,
		ImgUrl:       "https://via.placeholder.com/150",
		Email:        email,
		HashPassword: hash,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || VerifyPassword(password, user.HashPassword) != nil {
		return "", "", errors.New("invalid credentials")
	}

	access, err := s.generateAccessToken(user.ID.String())
	if err != nil {
		return "", "", err
	}
	refresh, err := s.generateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", err
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refresh,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return "", "", err
	}

	return access, refresh, nil
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

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	userID := claims["sub"].(string)

	session, err := s.userRepo.GetSessionByUserID(ctx, userID)
	if err != nil || session.RefreshToken != refreshToken || time.Now().After(session.ExpiresAt) {
		return "", errors.New("invalid or expired session")
	}
	return s.generateAccessToken(userID)
}

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

func GenerateRandomSalt(length int) (string, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %v", err)
	}
	return base64.RawStdEncoding.EncodeToString(salt), nil
}

func HashPassword(password string) (string, error) {
	salt, err := GenerateRandomSalt(saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), []byte(salt), iterations, memory, uint8(parallelism), keyLength)

	saltEncoded := base64.RawStdEncoding.EncodeToString([]byte(salt))
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s$%s", saltEncoded, hashEncoded), nil
}

func VerifyPassword(password string, hashedPassword string) error {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 2 {
		return errors.New("invalid hashed password format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return errors.New("failed to decode salt")
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("failed to decode stored hash")
	}

	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), keyLength)

	if !bytes.Equal(computedHash, storedHash) {
		return errors.New("invalid password")
	}

	return nil
}
