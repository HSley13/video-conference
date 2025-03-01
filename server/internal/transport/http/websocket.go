package http

import (
	"github.com/gofiber/websocket/v2"

	"github.com/gofiber/fiber/v2"

	"video-conference/internal/service"
	"video-conference/internal/utils"
)

func WebSocketHandler(svc *service.WebSocketService) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		userID, err := utils.GetUserIDFromContext(c.Context())
		if err != nil {
			c.WriteJSON(utils.ErrorResponse("Unauthorized"))
			c.Close()
			return
		}

		roomID := c.Params("roomID")
		svc.HandleConnection(c, userID, roomID)
	})
}
