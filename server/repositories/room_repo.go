package repositories

import (
	"context"
	"video-conference/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type RoomRepository struct {
	redis *redis.Client
	db    *gorm.DB
}

func NewRoomRepository(redis *redis.Client, db *gorm.DB) *RoomRepository {
	return &RoomRepository{redis: redis, db: db}
}

func (r *RoomRepository) AddParticipant(ctx context.Context, roomID, userID string) error {
	return r.redis.SAdd(ctx, "room:"+roomID+":participants", userID).Err()
}

func (r *RoomRepository) RemoveParticipant(ctx context.Context, roomID, userID string) error {
	return r.redis.SRem(ctx, "room:"+roomID+":participants", userID).Err()
}

func (r *RoomRepository) GetParticipants(ctx context.Context, roomID string) ([]string, error) {
	return r.redis.SMembers(ctx, "room:"+roomID+":participants").Result()
}

func (r *RoomRepository) PublishMessage(ctx context.Context, roomID string, message interface{}) error {
	return r.redis.Publish(ctx, "room:"+roomID, message).Err()
}

func (r *RoomRepository) SubscribeToRoom(ctx context.Context, roomID string) <-chan *redis.Message {
	pubsub := r.redis.Subscribe(ctx, "room:"+roomID)
	return pubsub.Channel()
}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	room := models.Room{}
	result := r.db.WithContext(ctx).First(&room, "id = ?", roomID)
	return &room, result.Error
}
