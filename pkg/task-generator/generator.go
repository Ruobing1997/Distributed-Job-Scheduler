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

func GenerateTask(name string, taskType int, cronExpression string,
	payload string, callBackURL string, retries int) *constants.TaskDB {
	id := uuid.New().String()
	taskDB := generateTaskDB(id, name, taskType, cronExpression, payload, callBackURL, retries)
	return taskDB
}

func generateTaskDB(id string, name string, taskType int, cronExpression string,
	payload string, callBackURL string, retries int) *constants.TaskDB {
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		panic(err)
	}
	executionTime := schedule.Next(time.Now())
	nextExecutionTime := schedule.Next(executionTime)
	taskDB := constants.TaskDB{
		ID:                id,
		Name:              name,
		JobType:           taskType,
		CronExpr:          cronExpression,
		Payload:           payload,
		CallbackURL:       callBackURL,
		Status:            0,
		ExecutionTime:     executionTime,
		NextExecutionTime: nextExecutionTime,
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
		Retries:           retries,
		Result:            0,
	}
	return &taskDB
}

// TODO: update content of task cache when determine the structure of task cache
func GenerateTaskCache(id string, jobType int, cronExpr string,
	executionTime time.Time, retriesLeft int, payload constants.PayloadJson) *constants.TaskCache {
	return &constants.TaskCache{
		ID:            id,
		JobType:       jobType,
		CronExpr:      cronExpr,
		ExecutionTime: executionTime,
		RetriesLeft:   retriesLeft,
		Payload:       payload,
	}
}

func GeneratePayloadJSON(executeType int, content string) *constants.PayloadJson {
	return &constants.PayloadJson{
		Type:    executeType,
		Content: content,
	}
}

func marshallTask(task constants.TaskDB) []byte {
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Fatal(err)
	}
	return taskJson
}
