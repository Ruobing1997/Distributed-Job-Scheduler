package task_manager

import (
	"context"
	"database/sql"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/mySQL"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/postgreSQL"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"time"
)

var timeTracker time.Time
var taskIDToLeaseMap map[string]time.Time

func Init() {
	timeTracker = time.Now()
	taskIDToLeaseMap = make(map[string]time.Time)
}

func GetLeaseByID(taskID string) time.Time {
	return taskIDToLeaseMap[taskID]
}

func UpdateLeaseByID(taskID string, lease time.Time) {
	taskIDToLeaseMap[taskID] = lease
}

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

func HandleIncomingTasks(client databasehandler.DatabaseClient,
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
	// TODO: Currently only add tasks that will be executed in 10 minute, change DURATION when necessary
	if data_structure_redis.CheckTasksInDuration(taskCache, DURATION) {
		addJob(taskCache)
		log.Printf("Task %s added to priority queue. Now the Q length is: %d",
			taskCache.ID, data_structure_redis.GetQLength())
	}
	return nil
}

func AddTasksToDBWithTickers(db *sql.DB) {
	tasks, _ := postgreSQL.GetTasksInInterval(db, time.Now(), time.Now().Add(DURATION), timeTracker)
	timeTracker = time.Now()
	for _, task := range tasks {
		taskCache := generateTaskCache(task.ID, task.ExecutionTime, task.NextExecutionTime, task.Payload)
		addJob(taskCache)
	}
}

func SubscribeToRedisChannel() {
	pubsub := data_structure_redis.GetClient().Subscribe(context.Background(), data_structure_redis.REDIS_CHANNEL)
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		log.Printf("redis subscribe failed: %v", err)
	}
	ch := pubsub.Channel()
	for msg := range ch {
		if msg.Payload == data_structure_redis.TASK_AVAILABLE {
			// check redis priority queue and dispatch tasks
			matureTasks := data_structure_redis.GetJobsForDispatchWithBuffer()
			for _, task := range matureTasks {
				id, status, err := dispatch.HandoutTasks(task)
				UpdateLeaseByID(task.ID, time.Now().Add(2*time.Second))

				if err != nil {
					updateDatabaseWithDispatchResult(id, status)
				}
			}
		}
	}
}

func updateDatabaseWithDispatchResult(id string, workerStatusCode int) {
	switch workerStatusCode {
	case dispatch.JobDispatched:
		// update task status to 1
		// happens once the job is dispatched to worker
		// only update the task_db table in postgresql
	case dispatch.WorkerSucceed:
		// update task status to 2
		// happens once the job is completed
		// update task_db and execution_record, delete the task from redis priority queue
	case dispatch.WorkerFailed:
		// update task status to 3
		// happens once the job is failed
		// check retries, if retries > 0, add the task to redis priority queue
		// update task_db and execution_record, reduce retries
		// if retries == 0, delete the task from redis priority queue
	}
}

func dispatchTasksIJson(task *constants.TaskCache) {
	workerStatusCode, err := dispatch.SendTasksToWorker(task)
	if err != nil {
		log.Printf("dispatch task %s failed: %v", task.ID, err)
		return
	}
	updateDatabaseWithDispatchResult(task.ID, workerStatusCode)
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
		ID:            id,
		ExecutionTime: executionTime,
		Payload:       payload,
	}
}
