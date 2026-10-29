package redis

import (
	"os"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
)

func NewClientFromEnv() *goredis.Client {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return goredis.NewClient(&goredis.Options{
		Addr:     envOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
