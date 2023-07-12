package generator

import (
	"context"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateTask(t *testing.T) {
	name := "Test Task"
	taskType := "test"
	schedule := "* * * * *"
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := generateTask(name, taskType, schedule, payload, callbackURL)

	if task.Name != name || task.Type != taskType || task.Schedule != schedule ||
		task.Payload != payload || task.CallbackURL != callbackURL {
		t.Errorf("generateTask() failed, expected task with name: %s, "+
			"type: %s, schedule: %s, payload: %s, callbackURL: %s, got: %+v",
			name, taskType, schedule, payload, callbackURL, task)
	}
}

func TestMarshallTask(t *testing.T) {
	name := "Test Task"
	taskType := "test"
	schedule := "* * * * *"
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := generateTask(name, taskType, schedule, payload, callbackURL)
	taskJson := marshallTask(task)

	if len(taskJson) == 0 {
		t.Errorf("marshallTask() failed, expected taskJson with length > 0, got: %d",
			len(taskJson))
	}
}

func TestStoreDataToRedis(t *testing.T) {
	mr, err := miniredis.Run()

	if err != nil {
		panic(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	database.InitializeRedisRing(map[string]string{"shard1": mr.Addr()})

	name := "Test Task"
	taskType := "test"
	schedule := "* * * * *"
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := generateTask(name, taskType, schedule, payload, callbackURL)
	taskJson := marshallTask(task)

	assert.NoError(t, err)

	val, err := client.Get(context.Background(), task.ID).Result()
	assert.NoError(t, err)
	assert.Equal(t, string(taskJson), val)
}
