package server

import (
	"context"
	"time"

	"video-conference/models"
	"video-conference/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

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

	return utils.SuccessResponse(c, fiber.Map{"id": room.ID})
}

func (s *Server) handleUserInfo(c *fiber.Ctx) error {
	uid := c.Params("id")
	if u, _ := s.userRepo.GetUserByID(c.Context(), uid); u != nil {
		return utils.SuccessResponse(c, fiber.Map{"id": u.ID, "userName": u.UserName, "imgUrl": u.ImgUrl})
	}
	return utils.RespondWithError(c, fiber.StatusNotFound, "user not found")
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
			list = append(list, fiber.Map{"userID": u.ID, "userName": u.UserName, "imgUrl": u.ImgUrl})
		}
	}
	_ = conn.WriteJSON(fiber.Map{"type": "users-list", "users": list})

	s.wsSvc.HandleConnection(ctx, conn, roomID, uid)
}
