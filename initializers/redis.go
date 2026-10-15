package initializers

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR")

	if addr == "" {
		addr = "localhost:6379" //tells go applictaion where to find redis
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()

	if err != nil {
		log.Printf("failed to connect redis: %v", err)
		client.Close()
		return
	}

	RedisClient = client

	log.Println("connected to redis succesfully")
}

//redis commands accepts a contetxt, so the operation can be cancelled or have  a deadline
