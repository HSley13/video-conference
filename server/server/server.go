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

	user := api.Group("/user", s.authRequired)
	user.Get("/userInfo/:id", s.handleUserInfo)
	// user.Post("/updataUserInfo", s.handleUpdateUserInfo)

	room := api.Group("/room", s.authRequired)
	room.Post("/", s.handleCreateRoom)
	room.Post("/join/:id", s.handleJoinRoom)

	ws := api.Group("/ws", s.authenticateWS)
	ws.Get("/:roomID", websocket.New(s.handleWebSocket))

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "version": "1.2.0"})
	})
}

func (s *Server) handleRegister(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "bad body")
	}

	acc, ref, uid, err := s.authSvc.Register(c.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusConflict, "registration failed")
	}

	s.authSvc.SetAuthCookies(c, acc, ref, uid)
	return utils.SuccessResponse(c, nil)
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "bad body")
	}

	acc, ref, uid, err := s.authSvc.Login(c.Context(), body.Email, body.Password)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	s.authSvc.SetAuthCookies(c, acc, ref, uid)

	return utils.SuccessResponse(c, nil)
}

func (s *Server) handleRefresh(c *fiber.Ctx) error {
	ref := c.Cookies("refresh_token")
	if ref == "" {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "no refresh token")
	}

	newAcc, err := s.authSvc.RefreshToken(c.Context(), ref)
	if err != nil {
		return utils.RespondWithError(c, fiber.StatusUnauthorized, "refresh failed")
	}

	s.authSvc.SetAuthCookies(c, newAcc, "", "")
	return utils.SuccessResponse(c, nil)
}

func (s *Server) handleCreateRoom(c *fiber.Ctx) error {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, "bad body")
	}

	owner := uuid.MustParse(c.Cookies("videoConferenceUserId"))
	room := models.Room{
		ID:              uuid.New(),
		OwnerID:         owner,
		Title:           body.Title,
		Description:     body.Description,
		MaxParticipants: 10,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.roomRepo.CreateRoom(c.Context(), &room); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "create failed")
	}
	if err := s.roomRepo.AddParticipant(c.Context(), room.ID.String(), owner.String()); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "join failed")
	}

	return utils.SuccessResponse(c, fiber.Map{"id": room.ID})
}

func (s *Server) handleJoinRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	user := uuid.MustParse(c.Cookies("videoConferenceUserId"))

	room, err := s.roomRepo.GetRoom(c.Context(), roomID)
	if err != nil || !room.IsActive {
		return utils.RespondWithError(c, fiber.StatusNotFound, "room not found")
	}
	if err := s.roomRepo.AddParticipant(c.Context(), roomID, user.String()); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "join failed")
	}

	return utils.SuccessResponse(c, fiber.Map{"id": roomID})
}

func (s *Server) authRequired(c *fiber.Ctx) error {
	claims, err := s.authSvc.ValidateToken(extractToken(c))
	if err != nil {
		return fiber.ErrUnauthorized
	}
	c.Locals("videoConferenceUserId", uuid.MustParse(claims["sub"].(string)))
	return c.Next()
}

func extractToken(c *fiber.Ctx) string {
	if t := c.Get("Authorization"); t != "" {
		if strings.HasPrefix(strings.ToLower(t), "bearer ") {
			return strings.TrimSpace(t[7:])
		}
		return t
	}
	if t := c.Cookies("access_token"); t != "" {
		return t
	}
	return c.Query("access_token")
}

func (s *Server) authenticateWS(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	claims, err := s.authSvc.ValidateToken(extractToken(c))
	if err != nil {
		return fiber.ErrUnauthorized
	}
	c.Locals("videoConferenceUserId", claims["sub"].(string))
	c.Locals("ctx", c.Context())
	return c.Next()
}

func (s *Server) handleWebSocket(conn *websocket.Conn) {
	ctx := conn.Locals("ctx").(context.Context)
	uid := conn.Locals("videoConferenceUserId").(string)
	roomID := conn.Params("roomID")

	if r, _ := s.roomRepo.GetRoom(ctx, roomID); r == nil {
		_ = conn.WriteJSON(fiber.Map{"error": "unknown room"})
		_ = conn.Close()
		return
	}

	ids, _ := s.roomRepo.GetParticipants(ctx, roomID)
	list := make([]fiber.Map, 0, len(ids))
	for _, id := range ids {
		if u, _ := s.userRepo.GetUserByID(ctx, id); u != nil {
			list = append(list, fiber.Map{"id": u.ID, "userName": u.UserName, "imgUrl": u.ImgUrl})
		}
	}
	_ = conn.WriteJSON(fiber.Map{"type": "users-list", "users": list})

	s.wsSvc.HandleConnection(ctx, conn, roomID, uid)
}

func (s *Server) handleUserInfo(c *fiber.Ctx) error {
	uid := c.Params("id")
	if u, _ := s.userRepo.GetUserByID(c.Context(), uid); u != nil {
		return utils.SuccessResponse(c, fiber.Map{"id": u.ID, "userName": u.UserName, "imgUrl": u.ImgUrl})
	}
	return utils.RespondWithError(c, fiber.StatusNotFound, "user not found")
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
