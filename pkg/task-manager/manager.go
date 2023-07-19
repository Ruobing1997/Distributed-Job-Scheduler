package task_manager

import (
	"context"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/mySQL"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"time"
)

// TODO: remember to init the databases and redisPQ in main.go
func CreateRedisPQ() {
	data_structure_redis.Init()
}

func addJob(e *constants.TaskCache) {
	data_structure_redis.AddJob(e)
}

func GetNextJob() *constants.TaskCache {
	return data_structure_redis.PopNextJob()
}

func storeTasksToDB(client databasehandler.DatabaseClient, taskDB *constants.TaskDB) error {
	err := client.InsertTask(context.Background(), taskDB)
	if err != nil {
		return err
	}
	return nil
}

func HandleTasks(client databasehandler.DatabaseClient,
	name string, taskType int, cronExpression string,
	payload string, callBackURL string, retries int) error {
	// get task generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, payload, callBackURL, retries)
	// insert task to database
	err := storeTasksToDB(client, taskDB)
	if err != nil {
		return err
	}
	// generate task for cache:
	taskCache := generateTaskCache(taskDB.ID, taskDB.ExecutionTime, taskDB.NextExecutionTime, taskDB.Payload)
	// insert task to priority queue
	// TODO: Currently only add tasks that will be executed in 1 minute, change DURATION when necessary
	if data_structure_redis.CheckTasksInDuration(taskCache, DURATION) {
		addJob(taskCache)

		log.Printf("Task %s added to priority queue. Now the Q length is: %d",
			taskCache.ID, data_structure_redis.GetQLength())
	}
	return nil
}

func DispatchTasks() {

}

func insertCacheToDB(taskCache *constants.TaskCache, worker_ip string, job_status int) error {
	query := `INSERT INTO execution_record (job_id, execution_time, worker_ip, job_status) VALUES (?, ?, ?, ?)`
	_, err := mySQL.GetDB().ExecContext(context.Background(), query, taskCache.ID, taskCache.ExecutionTime, worker_ip, job_status)
	if err != nil {
		return err
	}
	return nil
}

func generateTaskCache(id string, executionTime time.Time,
	nextExecutionTime time.Time, payload string) *constants.TaskCache {
	return &constants.TaskCache{
		ID:                id,
		ExecutionTime:     executionTime,
		NextExecutionTime: nextExecutionTime,
		Payload:           payload,
	}
}
