package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-conference/config"
	"video-conference/models"
	"video-conference/repositories"
	"video-conference/services"
	"video-conference/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type Server struct {
	app          *fiber.App
	cfg          *config.Config
	authSvc      *services.AuthService
	websocketSvc *services.WebSocketService
	roomRepo     *repositories.RoomRepository
}

func New(cfg *config.Config, authSvc *services.AuthService, websocketSvc *services.WebSocketService, roomRepo *repositories.RoomRepository) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: utils.GlobalErrorHandler,
	})
	return &Server{
		app:          app,
		cfg:          cfg,
		authSvc:      authSvc,
		websocketSvc: websocketSvc,
		roomRepo:     roomRepo,
	}
}

func (s *Server) SetupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	s.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
}

func (s *Server) SetupRoutes() {
	api := s.app.Group("/video-conference")

	auth := api.Group("/auth")
	auth.Post("/register", s.handleRegister)
	auth.Post("/login", s.handleLogin)
	auth.Post("/refresh", s.handleRefreshToken)

	ws := api.Group("/ws")
	ws.Use(s.authenticateWS)
	ws.Get("/:roomID/:userID", websocket.New(s.handleWebSocket))

	api.Get("/health", s.healthCheck)
}

func (s *Server) handleRegister(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := s.authSvc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusConflict, "Registration failed")
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(user))
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	accessToken, refreshToken, err := s.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	return c.JSON(utils.SuccessResponse(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}))
}

func (s *Server) handleRefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	newAccessToken, err := s.authSvc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "Token refresh failed")
	}

	return c.JSON(utils.SuccessResponse(fiber.Map{
		"access_token": newAccessToken,
	}))
}

func (s *Server) authenticateWS(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		// 	token := c.Get("Authorization", c.Cookies("access_token"))
		// 	if token == "" {
		// 		return fiber.ErrUnauthorized
		// 	}
		//
		// 	claims, err := s.authSvc.ValidateToken(token)
		// 	if err != nil {
		// 		return fiber.ErrUnauthorized
		// 	}
		//
		// 	userID, ok := claims["sub"].(string)
		// 	if !ok {
		// 		return fiber.ErrUnauthorized
		// 	}

		c.Locals("ctx", c.Context())
		// c.Locals("userID", userID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (s *Server) handleWebSocket(c *websocket.Conn) {
	ctx, ok := c.Locals("ctx").(context.Context)
	if !ok {
		c.WriteJSON(fiber.Map{"error": "Internal server error"})
		c.Close()
		return
	}

	userIDStr := c.Params("userID")
	roomIDStr := c.Params("roomID")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid userID %s: %v\n", userIDStr, err)
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		log.Printf("Invalid roomID %s, generating a new UUID: %v\n", roomIDStr, err)
		roomID = uuid.New()
	}

	_, err = s.roomRepo.GetRoom(ctx, roomID.String())
	if err != nil {
		log.Printf("Creating room %s: %v\n", roomID, err)
		newRoom := &models.Room{
			ID:              roomID,
			Name:            "New Room",
			OwnerID:         userID,
			IsActive:        true,
			MaxParticipants: 10,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err = s.roomRepo.CreateRoom(ctx, newRoom)
		if err != nil {
			log.Printf("Error creating room %s: %v\n", roomID, err)
			c.WriteJSON(fiber.Map{"error": "Internal server error"})
			c.Close()
			return
		}
	}

	allowed, err := s.websocketSvc.CanJoinRoom(ctx, roomID.String())
	if err != nil || !allowed {
		c.WriteJSON(fiber.Map{"error": "Cannot join room"})
		c.Close()
		return
	}

	log.Printf("User %s joined room %s\n", userID, roomID)
	s.websocketSvc.HandleConnection(ctx, c, roomID.String(), userID.String())
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

func (s *Server) Start() {
	s.SetupMiddleware()
	s.SetupRoutes()

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Starting server on :%s", s.cfg.Port)
	if err := s.app.Listen(":" + s.cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	<-idleConnsClosed
	log.Println("Server stopped gracefully")
}
