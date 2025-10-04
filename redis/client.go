package redis

import (
	"context"
	"time"

	rdb "github.com/redis/go-redis/v9"
)

type Client struct {
	rdbc    *rdb.Client
	address string
}

func NewClient(address string) (*Client, error) {
	opt, err := rdb.ParseURL(address)
	if err != nil {
		return nil, err
	}

	rdb := rdb.NewClient(opt)

	cmd := rdb.Ping(context.Background())
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	client := &Client{
		rdbc:    rdb,
		address: address,
	}

	return client, nil
}

func (c *Client) Set(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	return c.rdbc.Set(ctx, key, val, expire).Err()
}

func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	cmd := c.rdbc.Get(ctx, key)
	result, err := cmd.Result()

	if result != "" {
		return result, nil
	}
	return nil, err
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdbc.Del(ctx, key).Err()
}

func (c *Client) Close() error {
	return c.rdbc.Close()
}

// GetRedisClient returns the underlying Redis client for advanced operations
func (c *Client) GetRedisClient() *rdb.Client {
	return c.rdbc
}
