package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"video-conference/config"
	"video-conference/repositories"
	"video-conference/services"
	"video-conference/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	app      *fiber.App
	cfg      *config.Config
	authSvc  *services.AuthService
	wsSvc    *services.WebSocketService
	roomRepo *repositories.RoomRepository
	userRepo *repositories.UserRepository
}

func New(cfg *config.Config, auth *services.AuthService,
	ws *services.WebSocketService,
	room *repositories.RoomRepository,
	user *repositories.UserRepository,
) *Server {
	app := fiber.New(fiber.Config{ErrorHandler: utils.GlobalErrorHandler})
	return &Server{app, cfg, auth, ws, room, user}
}

func (s *Server) SetupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(logger.New(logger.Config{
		TimeFormat: time.RFC3339,
		Format:     "[${time}] ${status} – ${latency} ${method} ${path}\n",
	}))
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(s.cfg.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
}

func (s *Server) SetupRoutes() {
	api := s.app.Group("/video-conference")

	auth := api.Group("/auth")
	auth.Post("/register", s.handleRegister)
	auth.Post("/login", s.handleLogin)
	auth.Post("/refresh", s.handleRefresh)

	user := api.Group("/user", s.authSvc.AuthRequired)
	user.Get("/userInfo/:id", s.handleUserInfo)
	// user.Post("/updataUserInfo", s.handleUpdateUserInfo)

	room := api.Group("/room", s.authSvc.AuthRequired)
	room.Post("/", s.handleCreateRoom)
	room.Post("/join/:id", s.handleJoinRoom)

	ws := api.Group("/ws", s.authSvc.AuthenticateWS)
	ws.Get("/:roomID", websocket.New(s.handleWebSocket))

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "version": "1.2.0"})
	})
}

func (s *Server) Start() {
	s.SetupMiddleware()
	s.SetupRoutes()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("graceful shutdown …")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(ctx)
	}()

	log.Printf("HTTP / WS listening on :%s", s.cfg.Port)
	if err := s.app.Listen(":" + s.cfg.Port); err != nil {
		log.Fatalf("fiber.Listen: %v", err)
	}
}
