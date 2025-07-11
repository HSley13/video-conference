package db_aws

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"video-conference/models"
	"video-conference/seed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"golang.org/x/crypto/argon2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

func NewS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}
	return s3.NewFromConfig(cfg), nil
}

func GetDataFromS3(ctx context.Context, s3Client *s3.Client, key string) (string, error) {
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		return "", fmt.Errorf("BUCKET_NAME environment variable is not set")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3Client.GetObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("Failed to get object: %v", err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read object body: %v", err)
	}

	return string(body), nil
}

func StoreDataToS3(ctx context.Context, s3Client *s3.Client, key string, file multipart.File) (string, error) {
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		return "", fmt.Errorf("BUCKET_NAME environment variable is not set")
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	_, err := s3Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("Failed to upload object: %v", err)
	}

	psClient := s3.NewPresignClient(s3Client)
	psInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	psURL, err := psClient.PresignGetObject(ctx, psInput, func(po *s3.PresignOptions) {
		po.Expires = 7 * 24 * time.Hour
	})
	if err != nil {
		return "", fmt.Errorf("Failed to generate presigned URL: %v", err)
	}

	return psURL.URL, nil
}

func DeleteDataFromS3(ctx context.Context, s3Client *s3.Client, key string) error {
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		return fmt.Errorf("BUCKET_NAME environment variable is not set")
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("Failed to delete object: %v", err)
	}

	return nil
}

func InitDb(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Fatalf("Failed to enable UUID extension: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Room{},
		&models.Participant{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	seed.Seed(db)

	return db
}

func GetOrCreateUser(db *gorm.DB, username string) models.User {
	var user models.User
	if err := db.Where("name = ?", username).First(&user).Error; err != nil {
		user = models.User{UserName: username}
		if createErr := db.Create(&user).Error; createErr != nil {
			log.Fatalf("Failed to create user: %v", createErr)
		}
	}
	return user
}

func cleanExpiredCodes(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Cleaning expired codes...")
		result := db.Where("expire_at < ?", time.Now()).Delete(&models.Code{})
		if result.Error != nil {
			log.Printf("Failed to delete expired codes: %v", result.Error)
		} else {
			log.Printf("Deleted %d expired codes", result.RowsAffected)
		}
	}
}
