package repositories

import (
	"context"
	"encoding/json"
	"fmt"

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

func NewRoomRepository(rdb *redis.Client, db *gorm.DB) *RoomRepository {
	return &RoomRepository{redis: rdb, db: db}
}

func participantsKey(roomID string) string { return "room:" + roomID + ":participants" }
func channelKey(roomID string) string      { return "room:" + roomID }

func (r *RoomRepository) AddParticipant(ctx context.Context, roomID, userID string) error {
	return r.redis.SAdd(ctx, participantsKey(roomID), userID).Err()
}

func (r *RoomRepository) RemoveParticipant(ctx context.Context, roomID, userID string) error {
	return r.redis.SRem(ctx, participantsKey(roomID), userID).Err()
}

func (r *RoomRepository) GetParticipants(ctx context.Context, roomID string) ([]string, error) {
	return r.redis.SMembers(ctx, participantsKey(roomID)).Result()
}

func (r *RoomRepository) PublishMessage(ctx context.Context, roomID string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return r.redis.Publish(ctx, channelKey(roomID), payload).Err()
}

func (r *RoomRepository) SubscribeToRoom(
	ctx context.Context,
	roomID string,
) (*RoomSubscription, error) {
	ps := r.redis.Subscribe(ctx, channelKey(roomID))

	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	return &RoomSubscription{
		PubSub:  ps,
		Channel: ps.Channel(),
	}, nil
}

func (r *RoomRepository) UnsubscribeFromRoom(_ context.Context, sub *RoomSubscription) error {
	if sub == nil || sub.PubSub == nil {
		return nil
	}
	return sub.PubSub.Close()
}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	var room models.Room
	if err := r.db.WithContext(ctx).First(&room, "id = ?", roomID).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}
