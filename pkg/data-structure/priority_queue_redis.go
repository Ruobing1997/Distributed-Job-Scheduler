package data_structure_redis

import (
	"context"
	"encoding/json"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var client *redis.Client
var Qlen int64

func Init() {
	client = redis.NewClient(&redis.Options{
		Addr:     REDISPQ_Addr,
		Password: REDISPQ_Password,
		DB:       REDISPQ_DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("redis connection failed: %v", err)
	}
	Qlen = 0
}

func AddJob(e *constants.TaskCache) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("marshal task cache failed: %v", err)
	}
	var z = redis.Z{
		Score:  float64(e.ExecutionTime.Unix()),
		Member: data,
	}
	client.ZAdd(context.Background(), REDIS_PQ_KEY, z)
	Qlen++
}

func PopNextJob() *constants.TaskCache {
	result, err := client.ZPopMin(context.Background(), REDIS_PQ_KEY).Result()
	if err != nil {
		log.Printf("get next job failed: %v", err)
	}

	// Decode the job name and execution time from the JSON string
	var e constants.TaskCache
	if len(result) > 0 {
		Qlen--
		if err := json.Unmarshal([]byte(result[0].Member.(string)), &e); err != nil {
			log.Printf("unmarshal task cache failed: %v", err)
		}
	} else {
		log.Printf("no task cache in redis")
	}
	return &e
}

func CheckTasksInDuration(curTask *constants.TaskCache, duration time.Duration) bool {
	now := time.Now()
	return curTask.ExecutionTime.After(now) && curTask.ExecutionTime.Before(now.Add(duration))
}

func GetQLength() int64 {
	return Qlen
}

func GetNextJob() *constants.TaskCache {
	result, err := client.ZRange(context.Background(), REDIS_PQ_KEY, 0, 0).Result()
	if err != nil {
		log.Printf("get next job failed: %v", err)
	}

	// Decode the job name and execution time from the JSON string
	var e constants.TaskCache
	if len(result) > 0 {
		if err := json.Unmarshal([]byte(result[0]), &e); err != nil {
			log.Printf("unmarshal task cache failed: %v", err)
		}
	} else {
		log.Printf("no task cache in redis")
	}
	return &e
}
