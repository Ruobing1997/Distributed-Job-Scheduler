package data_structure_redis

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
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
			ID:                "TestJob1",
			ExecutionTime:     time.Now().Add(time.Minute),
			NextExecutionTime: time.Now().Add(time.Minute),
			Payload:           "TestPayload1",
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
			ID:                "TestJob2",
			ExecutionTime:     time.Now().Add(2 * time.Minute),
			NextExecutionTime: time.Now().Add(2 * time.Minute),
			Payload:           "TestPayload2",
		}
		task2 := &constants.TaskCache{
			ID:                "TestJob3",
			ExecutionTime:     time.Now().Add(1 * time.Minute),
			NextExecutionTime: time.Now().Add(1 * time.Minute),
			Payload:           "TestPayload3",
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
