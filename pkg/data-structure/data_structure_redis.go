package data_structure_redis

import (
	"context"
	"encoding/json"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

var client *redis.Client
var Qlen int64

func Init() *redis.Client {

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := os.Getenv("REDIS_DB")
	db := 0
	if redisDB != "" {
		db, _ = strconv.Atoi(redisDB)
	}

	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	return maximumConnectionRetry()
}

func maximumConnectionRetry() *redis.Client {
	for i := 0; i < 5; i++ {
		pong, err := client.Ping(context.Background()).Result()
		if err == nil {
			log.Printf("redis connection succeeded: %s", pong)
			Qlen = 0
			return client
		}
		log.Printf("redis connection failed: %v, retrying", err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("redis connection failed after 5 retries")
	return nil
}

func AddJob(e *constants.TaskCache) {
	if (client.HExists(context.Background(), REDIS_MAP_KEY, e.ID)).Val() {
		log.Printf("update %s already exists in redis", e.ID)
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("marshal update cache failed: %v", err)
	}
	var z = redis.Z{
		Score:  float64(e.ExecutionTime.Unix()),
		Member: data,
	}
	client.ZAdd(context.Background(), REDIS_PQ_KEY, z)
	Qlen++
	client.HSet(context.Background(), REDIS_MAP_KEY, e.ID, data)

	if CheckWithinThreshold(e.ExecutionTime) {
		client.Publish(context.Background(), REDIS_CHANNEL, TASK_AVAILABLE)
	}
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
			log.Printf("unmarshal update cache failed: %v", err)
		}
		client.HDel(context.Background(), REDIS_MAP_KEY, e.ID)
	} else {
		log.Printf("no update cache in redis")
	}
	return &e
}

// RemoveJobByID When the manager knows the job is dispatched and processed successfully by workers, remove it from queue and map
func RemoveJobByID(id string) {
	jobData := client.HGet(context.Background(), REDIS_MAP_KEY, id)
	if jobData.Err() != nil {
		log.Printf("get job data failed: %v", jobData.Err())
	} else {
		client.ZRem(context.Background(), REDIS_PQ_KEY, jobData.Val())
		client.HDel(context.Background(), REDIS_MAP_KEY, id)
		Qlen--
	}
}

func GetJobByID(id string) *constants.TaskCache {
	jobData, err := client.HGet(context.Background(), REDIS_MAP_KEY, id).Result()
	if err != nil {
		log.Printf("get job data failed: %v", err)
	}
	var e constants.TaskCache
	if err := json.Unmarshal([]byte(jobData), &e); err != nil {
		log.Printf("unmarshal update cache failed: %v", err)
	}
	return &e
}

func CheckTasksInDuration(executionTime time.Time, duration time.Duration) bool {
	now := time.Now()
	return executionTime.After(now) && executionTime.Before(now.Add(duration))
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
			log.Printf("unmarshal update cache failed: %v", err)
		}
	} else {
		log.Printf("no update cache in redis")
		return nil
	}
	return &e
}

func CheckWithinThreshold(executionTime time.Time) bool {
	result := time.Until(executionTime) <= PROXIMITY_THRESHOLD
	return result
}

func GetJobsForDispatchWithBuffer() []*constants.TaskCache {
	adjustedTime := time.Now().Add(-DISPATCHBUFFER)
	var scoreRange = &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", adjustedTime.Unix()),
	}
	results, err := client.ZRangeByScoreWithScores(context.Background(), REDIS_PQ_KEY, scoreRange).Result()
	if err != nil {
		log.Printf("get next job failed: %v", err)
	}

	var matureTasks []*constants.TaskCache
	for _, result := range results {
		var e constants.TaskCache
		if err := json.Unmarshal([]byte(result.Member.(string)), &e); err != nil {
			log.Printf("unmarshal update cache failed: %v", err)
		}
		matureTasks = append(matureTasks, &e)
	}
	return matureTasks
}

func PopJobsForDispatchWithBuffer() []*constants.TaskCache {
	adjustedTime := time.Now().Add(-DISPATCHBUFFER)
	var scoreRange = &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", adjustedTime.Unix()),
	}
	results, err := client.ZRangeByScoreWithScores(context.Background(), REDIS_PQ_KEY, scoreRange).Result()
	if err != nil {
		log.Printf("get next job failed: %v", err)
	}

	var matureTasks []*constants.TaskCache
	for _, result := range results {
		var e constants.TaskCache
		if err := json.Unmarshal([]byte(result.Member.(string)), &e); err != nil {
			log.Printf("unmarshal update cache failed: %v", err)
		}
		client.ZRem(context.Background(), REDIS_PQ_KEY, result.Member)
		client.HDel(context.Background(), REDIS_MAP_KEY, e.ID)
		Qlen--
		matureTasks = append(matureTasks, &e)
	}
	return matureTasks
}

func SetLeaseWithID(taskID string, duration time.Duration) error {
	leaseKey := fmt.Sprintf("lease:update:%s", taskID)
	err := client.SetEx(context.Background(), leaseKey, REDIS_LEASE_MAP_VALUE_PROCESSING, duration).Err()
	if err != nil {
		return fmt.Errorf("set lease failed: %v", err)
	}
	return nil
}
