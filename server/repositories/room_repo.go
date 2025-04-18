package repositories

import (
	"context"
	"encoding/json"
	"video-conference/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type RoomRepository struct {
	redis *redis.Client
	db    *gorm.DB
}

type RoomSubscription struct {
	PubSub  *redis.PubSub
	Channel <-chan *redis.Message
}

func NewRoomRepository(redis *redis.Client, db *gorm.DB) *RoomRepository {
	return &RoomRepository{redis: redis, db: db}
}

func (r *RoomRepository) AddParticipant(ctx context.Context, roomID string, userID string) error {
	return r.redis.SAdd(ctx, "room:"+roomID+":participants", userID).Err()
}

func (r *RoomRepository) RemoveParticipant(ctx context.Context, roomID string, userID string) error {
	return r.redis.SRem(ctx, "room:"+roomID+":participants", userID).Err()
}

func (r *RoomRepository) GetParticipants(ctx context.Context, roomID string) ([]string, error) {
	return r.redis.SMembers(ctx, "room:"+roomID+":participants").Result()
}

func (r *RoomRepository) PublishMessage(ctx context.Context, roomID string, message interface{}) error {
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.redis.Publish(ctx, "room:"+roomID, jsonMsg).Err()

}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	room := models.Room{}
	result := r.db.WithContext(ctx).First(&room, "id = ?", roomID)
	return &room, result.Error
}

func (r *RoomRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *RoomRepository) SubscribeToRoom(ctx context.Context, roomID string) (*RoomSubscription, error) {
	pubsub := r.redis.Subscribe(ctx, "room:"+roomID)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		pubsub.Close()
		return nil, err
	}

	return &RoomSubscription{
		PubSub:  pubsub,
		Channel: pubsub.Channel(),
	}, nil
}

func (r *RoomRepository) UnsubscribeFromRoom(ctx context.Context, sub *RoomSubscription) error {
	if sub == nil || sub.PubSub == nil {
		return nil
	}
	return sub.PubSub.Close()
}
