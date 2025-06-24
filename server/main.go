package main

import (
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
	/* -------------------- env / config -------------------- */
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
	cfg := config.Load()

	/* -------------------- database ------------------------ */
	db := db_aws.InitDb(cfg.PostgresDSN)

	/* -------------------- redis --------------------------- */
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	/* ---------------- repositories / services ------------- */
	userRepo := repositories.NewUserRepository(db)
	roomRepo := repositories.NewRoomRepository(redisClient, db)

	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret)
	wsSvc := services.NewWebSocketService(
		roomRepo,
		userRepo,
		cfg.WebRTCIceServers,
		cfg.MaxConnections,
	)

	/* -------------------- server -------------------------- */
	srv := server.New(cfg, authSvc, wsSvc, roomRepo, userRepo)
	srv.Start()
}
