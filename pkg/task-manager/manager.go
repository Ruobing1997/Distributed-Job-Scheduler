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
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"log"
	"strings"
	"time"
)

var timeTracker time.Time
var taskIDToLeaseMap map[string]time.Time
var client *postgreSQL.Client
var mySQLClient *mySQL.Client

func Init() {
	timeTracker = time.Now()
	taskIDToLeaseMap = make(map[string]time.Time)
	InitRedisDataStructure()
	SubscribeToRedisChannel()
	client = postgreSQL.NewpostgreSQLClient()
}

func GetLeaseByID(taskID string) time.Time {
	return taskIDToLeaseMap[taskID]
}

func UpdateLeaseByID(taskID string, lease time.Time) {
	taskIDToLeaseMap[taskID] = lease
}

// TODO: remember to init the databases and redisPQ in main.go (call init)
func InitRedisDataStructure() {
	data_structure_redis.Init()
}

func addJob(e *constants.TaskCache) {
	data_structure_redis.AddJob(e)
}

func GetNextJob() *constants.TaskCache {
	return data_structure_redis.PopNextJob()
}

// HandleNewTasks handles new tasks from API,  这里应该是入口函数。主要做创建任务的逻辑
func HandleNewTasks(client databasehandler.DatabaseClient,
	name string, taskType int, cronExpression string,
	format int, script string, callBackURL string, retries int) error {
	// get task generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, format, script, retries)
	// insert task to database
	err := client.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, taskDB)
	if err != nil {
		return err
	}
	// generate task for cache:
	taskCache := generator.GenerateTaskCache(
		taskDB.ID,
		taskDB.JobType,
		taskDB.CronExpr,
		taskDB.ExecutionTime,
		taskDB.Retries,
		taskDB.Payload)
	// insert task to priority queue
	// TODO: Currently only add tasks that will be executed in 10 minute, change DURATION when necessary
	if data_structure_redis.CheckTasksInDuration(taskCache, DURATION) {
		addJob(taskCache)
		log.Printf("Task %s added to priority queue. Now the Q length is: %d",
			taskCache.ID, data_structure_redis.GetQLength())
	}
	return nil
}

func AddTasksFromDBWithTickers(db *sql.DB) {
	tasks, _ := client.GetTasksInInterval(time.Now(), time.Now().Add(DURATION), timeTracker)
	timeTracker = time.Now()
	for _, task := range tasks {
		taskCache := generator.GenerateTaskCache(
			task.ID, task.JobType, task.CronExpr, task.ExecutionTime, task.Retries, task.Payload)
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
			executeMatureTasks()
		}
	}
}

func executeMatureTasks() {
	matureTasks := data_structure_redis.PopJobsForDispatchWithBuffer()
	for _, task := range matureTasks {
		// TODO: need to update the lease time, the current duration is 2 seconds
		data_structure_redis.SetLeaseWithID(task.ID, 2*time.Second)
		// TODO: _ is the workerID, currently not used, will used for routing key in future
		_, status, err := dispatch.HandoutTasksForExecuting(task)
		if err != nil {
			log.Printf("dispatch task %s failed: %v", task.ID, err)
		}
		updateDatabaseWithDispatchResult(task, status)
	}
}

// updateLeaseMapWithDispatchResult updates the lease map based on the dispatch result
func updateLeaseMapWithDispatchResult(task *constants.TaskCache, jobStatusCode int) {
	switch jobStatusCode {
	case dispatch.JobSucceed:
		// delete the task from the lease map
		delete(taskIDToLeaseMap, task.ID)
	case dispatch.JobFailed:

	}
}

func updateDatabaseWithDispatchResult(task *constants.TaskCache, jobStatusCode int) {
	switch jobStatusCode {
	case dispatch.JobDispatched:
		// designed for future use, with grpc we will get response in sync manner
		// update task status to 1
		// happens once the job is dispatched to worker
		// only update the task_db table in postgresql
		client.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID, map[string]interface{}{"job_status": jobStatusCode})
	case dispatch.JobSucceed:
		// update task status to 2 in task_db, the task should be deleted from ExecutionRecord
		// TODO: Integrate tables' manipulation together, currently we have two functions do similar thing
		client.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
		if task.JobType == constants.Recurring {
			// it is recur job, update execution time, update time, status, (previous execution time) may not need, since we have update time
			cronExpr := task.CronExpr
			newExecutionTime := generator.DecryptCronExpress(cronExpr)
			updateVars := map[string]interface{}{
				"execution_time": newExecutionTime,
				"update_time":    time.Now(),
				"status":         constants.JOBREENTERED,
			}
			client.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
		} else {
			// if it is one time job, only update the task_db table
			updateVars := map[string]interface{}{
				"update_time": time.Now(),
				"status":      constants.JOBSUCCEED,
			}
			client.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
		}
		// happens once the job is completed
		// update task_db and execution_record, delete the task from redis priority queue -> no need, it is already popped
	case dispatch.JobFailed:
		// update task status to 3
		// task_db only update when the task is completely failed (retries == 0 or outdated)

		// happens once the job is failed
		// check retries, if retries > 0, add the task to redis priority queue
		// update task_db and execution_record, reduce retries
		// if retries == 0, delete the task from redis priority queue

		if checkJobCompletelyFailed(task) {
			// task completely failed need to update all information for report user.
			client.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
			client.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
				map[string]interface{}{"job_status": jobStatusCode})
		} else {
			// task not completely failed, update the retries left, time
			// and add to redis priority queue
			RescheduleFailedJobs(task)
		}
	}

}

// RescheduleFailedJobs reenter the task to queue and update the retries
func RescheduleFailedJobs(task *constants.TaskCache) {
	curRetries := task.RetriesLeft - 1
	client.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID,
		map[string]interface{}{"retries_left": curRetries})
	newExecutionTime, _ := cron.ParseStandard(task.CronExpr)
	newNextExecutionTime := newExecutionTime.Next(time.Now())
	err := client.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
		map[string]interface{}{"retries_left": curRetries,
			"execution_time": newNextExecutionTime})
	task.ExecutionTime = newNextExecutionTime
	task.RetriesLeft = curRetries
	data_structure_redis.AddJob(task)
}

// TODO: currently we use database to check the retries left. We may not need to.
func checkJobCompletelyFailed(data interface{}) bool {
	jobType, retriesLeft, executionTime := getJobInfo(data)
	if jobType == constants.OneTime {
		if checkJobOutdated(executionTime) {
			// task completely failed need to update all information for report user.
			return true
		}
	}
	// it will be either the task is recurring or task is one time but not outdated,
	// we check retries left. If retries left is <= 0, the task is completely failed
	return retriesLeft <= 0
}

func checkJobCompletelySucceed(status int) bool {
	return status == constants.JOBSUCCEED
}

// check if the task is outdated, current time is later than the execution time
func checkJobOutdated(executionTime time.Time) bool {
	return executionTime.Before(time.Now())
}

func dispatchTasksIJson(task *constants.TaskCache) {
	workerStatusCode, err := dispatch.SendTasksToWorker(task)
	if err != nil {
		log.Printf("dispatch task %s failed: %v", task.ID, err)
		return
	}
	updateDatabaseWithDispatchResult(task, workerStatusCode)
}

func insertCacheToDB(taskCache *constants.TaskCache, worker_ip string, job_status int) error {
	query := `INSERT INTO execution_record (job_id, execution_time, worker_ip, job_status) VALUES (?, ?, ?, ?)`
	_, err := mySQLClient.Db.ExecContext(context.Background(), query, taskCache.ID, taskCache.ExecutionTime, worker_ip, job_status)
	if err != nil {
		return err
	}
	return nil
}

// HandleExpiryTasks handles the expiry tasks
func HandleExpiryTasks(msg *redis.Message) {
	// The key format is "lease:task:<task_id>"
	if strings.HasPrefix(msg.Payload, "lease:task:") {
		taskID := strings.TrimPrefix(msg.Payload, "lease:task:")
		log.Printf("Task %s is expired", taskID)
		HandleUnRenewLeaseJobThroughDB(taskID)
	}
}

func HandleUnRenewLeaseJobThroughDB(id string) error {
	// check job type (through runtime db table)
	// if job type is one time, directly return fail
	// if job type is recur, check retry times
	// if > 0, reenter queue, update redis, db; otherwise return fail
	record, err := client.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, id)
	runTimeTask := record.(*constants.RunTimeTask)
	if err != nil {
		return err
	}
	if checkJobCompletelyFailed(runTimeTask) {
		// task completely failed need to update all information for report user.
		client.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, runTimeTask.ID)
		client.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
			runTimeTask.ID, map[string]interface{}{"job_status": constants.JOBFAILED})
	} else {
		// task not completely failed, update the retries left
		task := generator.GenerateTaskCache(
			runTimeTask.ID,
			runTimeTask.JobType,
			runTimeTask.CronExpr,
			runTimeTask.ExecutionTime,
			runTimeTask.RetriesLeft,
			runTimeTask.Payload,
		)
		RescheduleFailedJobs(task)
	}
	return nil
}

func getJobInfo(data interface{}) (int, int, time.Time) {
	switch v := data.(type) {
	case constants.TaskCache:
		return v.JobType, v.RetriesLeft, v.ExecutionTime
	case constants.RunTimeTask:
		return v.JobType, v.RetriesLeft, v.ExecutionTime
	default:
		panic("unsupported data type")
	}
}
