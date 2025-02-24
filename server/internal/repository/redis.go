package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
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

func (r *WSRepository) SubscribeChannel(ctx context.Context, channel string) <-chan *redis.Message {
	pubsub := r.client.Subscribe(ctx, channel)
	return pubsub.Channel()
}

func (r *WSRepository) Close() error {
	return r.client.Close()
}
