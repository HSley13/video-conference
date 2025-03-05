package main

import (
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"video-conference/config"
	"video-conference/models"
	"video-conference/repositories"
	"video-conference/server"
	"video-conference/services"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

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
	websocketSvc := services.NewWebSocketService(roomRepo, cfg.WebRTCIceServers, cfg.MaxConnections)

	srv := server.New(cfg, authSvc, websocketSvc)
	srv.Start()
}
