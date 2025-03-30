package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"log"
	"video-conference/config"
	"video-conference/db_aws"
	"video-conference/repositories"
	"video-conference/server"
	"video-conference/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config.Load()

	db := db_aws.InitDb(cfg.PostgresDSN)

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
