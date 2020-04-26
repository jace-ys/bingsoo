package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type Client interface {
	Transact(ctx context.Context, fn func(redis.Conn) error) error
	Close() error
}

type Config struct {
	Host string
}

type RedisClient struct {
	config *Config
	*redis.Pool
}

func NewRedisClient(host string) (*RedisClient, error) {
	r := RedisClient{
		config: &Config{
			Host: host,
		},
	}

	if err := r.init(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *RedisClient) init() error {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.config.Host)
		},
	}
	r.Pool = pool

	return nil
}

func (r *RedisClient) Transact(ctx context.Context, fn func(redis.Conn) error) error {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}
	defer conn.Close()

	if err := fn(conn); err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}

	return nil
}
