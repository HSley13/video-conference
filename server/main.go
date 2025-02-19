package main

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Room struct {
	Peers map[*websocket.Conn]string
	Lock  sync.Mutex
}

var rooms = struct {
	sync.Map
}{}

func main() {
	app := fiber.New()

	app.Use("/ws/:roomID", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:roomID", websocket.New(func(c *websocket.Conn) {
		roomID := c.Params("roomID")
		peerID := c.Query("peerID")
		if peerID == "" {
			log.Println("Missing peerID in query parameters")
			return
		}

		room := getOrCreateRoom(roomID)

		room.Lock.Lock()
		room.Peers[c] = peerID
		room.Lock.Unlock()

		defer func() {
			room.Lock.Lock()
			delete(room.Peers, c)
			room.Lock.Unlock()
			c.Close()
		}()

		for {
			var msg map[string]interface{}
			if err := c.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					log.Printf("Connection closed unexpectedly: %v", err)
				}
				break
			}
			handleSignalingMessage(room, c, msg)
		}
	}))

	log.Fatal(app.Listen(":3000"))
}

func getOrCreateRoom(id string) *Room {
	actual, _ := rooms.LoadOrStore(id, &Room{
		Peers: make(map[*websocket.Conn]string),
	})
	return actual.(*Room)
}

func handleSignalingMessage(room *Room, sender *websocket.Conn, msg map[string]interface{}) {
	room.Lock.Lock()
	defer room.Lock.Unlock()

	targetPeerID, ok := msg["target"].(string)
	if !ok {
		for conn := range room.Peers {
			if conn != sender {
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("Write error: %v", err)
				}
			}
		}
		return
	}

	for conn, peerID := range room.Peers {
		if peerID == targetPeerID {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Write error: %v", err)
			}
			break
		}
	}
}
