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

func New(
	cfg *config.Config,
	auth *services.AuthService,
	ws *services.WebSocketService,
	room *repositories.RoomRepository,
	user *repositories.UserRepository,
) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: utils.GlobalErrorHandler,
	})

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

	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(s.cfg.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	s.app.Use(logger.New(logger.Config{
		TimeFormat: time.RFC3339,
		Format:     "[${time}] ${status} – ${latency} ${method} ${path}\n",
	}))
}

func (s *Server) SetupRoutes() {
	api := s.app.Group("/video-conference")

	auth := api.Group("/auth")
	auth.Post("/register", s.handleRegister)
	auth.Post("/login", s.handleLogin)
	auth.Post("/refresh", s.handleRefreshToken)

	rooms := api.Group("/rooms")
	rooms.Post("/", s.handleCreateRoom)
	// rooms.Get("/:id", s.handleGetRoom)
	// rooms.Put("/:id", s.handleUpdateRoom)
	// rooms.Delete("/:id", s.handleDeleteRoom)
	//
	// users := api.Group("/users")
	// users.Get("/", s.handleGetUsers)
	// users.Get("/:id", s.handleGetUser)
	// users.Put("/:id", s.handleUpdateUser)
	// users.Delete("/:id", s.handleDeleteUser)

	ws := api.Group("/ws", s.authenticateWS)
	ws.Get("/:roomID/:userID", websocket.New(s.handleWebSocket))

	api.Get("/health", s.healthCheck)
}

func (s *Server) handleRegister(c *fiber.Ctx) error {
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}

	user, err := s.authSvc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusConflict, "registration failed")
	}

	log.Printf("[AUTH] register → %s", user.Email)
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(user))
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req struct{ Email, Password string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}

	access, refresh, err := s.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	log.Printf("[AUTH] login → %s", req.Email)
	return c.JSON(utils.SuccessResponse(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
	}))
}

func (s *Server) handleRefreshToken(c *fiber.Ctx) error {
	var req struct{ RefreshToken string }
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}

	access, err := s.authSvc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "token refresh failed")
	}

	return c.JSON(utils.SuccessResponse(fiber.Map{"access_token": access}))
}

func (s Server) handleCreateRoom(c *fiber.Ctx) error {
	var Body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&Body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid body")
	}

	OwnerID := uuid.MustParse(c.Locals("jwtUserID").(string))
	if OwnerID == uuid.Nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "invalid jwtUserID")
	}

	room := models.Room{
		OwnerID:     OwnerID,
		Title:       Body.Title,
		Description: Body.Description,
		IsActive:    true,
	}

	if err := s.roomRepo.CreateRoom(c.Context(), &room).Error; err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "failed to create room")
	}

	return c.JSON(utils.SuccessResponse(fiber.Map{"roomID": room.ID}))
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
		log.Print("[WS-AUTH] missing token")
		return fiber.ErrUnauthorized
	}

	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	claims, err := s.authSvc.ValidateToken(token)
	if err != nil {
		log.Printf("[WS-AUTH] invalid token: %v", err)
		return fiber.ErrUnauthorized
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return fiber.ErrUnauthorized
	}

	c.Locals("ctx", c.Context())
	c.Locals("jwtUserID", sub)
	return c.Next()
}

func (s *Server) handleWebSocket(c *websocket.Conn) {
	ctx := c.Locals("ctx").(context.Context)
	jwtUser := c.Locals("jwtUserID").(string)
	paramUID := c.Params("userID")
	paramRID := c.Params("roomID")

	userUUID, err1 := uuid.Parse(paramUID)
	roomUUID, err2 := uuid.Parse(paramRID)
	if err1 != nil || err2 != nil {
		_ = c.WriteJSON(fiber.Map{"error": "invalid roomId or userId"})
		_ = c.Close()
		return
	}

	if userUUID.String() != jwtUser {
		_ = c.WriteJSON(fiber.Map{"error": "user id mismatch"})
		_ = c.Close()
		return
	}

	log.Printf("[WS] handshake OK user=%s room=%s", jwtUser, roomUUID)

	if _, err := s.roomRepo.GetRoom(ctx, roomUUID.String()); err != nil {
		_ = s.roomRepo.CreateRoom(ctx, &models.Room{
			ID:              roomUUID,
			Name:            "Room " + roomUUID.String()[:8],
			OwnerID:         userUUID,
			MaxParticipants: 10,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	ids, _ := s.roomRepo.GetParticipants(ctx, roomUUID.String())
	list := make([]fiber.Map, 0, len(ids))
	for _, id := range ids {
		if u, _ := s.userRepo.GetUserByID(ctx, id); u != nil {
			list = append(list, fiber.Map{"id": u.ID, "name": u.Name, "imgUrl": u.ImgUrl})
		}
	}
	_ = c.WriteJSON(fiber.Map{"type": "users-list", "users": list})

	s.wsSvc.HandleConnection(ctx, c, roomUUID.String(), userUUID.String())
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "healthy", "version": "1.0.5"})
}

func (s *Server) Start() {
	s.SetupMiddleware()
	s.SetupRoutes()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("graceful shutdown…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(ctx)
	}()

	log.Printf("HTTP/WS listening on :%s", s.cfg.Port)
	if err := s.app.Listen(":" + s.cfg.Port); err != nil {
		log.Fatalf("fiber.Listen: %v", err)
	}
}
