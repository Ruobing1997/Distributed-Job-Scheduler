package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"time"

	"github.com/google/uuid"
)

func generateTask(name string, taskType string, schedule string,
	payload string, callbackURL string) constants.Task {
	id := uuid.New().String()
	task := constants.Task{
		ID:          id,
		Name:        name,
		Type:        taskType,
		Schedule:    schedule,
		Payload:     payload,
		CallbackURL: callbackURL,
		Status:      0,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		Retries:     0,
		Result:      "",
	}

	_ = storeDataToRedis(id, marshallTask(task))

	return task
}

func marshallTask(task constants.Task) []byte {
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Fatal(err)
	}
	return taskJson
}

// use redis ringOptions to store data, the ringOptions uses Consistent Hashing
func storeDataToRedis(key string, value []byte) error {
	redisRing := database.GetRedisRing()

	err := redisRing.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	fmt.Printf("Stored key %s with value %s to the Ring\n", key, value)
	return nil
}
