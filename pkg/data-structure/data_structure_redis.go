// Package data_structure_redis Description: redis data structure
// The redis will be used as a priority queue to store the tasks
// and a map to store the task id and the task itself.
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

// Init initializes the redis client
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

// maximumConnectionRetry tries to connect to redis for 5 times
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

// AddJob adds a new job to the redis priority queue and map
func AddJob(e *constants.TaskCache) {
	if (client.HExists(context.Background(), REDIS_MAP_KEY, e.ID)).Val() {
		log.Printf("task %s already exists in redis", e.ID)
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
}

// AddRetry adds a new job to the redis list
func AddRetry(e *constants.TaskCache) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("marshal update cache failed: %v", err)
	}
	client.LPush(context.Background(), REDIS_RETRY_KEY, data)
	client.Publish(context.Background(), REDIS_CHANNEL, RETRY_AVAILABLE)
}

// PopRetry pops a new job from the redis list
func PopRetry() (*constants.TaskCache, error) {
	result, err := client.RPop(context.Background(), REDIS_RETRY_KEY).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no retry task in redis")
		}
		log.Printf("get next job failed: %v", err)
		return nil, err
	}

	var e constants.TaskCache
	// Decode the job name and execution time from the JSON string
	if len(result) > 0 {
		if err := json.Unmarshal([]byte(result), &e); err != nil {
			log.Printf("unmarshal update cache failed: %v", err)
			return nil, err
		}
	}
	return &e, nil
}

// PopNextJob pops a new job from the redis priority queue
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
		log.Printf("no task cache in redis")
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

// CheckTasksInDuration checks if the task is in the duration
func CheckTasksInDuration(executionTime time.Time, duration time.Duration) bool {
	now := time.Now().UTC()
	return executionTime.After(now) && executionTime.Before(now.Add(duration))
}

// GetQLength returns the length of the priority queue
func GetQLength() int64 {
	return Qlen
}

// PopJobsForDispatchWithBuffer pops a batch of jobs from the redis priority queue
func PopJobsForDispatchWithBuffer() []*constants.TaskCache {
	adjustedTime := time.Now().UTC()
	var scoreRange = &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", adjustedTime.Unix()),
	}
	results, err := client.ZRangeByScoreWithScores(context.Background(), REDIS_PQ_KEY, scoreRange).Result()
	if err != nil {
		log.Printf("get next job failed: %v", err)
	}

	var matureTasks []*constants.TaskCache
	var taskIDs []string

	for _, result := range results {
		var e constants.TaskCache
		if err := json.Unmarshal([]byte(result.Member.(string)), &e); err != nil {
			log.Printf("unmarshal update cache failed: %v", err)
		}
		taskIDs = append(taskIDs, e.ID)
		matureTasks = append(matureTasks, &e)
	}
	client.ZRemRangeByScore(context.Background(), REDIS_PQ_KEY, scoreRange.Min, scoreRange.Max)
	client.HDel(context.Background(), REDIS_MAP_KEY, taskIDs...)
	Qlen -= int64(len(taskIDs))

	return matureTasks
}

// SetLeaseWithID sets a lease for a task
func SetLeaseWithID(taskID string, execID string, duration time.Duration) error {
	log.Printf("----------------------------")
	log.Printf("set lease for task %s + execute %s", taskID, execID)
	leaseKey := fmt.Sprintf("lease:task:%sexecute:%s", taskID, execID)
	err := client.SetEx(context.Background(), leaseKey,
		REDIS_LEASE_MAP_VALUE_PROCESSING, duration).Err()
	if err != nil {
		return fmt.Errorf("set lease failed: %v", err)
	}
	return nil
}

// RemoveLeaseWithID removes a lease for a task
func RemoveLeaseWithID(ctx context.Context, taskID string, execID string) error {
	leaseKey := fmt.Sprintf("lease:task:%sexecute:%s", taskID, execID)
	err := client.Del(ctx, leaseKey).Err()
	if err != nil {
		return fmt.Errorf("remove lease failed: %v", err)
	}
	return nil
}
