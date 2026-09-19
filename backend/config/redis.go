package config

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) (*redis.Client, error) {
	log.Printf("[Redis] Connecting to %s...", cfg.RedisURL)

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		// Fallback to simple host:port
		opt = &redis.Options{
			Addr: cfg.RedisURL,
		}
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] WARNING: Ping failed (%v). Ensure Redis is running.", err)
		return client, err
	}

	log.Printf("[Redis] Successfully connected to Redis!")
	return client, nil
}
