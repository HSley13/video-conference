package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"video-conference/config"
	"video-conference/models"
	"video-conference/repositories"
	"video-conference/seed"
	"video-conference/server"
	"video-conference/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	seed.Seed(db)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Room{},
		&models.Participant{},
	); err != nil {
		log.Fatal("Database migration failed:", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}
	redisClient := redis.NewClient(redisOpts)

	userRepo := repositories.NewUserRepository(db)
	roomRepo := repositories.NewRoomRepository(redisClient, db)

	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret)
	websocketSvc := services.NewWebSocketService(roomRepo, userRepo, cfg.WebRTCIceServers, cfg.MaxConnections)

	srv := server.New(cfg, authSvc, websocketSvc, roomRepo)
	srv.Start()
}
