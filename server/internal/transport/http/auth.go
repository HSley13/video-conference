package http

import (
	"github.com/gofiber/fiber/v2"

	"video-conference/internal/service"
	"video-conference/internal/utils"
)

func RegisterHandler(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type RegisterRequest struct {
			Username string `json:"username" validate:"required,min=3"`
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=8"`
		}

		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid request body"))
		}

		if errors := utils.ValidateStruct(req); len(errors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ValidationErrorResponse(errors))
		}

		user, err := svc.Register(c.Context(), req.Username, req.Email, req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(err.Error()))
		}

		return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(user))
	}
}

func LoginHandler(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type LoginRequest struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required"`
		}

		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid request body"))
		}

		if errors := utils.ValidateStruct(req); len(errors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ValidationErrorResponse(errors))
		}

		tokens, err := svc.Login(c.Context(), req.Email, req.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Invalid credentials"))
		}

		return c.JSON(utils.SuccessResponse(tokens))
	}
}
