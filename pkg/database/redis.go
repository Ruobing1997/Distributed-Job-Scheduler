package database

import "github.com/redis/go-redis/v9"

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
