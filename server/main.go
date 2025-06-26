package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"

	"video-conference/config"
	"video-conference/db_aws"
	"video-conference/repositories"
	"video-conference/server"
	"video-conference/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found – falling back to shell env")
	}
	cfg := config.Load()

	db := db_aws.InitDb(cfg.PostgresDSN)

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis ParseURL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	ctx := context.Background()
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		log.Fatalf("redis FlushDB: %v", err)
	}
	log.Println("redis: database flushed")

	userRepo := repositories.NewUserRepository(db)
	roomRepo := repositories.NewRoomRepository(redisClient, db)

	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret)
	wsSvc := services.NewWebSocketService(
		roomRepo,
		userRepo,
		cfg.WebRTCIceServers,
		cfg.MaxConnections,
	)

	srv := server.New(cfg, authSvc, wsSvc, roomRepo, userRepo)
	srv.Start()
}
