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
	ws *services.WebSocketService,
	room *repositories.RoomRepository,
	user *repositories.UserRepository,
) *Server {
	app := fiber.New(fiber.Config{ErrorHandler: utils.GlobalErrorHandler})

	return &Server{
		app:      app,
		cfg:      cfg,
		authSvc:  auth,
		wsSvc:    ws,
		roomRepo: room,
		userRepo: user,
	}
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
	auth.Post("/refresh", s.handleRefreshToken)

	rooms := api.Group("/room", s.authRequired)
	rooms.Post("/", s.handleCreateRoom)
	rooms.Post("/join/:id", s.handleJoinRoom)

	ws := api.Group("/ws", s.authenticateWS)
	ws.Get("/:roomID/:userID", websocket.New(s.handleWebSocket))

	api.Get("/health", s.healthCheck)
}

func (s *Server) handleRegister(c *fiber.Ctx) error {
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}
	u, err := s.authSvc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusConflict, "registration failed")
	}
	log.Printf("[AUTH] register → %s", u.Email)
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(u))
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}
	acc, ref, err := s.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "invalid credentials")
	}
	return c.JSON(utils.SuccessResponse(fiber.Map{
		"access_token":  acc,
		"refresh_token": ref,
	}))
}

func (s *Server) handleRefreshToken(c *fiber.Ctx) error {
	var req struct{ RefreshToken string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}
	acc, err := s.authSvc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "refresh failed")
	}
	return c.JSON(utils.SuccessResponse(fiber.Map{"access_token": acc}))
}

func (s *Server) handleCreateRoom(c *fiber.Ctx) error {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}

	ownerID := c.Locals("jwtUserID").(uuid.UUID)
	room := models.Room{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Title:       body.Title,
		Description: body.Description,
		IsActive:    true,
	}
	if err := s.roomRepo.CreateRoom(c.Context(), &room); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "create failed")
	}
	return c.JSON(utils.SuccessResponse(fiber.Map{"room": room}))
}

func (s *Server) handleJoinRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Locals("jwtUserID").(uuid.UUID)

	room, err := s.roomRepo.GetRoom(c.Context(), roomID)
	if err != nil || !room.IsActive {
		return utils.RespondWithError(c, fiber.StatusNotFound, "room not found")
	}
	if err := s.roomRepo.AddParticipant(c.Context(), roomID, userID.String()); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "join failed")
	}

	ids, _ := s.roomRepo.GetParticipants(c.Context(), roomID)
	list := make([]fiber.Map, 0, len(ids))
	for _, id := range ids {
		if u, _ := s.userRepo.GetUserByID(c.Context(), id); u != nil {
			list = append(list, fiber.Map{"id": u.ID, "name": u.Name, "imgUrl": u.ImgUrl})
		}
	}

	return c.JSON(utils.SuccessResponse(fiber.Map{
		"room_id":      roomID,
		"participants": list,
	}))
}

func (s *Server) authenticateWS(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	token := extractToken(c)
	claims, err := s.authSvc.ValidateToken(token)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	sub := claims["sub"].(string)
	c.Locals("jwtUserID", sub)
	c.Locals("ctx", c.Context())
	return c.Next()
}

func (s *Server) handleWebSocket(c *websocket.Conn) {
	ctx := c.Locals("ctx").(context.Context)
	userID := c.Locals("jwtUserID").(string)
	roomID := c.Params("roomID")

	if ok, _ := s.roomRepo.GetRoom(ctx, roomID); ok == nil {
		_ = c.WriteJSON(fiber.Map{"error": "unknown room"})
		_ = c.Close()
		return
	}

	ids, _ := s.roomRepo.GetParticipants(ctx, roomID)
	list := make([]fiber.Map, 0, len(ids))
	for _, id := range ids {
		if u, _ := s.userRepo.GetUserByID(ctx, id); u != nil {
			list = append(list, fiber.Map{"id": u.ID, "name": u.Name, "imgUrl": u.ImgUrl})
		}
	}
	_ = c.WriteJSON(fiber.Map{"type": "users-list", "users": list})

	s.wsSvc.HandleConnection(ctx, c, roomID, userID)
}

func (s *Server) authRequired(c *fiber.Ctx) error {
	token := extractToken(c)
	claims, err := s.authSvc.ValidateToken(token)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	c.Locals("jwtUserID", uuid.MustParse(claims["sub"].(string)))
	return c.Next()
}

func extractToken(c *fiber.Ctx) string {
	t := c.Get("Authorization")
	if t == "" {
		t = c.Cookies("access_token")
	}
	if t == "" {
		t = c.Query("access_token")
	}
	if strings.HasPrefix(strings.ToLower(t), "bearer ") {
		t = strings.TrimSpace(t[7:])
	}
	return t
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "healthy", "version": "1.0.6"})
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
