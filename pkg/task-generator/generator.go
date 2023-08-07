// Package generator is used to generate update for update manager.
// It will handle the error cases and generate a formatted update.

package generator

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"time"
)

// GenerateTask is used to generate a task for task manager.
func GenerateTask(jobName string, jobType int, cronExpr string, format int,
	script string, retries int) *constants.TaskDB {
	id := uuid.New().String()
	taskDB := generateTaskDB(id, jobName, jobType, cronExpr, format, script, retries)
	return taskDB
}

// generateTaskDB is used to generate a task for database record.
func generateTaskDB(id string, jobName string, jobType int, cronExpr string, format int,
	script string, retries int) *constants.TaskDB {
	executionTime := DecryptCronExpress(cronExpr)
	payload := GeneratePayload(format, script)
	callbackURL := GenerateCallBackURL(id, constants.DOMAIN)
	taskDB := constants.TaskDB{
		ID:                    id,
		JobName:               jobName,
		JobType:               jobType,
		CronExpr:              cronExpr,
		Payload:               payload,
		CallbackURL:           callbackURL,
		Status:                0,
		ExecutionTime:         executionTime,
		PreviousExecutionTime: executionTime,
		CreateTime:            time.Now().UTC(),
		UpdateTime:            time.Now().UTC(),
		Retries:               retries,
	}
	return &taskDB
}

// GenerateCallBackURL is used to generate a callback url for task manager.
func GenerateCallBackURL(id string, domain string) string {
	return fmt.Sprintf("http://%s:8080/tasks/status/%s", domain, id)
}

// DecryptCronExpress is used to decrypt cron expression to execution time.
func DecryptCronExpress(cronExpr string) time.Time {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		panic(err)
	}
	executionTime := schedule.Next(time.Now().UTC())
	return executionTime
}

// GenerateTaskCache is used to generate a task cache to store in redis.
func GenerateTaskCache(id string, jobType int, cronExpr string,
	executionTime time.Time, retriesLeft int, payload *constants.Payload) *constants.TaskCache {
	return &constants.TaskCache{
		ID:            id,
		ExecutionTime: executionTime,
		JobType:       jobType,
		Payload:       payload,
		RetriesLeft:   retriesLeft,
		CronExpr:      cronExpr,
	}
}

// GenerateExecutionID is used to generate a execution id for task manager.
func GenerateExecutionID() string {
	return uuid.New().String()
}

// GeneratePayload is used to generate a payload for task.
func GeneratePayload(jobType int, script string) *constants.Payload {
	return &constants.Payload{
		Format: jobType,
		Script: script,
	}
}

// GenerateRunTimeTaskThroughTaskCache is used to generate a runtime task from task cache for task manager.
func GenerateRunTimeTaskThroughTaskCache(task *constants.TaskCache,
	jobStatus int, workerID string) *constants.RunTimeTask {
	runningTaskInfo := &constants.RunTimeTask{
		ID:            task.ID,
		ExecutionTime: task.ExecutionTime,
		JobType:       task.JobType,
		JobStatus:     jobStatus,
		RetriesLeft:   task.RetriesLeft,
		CronExpr:      task.CronExpr,
		WorkerID:      workerID,
		ExecutionID:   task.ExecutionID,
	}
	runningTaskInfo.Payload = GeneratePayload(task.Payload.Format, task.Payload.Script)
	return runningTaskInfo
}
