package tfa

import (
	"flag"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	redisAddr     string
	redisPassword string
	redisTTL      time.Duration
)

func init() {
	flag.StringVar(&redisAddr, "redis-addr", "localhost:6379", "hostname:port for redis to listen to")
	flag.StringVar(&redisPassword, "redis-password", "", "redis password")
	flag.DurationVar(&redisTTL, "redis-ttl", time.Hour, "session expiring duration")
}

// RedisClient ...
type RedisClient struct {
	*redis.Client
}

// NewRedisClient return the new redisClient instance
func NewRedisClient() *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	return &RedisClient{
		Client: client,
	}
}

// GetString ...
func (r *RedisClient) GetString(key string) (string, error) {
	return r.Get(key).Result()
}

// SetString ...
func (r *RedisClient) SetString(key, value string) error {
	return r.Set(key, value, redisTTL).Err()
}

func (r *RedisClient) DeleteString(key string) error {
	return r.Del(key).Err()
}

func (r *RedisClient) GetToken(s string) string {
	elements := strings.Split(s, "|")
	if len(elements) > 0 {
		return elements[0]
	}

	return ""
}
