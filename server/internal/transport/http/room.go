package http

import (
	"github.com/gofiber/fiber/v2"

	"video-conference/internal/service"
	"video-conference/internal/utils"
)

func CreateRoomHandler(svc *service.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		type CreateRoomRequest struct {
			Name        string `json:"name" validate:"required,min=3"`
			Description string `json:"description"`
		}

		var req CreateRoomRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid request"))
		}

		if errors := utils.ValidateStruct(req); len(errors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ValidationErrorResponse(errors))
		}

		room, err := svc.CreateRoom(c.Context(), userID, req.Name, req.Description)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(err.Error()))
		}

		return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(room))
	}
}

func JoinRoomHandler(svc *service.RoomService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid room ID"))
		}

		participant, err := svc.JoinRoom(c.Context(), userID, uint(roomID))
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse(err.Error()))
		}

		return c.JSON(utils.SuccessResponse(participant))
	}
}
