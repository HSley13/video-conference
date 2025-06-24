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
	app      *fiber.App
	cfg      *config.Config
	authSvc  *services.AuthService
	wsSvc    *services.WebSocketService
	roomRepo *repositories.RoomRepository
	userRepo *repositories.UserRepository
}

func New(cfg *config.Config, auth *services.AuthService,
	ws *services.WebSocketService, roomRepo *repositories.RoomRepository,
	userRepo *repositories.UserRepository) *Server {

	app := fiber.New(fiber.Config{ErrorHandler: utils.GlobalErrorHandler})
	return &Server{app: app, cfg: cfg, authSvc: auth, wsSvc: ws, roomRepo: roomRepo, userRepo: userRepo}
}

func (s *Server) SetupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(s.cfg.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	s.app.Use(logger.New(logger.Config{
		TimeFormat: time.RFC3339,
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
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
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid body")
	}
	user, err := s.authSvc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusConflict, "Registration failed")
	}
	log.Printf("REGISTER ok -> %s", user.Email)
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(user))
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid body")
	}
	access, refresh, err := s.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}
	log.Printf("LOGIN ok -> %s", req.Email)
	return c.JSON(utils.SuccessResponse(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
	}))
}

func (s *Server) handleRefreshToken(c *fiber.Ctx) error {
	var req struct{ RefreshToken string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "Invalid body")
	}
	newAcc, err := s.authSvc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "Token refresh failed")
	}
	return c.JSON(utils.SuccessResponse(fiber.Map{"access_token": newAcc}))
}

func (s *Server) authenticateWS(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	token := c.Get("Authorization")
	if token == "" {
		token = c.Cookies("access_token")
	}
	if token == "" {
		token = c.Query("access_token")
	}
	if token == "" {
		log.Printf("[AUTH] missing token -> 401")
		return fiber.ErrUnauthorized
	}
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	claims, err := s.authSvc.ValidateToken(token)
	if err != nil {
		log.Printf("[AUTH] invalid token -> 401 : %v", err)
		return fiber.ErrUnauthorized
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return fiber.ErrUnauthorized
	}

	c.Locals("ctx", c.Context())
	c.Locals("jwtUserID", userID)
	return c.Next()
}

func (s *Server) handleWebSocket(c *websocket.Conn) {
	ctx := c.Locals("ctx").(context.Context)
	userID := c.Locals("jwtUserID").(string)

	paramUserID, _ := uuid.Parse(c.Params("userID"))
	roomID, _ := uuid.Parse(c.Params("roomID"))

	if paramUserID.String() != userID {
		_ = c.WriteJSON(fiber.Map{"error": "user id mismatch"})
		c.Close()
		return
	}

	log.Printf("[WS] handshake OK user=%s room=%s", userID, roomID)

	if _, err := s.roomRepo.GetRoom(ctx, roomID.String()); err != nil {
		_ = s.roomRepo.CreateRoom(ctx, &models.Room{
			ID:              roomID,
			Name:            "Room " + roomID.String()[:8],
			OwnerID:         paramUserID,
			MaxParticipants: 10,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	participants, _ := s.roomRepo.GetParticipants(ctx, roomID.String())
	var list []fiber.Map
	for _, pid := range participants {
		if u, _ := s.userRepo.GetUserByID(ctx, pid); u != nil {
			list = append(list, fiber.Map{"id": u.ID, "name": u.Name, "imgUrl": u.ImgUrl})
		}
	}
	_ = c.WriteJSON(fiber.Map{"type": "users-list", "users": list})

	s.wsSvc.HandleConnection(ctx, c, roomID.String(), paramUserID.String())
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "healthy", "version": "1.0.3"})
}

func (s *Server) Start() {
	s.SetupMiddleware()
	s.SetupRoutes()

	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		log.Println("SIGTERM received – shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(ctx)
		close(done)
	}()

	log.Printf("HTTP/WebSocket listening on :%s", s.cfg.Port)
	if err := s.app.Listen(":" + s.cfg.Port); err != nil {
		log.Fatalf("fiber.Listen failed: %v", err)
	}
	<-done
	log.Println("Server exited")
}
