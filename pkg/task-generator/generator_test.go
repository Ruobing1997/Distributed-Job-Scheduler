package generator

import (
	"fmt"
	"testing"

	constants "git.woa.com/robingowang/MoreFun_SuperNova/utils"
	"github.com/google/uuid"
)

func TestGenerateTask(t *testing.T) {
	name := "Test Task"
	taskType := "test"
	schedule := "* * * * *"
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := GenerateTask(name, taskType, schedule, payload, callbackURL)
	fmt.Println(task)

	if task.Name != name || task.Type != taskType || task.Schedule != schedule || task.Payload != payload || task.CallbackURL != callbackURL {
		t.Errorf("GenerateTask() failed, expected task with name: %s, type: %s, schedule: %s, payload: %s, callbackURL: %s, got: %+v", name, taskType, schedule, payload, callbackURL, task)
	}
}

func TestHashIDtoShardID(t *testing.T) {
	id := uuid.New().String()
	fmt.Printf("The UUID is %s", id)

	shardID := hashIDtoShardID(id)
	fmt.Printf("The shardID is %d", shardID)

	if shardID < 0 || shardID >= constants.SHARDSAMOUNT {
		t.Errorf("hashIDtoShardID() failed, expected shardID between 0 and %d, got: %d", constants.SHARDSAMOUNT-1, shardID)
	}
}
