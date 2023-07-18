package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var redisRing *redis.Ring

func InitializeRedisRing(nodes map[string]string) {
	redisRing = redis.NewRing(&redis.RingOptions{
		Addrs: nodes,
	})
}

func GetRedisRing() *redis.Ring {
	return redisRing
}

func SetRedisRing(ring *redis.Ring) {
	redisRing = ring
}

// StoreDataToRedis use redis ringOptions to store data, the ringOptions uses Consistent Hashing
func StoreDataToRedis(key string, value []byte) error {
	redisRing := GetRedisRing()

	err := redisRing.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	fmt.Printf("Stored key %s with value %s to the Ring\n", key, value)
	return nil
}
