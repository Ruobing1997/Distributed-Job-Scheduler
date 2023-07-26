/*
Package generator is used to generate task for task manager.
It will handle the error cases and generate a formatted task.
*/
package generator

import (
	"encoding/json"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func GenerateTask(jobName string, jobType int, cronExpr string, format int,
	script string, retries int) *constants.TaskDB {
	id := uuid.New().String()
	taskDB := generateTaskDB(id, jobName, jobType, cronExpr, format, script, retries)
	return taskDB
}

func generateTaskDB(id string, jobName string, jobType int, cronExpr string, format int,
	script string, retries int) *constants.TaskDB {
	executionTime := DecryptCronExpress(cronExpr)
	payload := GeneratePayload(format, script)
	taskDB := constants.TaskDB{
		ID:            id,
		JobName:       jobName,
		JobType:       jobType,
		CronExpr:      cronExpr,
		Payload:       payload,
		CallbackURL:   "", // TODO: add callback url
		Status:        0,
		ExecutionTime: executionTime,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Retries:       retries,
	}
	return &taskDB
}

func DecryptCronExpress(cronExpr string) time.Time {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		panic(err)
	}
	executionTime := schedule.Next(time.Now())
	return executionTime
}

// TODO: update content of task cache when determine the structure of task cache
func GenerateTaskCache(id string, jobType int, cronExpr string,
	executionTime time.Time, retriesLeft int, payload *constants.Payload) *constants.TaskCache {
	return &constants.TaskCache{
		ID:            id,
		JobType:       jobType,
		CronExpr:      cronExpr,
		ExecutionTime: executionTime,
		RetriesLeft:   retriesLeft,
		Payload:       payload,
	}
}

func GeneratePayload(jobType int, script string) *constants.Payload {
	return &constants.Payload{
		Format: jobType,
		Script: script,
	}
}

func marshallTask(task constants.TaskDB) []byte {
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Fatal(err)
	}
	return taskJson
}
