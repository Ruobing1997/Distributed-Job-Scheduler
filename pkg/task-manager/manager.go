package task_manager

import (
	"context"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
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
	"sync"
	"time"
)

var (
	timeTracker     time.Time
	databaseClient  *postgreSQL.Client
	mySQLClient     *mySQL.Client
	redisClient     *redis.Client
	redisFreeSignal chan struct{}
)

func Init() {
	timeTracker = time.Now()
	// TODO: make the user to choose database type
	databaseClient = postgreSQL.NewpostgreSQLClient()
	redisClient = data_structure_redis.Init()
	go ListenForExpiryNotifications()
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				AddTasksFromDBWithTickers()
			}
		}
	}()
}

func Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ListenForExpiryNotifications()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeToRedisChannel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				AddTasksFromDBWithTickers()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		MonitorHeadNode()
	}()

	wg.Wait()
}

func addJobToRedis(taskDB *constants.TaskDB) {
	// generate task for cache:
	taskCache := generator.GenerateTaskCache(
		taskDB.ID,
		taskDB.JobType,
		taskDB.CronExpr,
		taskDB.ExecutionTime,
		taskDB.Retries,
		taskDB.Payload)
	data_structure_redis.AddJob(taskCache)
	log.Printf("Task %s added to priority queue. Now the Q length is: %d",
		taskCache.ID, data_structure_redis.GetQLength())
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
	// TODO: Currently only add tasks that will be executed in 10 minute, change DURATION when necessary
	if data_structure_redis.CheckTasksInDuration(taskDB.ExecutionTime, DURATION) {
		// insert task to priority queue
		addJobToRedis(taskDB)
	}
	return nil
}

func AddTasksFromDBWithTickers() {
	tasks, _ := databaseClient.GetTasksInInterval(time.Now(), time.Now().Add(DURATION), timeTracker)
	timeTracker = time.Now()
	for _, task := range tasks {
		addJobToRedis(task)
	}
}

func SubscribeToRedisChannel() {
	pubsub := redisClient.Subscribe(context.Background(), data_structure_redis.REDIS_CHANNEL)
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		log.Printf("redis subscribe failed: %v", err)
	}
	ch := pubsub.Channel()
	for msg := range ch {
		if msg.Payload == data_structure_redis.TASK_AVAILABLE {
			redisFreeSignal <- struct{}{}
			// check redis priority queue and dispatch tasks
			executeMatureTasks()
			<-redisFreeSignal
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

func MonitorHeadNode() {
	for {
		if isRedisFree() {
			headNodeTimer := time.NewTimer(1 * time.Minute)
			select {
			case <-headNodeTimer.C:
				taskCache := GetNextJob()
				if taskCache != nil &&
					data_structure_redis.CheckWithinThreshold(taskCache.ExecutionTime) {
					redisFreeSignal <- struct{}{}
					executeMatureTasks()
					<-redisFreeSignal
				}
			case <-redisFreeSignal:
				headNodeTimer.Stop()
				<-redisFreeSignal
			}
		}
	}
}

func isRedisFree() bool {
	select {
	case <-redisFreeSignal:
		return false
	default:
		return true
	}
}

func updateDatabaseWithDispatchResult(task *constants.TaskCache, jobStatusCode int) {
	switch jobStatusCode {
	case dispatch.JobDispatched:
		// designed for future use, with grpc we will get response in sync manner
		// update task status to 1
		// happens once the job is dispatched to worker
		// only update the task_db table in postgresql
		databaseClient.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID, map[string]interface{}{"job_status": jobStatusCode})
	case dispatch.JobSucceed:
		// the task should be deleted from ExecutionRecord
		databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
		if task.JobType == constants.Recurring {
			// it is recur job, update execution time, update time, status, (previous execution time) may not need, since we have update time
			cronExpr := task.CronExpr
			newExecutionTime := generator.DecryptCronExpress(cronExpr)
			updateVars := map[string]interface{}{
				"execution_time": newExecutionTime,
				"update_time":    time.Now(),
				"status":         constants.JOBREENTERED,
			}
			databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
			if data_structure_redis.CheckTasksInDuration(newExecutionTime, DURATION) {
				data_structure_redis.AddJob(task)
			}
		} else {
			// if it is one time job, only update the task_db table
			updateVars := map[string]interface{}{
				"update_time": time.Now(),
				"status":      constants.JOBSUCCEED,
			}
			databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
		}
	case dispatch.JobFailed:
		// update task status to 3
		// task_db only update when the task is completely failed (retries == 0 or outdated)

		if checkJobCompletelyFailed(task) {
			// task completely failed need to update all information for report user.
			databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
			databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
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
	databaseClient.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID,
		map[string]interface{}{"retries_left": curRetries})
	newExecutionTime, _ := cron.ParseStandard(task.CronExpr)
	newNextExecutionTime := newExecutionTime.Next(time.Now())
	databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
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
	record, err := databaseClient.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, id)
	runTimeTask := record.(*constants.RunTimeTask)
	if err != nil {
		return err
	}
	if checkJobCompletelyFailed(runTimeTask) {
		// task completely failed need to update all information for report user.
		databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, runTimeTask.ID)
		databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
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

func ListenForExpiryNotifications() {
	pubsub := redisClient.Subscribe(context.Background(), REDIS_EXPIRY_CHANNEL)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			log.Printf("receive message failed: %v", err)
			continue
		}
		HandleExpiryTasks(msg)
	}
}
