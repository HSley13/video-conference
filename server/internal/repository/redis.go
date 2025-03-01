package repository

import (
	"context"
	"time"

	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"video-conference/internal/config"
)

type WSRepository struct {
	client *redis.Client
}

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

func (r *WSRepository) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, channel, payload).Err()
}

func (r *WSRepository) Subscribe(ctx context.Context, channel uuid.UUID) <-chan *redis.Message {
	pubsub := r.client.Subscribe(ctx, channel.String())
	return pubsub.Channel()
}

func (r *WSRepository) Unsubscribe(ctx context.Context, channel uuid.UUID) {
	r.client.Unlink(ctx, channel.String())
}

func (r *WSRepository) Close() error {
	return r.client.Close()
}
