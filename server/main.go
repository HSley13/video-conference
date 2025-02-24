package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"video-conference/internal/config"
	"video-conference/internal/database"
	"video-conference/internal/repository"
	"video-conference/internal/service"
	"video-conference/internal/transport/http"
	"video-conference/internal/transport/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	redisClient, err := repository.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	wsRepo := repository.NewWSRepository(redisClient)

	authSvc := service.NewAuthService(userRepo)
	roomSvc := service.NewRoomService(roomRepo)
	wsSvc := service.NewWebSocketService(wsRepo, roomRepo, userRepo)

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Post("/api/auth/register", http.RegisterHandler(authSvc))
	app.Post("/api/auth/login", http.LoginHandler(authSvc))

	app.Use(middleware.JWTAuth(cfg.JWTSecret))
	app.Post("/api/rooms", http.CreateRoomHandler(roomSvc))
	app.Post("/api/rooms/:id/join", http.JoinRoomHandler(roomSvc))

	app.Use("/ws/:roomID", middleware.WSUpgrade)
	app.Get("/ws/:roomID", websocket.New(http.WebSocketHandler(wsSvc)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctxTimeout); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped gracefully")
}
