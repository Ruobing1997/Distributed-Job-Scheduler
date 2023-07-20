package data_structure_redis

import (
	"context"
	"encoding/json"
	"fmt"
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

func GetClient() *redis.Client {
	return client
}

func AddJob(e *constants.TaskCache) {
	if (client.HExists(context.Background(), REDIS_MAP_KEY, e.ID)).Val() {
		log.Printf("task %s already exists in redis", e.ID)
		return
	}
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
	client.HSet(context.Background(), REDIS_MAP_KEY, e.ID, data)

	if checkWithinThreshold(e.ExecutionTime) {
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
			log.Printf("unmarshal task cache failed: %v", err)
		}
		client.HDel(context.Background(), REDIS_MAP_KEY, e.ID)
	} else {
		log.Printf("no task cache in redis")
	}
	return &e
}

// RemoveJobByID When the manager knows the job is dispatched and processed successfully by workers, remove it from queue and map
func RemoveJobByID(id string) {
	jobData, err := client.HGet(context.Background(), REDIS_MAP_KEY, id).Result()
	if err != nil {
		log.Printf("get job data failed: %v", err)
	} else {
		client.ZRem(context.Background(), REDIS_PQ_KEY, jobData)
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
		log.Printf("unmarshal task cache failed: %v", err)
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

func checkWithinThreshold(executionTime time.Time) bool {
	return time.Until(executionTime) <= PROXIMITY_THRESHOLD
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
			log.Printf("unmarshal task cache failed: %v", err)
		}
		matureTasks = append(matureTasks, &e)
	}
	return matureTasks
}
