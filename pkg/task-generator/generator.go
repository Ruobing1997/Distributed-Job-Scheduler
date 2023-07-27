/*
Package generator is used to generate update for update manager.
It will handle the error cases and generate a formatted update.
*/
package generator

import (
	"encoding/json"
	"fmt"
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
	callbackURL := GenerateCallBackURL(id, constants.DOMAIN)
	taskDB := constants.TaskDB{
		ID:            id,
		JobName:       jobName,
		JobType:       jobType,
		CronExpr:      cronExpr,
		Payload:       payload,
		CallbackURL:   callbackURL,
		Status:        0,
		ExecutionTime: executionTime,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Retries:       retries,
	}
	return &taskDB
}

func GenerateCallBackURL(id string, domain string) string {
	return fmt.Sprintf("http://%s:8080/tasks/status/%s", domain, id)
}

func DecryptCronExpress(cronExpr string) time.Time {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		panic(err)
	}
	executionTime := schedule.Next(time.Now())
	return executionTime
}

// TODO: update content of update cache when determine the structure of update cache
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
