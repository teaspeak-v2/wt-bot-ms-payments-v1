package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

type Client struct {
	client *redis.Client
}

func New(opts *redis.Options) *Client {
	if opts == nil || opts.Addr == "" {
		return nil
	}
	return &Client{client: redis.NewClient(opts)}
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func (c *Client) SetSubscription(ctx context.Context, sub *models.Subscription, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	b, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, subKey(sub.ID), b, ttl).Err()
}

func (c *Client) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil
	}
	val, err := c.client.Get(ctx, subKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var sub models.Subscription
	if err := json.Unmarshal(val, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (c *Client) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, subKey(id)).Err()
}

func subKey(id uuid.UUID) string { return fmt.Sprintf("payments:subscriptions:%s", id.String()) }
