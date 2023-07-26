package data_structure_redis

import (
	"context"
	"encoding/json"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddJobAndPopNextJob(t *testing.T) {
	// Initialize the Redis client
	Init()

	// Test case 1: Add a single job and get it
	t.Run("AddAndPopSingleJob", func(t *testing.T) {
		// Create a TaskCache instance
		task := &constants.TaskCache{
			ID:            "TestJob1",
			ExecutionTime: time.Now().Add(time.Minute),
			Payload:       "TestPayload1",
		}

		// Add the job
		AddJob(task)
		Qlen := GetQLength()
		if Qlen != 1 {
			t.Errorf("Expected Q length 1, got %d", Qlen)
		}
		// Get the next job
		nextJob := PopNextJob()

		// Check if the added job and the retrieved job are the same
		if nextJob.ID != task.ID {
			t.Errorf("Expected job name %s, got %s", task.ID, nextJob.ID)
		}

		Qlen = GetQLength()
		if Qlen != 0 {
			t.Errorf("Expected Q length 0, got %d", Qlen)
		}
	})

	// Test case 2: Add multiple jobs and get them in order
	t.Run("AddAndPopMultipleJobs", func(t *testing.T) {
		// Create TaskCache instances
		task1 := &constants.TaskCache{
			ID:            "TestJob2",
			ExecutionTime: time.Now().Add(2 * time.Minute),
			Payload:       "TestPayload2",
		}
		task2 := &constants.TaskCache{
			ID:            "TestJob3",
			ExecutionTime: time.Now().Add(1 * time.Minute),
			Payload:       "TestPayload3",
		}

		// Add the jobs
		AddJob(task1)
		AddJob(task2)

		Qlen := GetQLength()
		if Qlen != 2 {
			t.Errorf("Expected Q length 2, got %d", Qlen)
		}
		// Get the first job (should be task2)
		nextJob1 := PopNextJob()
		if nextJob1.ID != task2.ID {
			t.Errorf("Expected job name %s, got %s", task2.ID, nextJob1.ID)
		}
		Qlen = GetQLength()
		if Qlen != 1 {
			t.Errorf("Expected Q length 1, got %d", Qlen)
		}
		// Get the second job (should be task1)
		nextJob2 := PopNextJob()
		if nextJob2.ID != task1.ID {
			t.Errorf("Expected job name %s, got %s", task1.ID, nextJob2.ID)
		}
		Qlen = GetQLength()
		if Qlen != 0 {
			t.Errorf("Expected Q length 0, got %d", Qlen)
		}
	})

	// Test case 3: Get a job from an empty queue
	t.Run("PopJobFromEmptyQueue", func(t *testing.T) {
		// Get the next job (should be nil)
		nextJob := PopNextJob()
		if nextJob.ID != "" {
			t.Errorf("Expected an empty job, got %s", nextJob.ID)
		}
		Qlen := GetQLength()
		if Qlen != 0 {
			t.Errorf("Expected Q length 0, got %d", Qlen)
		}
	})
}

func setupTestRedisTasks(client *redis.Client) {
	// 这里我们添加5个任务，其中3个任务的执行时间在DISPATCHBUFFER之内，2个在之外
	now := time.Now()
	tasks := []constants.TaskCache{
		{ID: "task1", ExecutionTime: now.Add(-5 * time.Minute)},
		{ID: "task2", ExecutionTime: now.Add(-15 * time.Minute)},
		{ID: "task3", ExecutionTime: now.Add(-25 * time.Minute)},
		{ID: "task4", ExecutionTime: now.Add(5 * time.Minute)},
		{ID: "task5", ExecutionTime: now.Add(15 * time.Minute)},
	}

	for _, task := range tasks {
		data, _ := json.Marshal(task)
		client.ZAdd(context.Background(), REDIS_PQ_KEY, redis.Z{
			Score:  float64(task.ExecutionTime.Unix()),
			Member: data,
		})
		client.HSet(context.Background(), REDIS_MAP_KEY, task.ID, data)
	}
}

func TestPopJobsForDispatchWithBuffer(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()

	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	setupTestRedisTasks(client)

	matureTasks := PopJobsForDispatchWithBuffer()

	assert.Equal(t, 3, len(matureTasks))
	now := time.Now()
	for _, task := range matureTasks {
		assert.True(t, task.ExecutionTime.Before(now))
	}

	for _, task := range matureTasks {
		existsInZSet := client.ZScore(context.Background(), REDIS_PQ_KEY, task.ID).Val()
		existsInHash := client.HExists(context.Background(), REDIS_MAP_KEY, task.ID).Val()
		assert.Equal(t, float64(0), existsInZSet)
		assert.False(t, existsInHash)
	}
}
