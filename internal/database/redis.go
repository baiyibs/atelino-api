package database

import (
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	host := os.Getenv("REDIS_HOST")
	portStr := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	addr := fmt.Sprintf("%s:%s", host, portStr)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 5,
	})
	log.Println("Redis 初始化成功")
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
