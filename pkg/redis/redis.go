package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type ClientConfig struct {
	Host string
}

type Client struct {
	config *ClientConfig
	*redis.Pool
}

func NewClient(host string) (*Client, error) {
	r := Client{
		config: &ClientConfig{
			Host: host,
		},
	}

	if err := r.init(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *Client) init() error {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", c.config.Host)
		},
	}
	c.Pool = pool

	return nil
}

func (c *Client) Transact(ctx context.Context, fn func(redis.Conn) error) error {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}
	defer conn.Close()

	if err := fn(conn); err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}

	return nil
}
