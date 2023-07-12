package generator

import (
	"crypto/sha1"
	"fmt"
	"time"

	constants "git.woa.com/robingowang/MoreFun_SuperNova/utils"

	"github.com/google/uuid"
)

func GenerateTask(name string, taskType string, schedule string,
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

	shardID := hashIDtoShardID(id)
	fmt.Println(shardID)
	return task
}

// hash the uuid to a shard id (0-shards amount)
func hashIDtoShardID(id string) int {
	hasher := sha1.New()
	hasher.Write([]byte(id))
	hashedUUID := hasher.Sum(nil)

	hashedUUIDInt := int(hashedUUID[0]) |
		int(hashedUUID[1])<<constants.EIGHTBIT |
		int(hashedUUID[2])<<constants.SIXTEENBIT |
		int(hashedUUID[3])<<constants.TWENTYFOURBIT
	result := hashedUUIDInt % constants.SHARDSAMOUNT
	return result
}
