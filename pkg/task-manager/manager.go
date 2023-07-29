package task_manager

import (
	"context"
	"fmt"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/postgreSQL"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	timeTracker          time.Time
	databaseClient       *postgreSQL.Client
	redisClient          *redis.Client
	testMessageChannel   = make(chan string)
	isTaskBeingProcessed bool
	taskProcessingLock   sync.Mutex
)

func Init() {
	timeTracker = time.Now()
	// TODO: make the user to choose database type
	databaseClient = postgreSQL.NewpostgreSQLClient()
	redisClient = data_structure_redis.Init()
}

func Start() {
	startedChan := make(chan bool, 4)
	var wg sync.WaitGroup
	startGoroutine := func(f func(chan bool)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f(startedChan)
		}()
	}

	startGoroutine(func(ch chan bool) {
		ch <- true
		ListenForExpiryNotifications()
	})

	startGoroutine(func(ch chan bool) {
		ch <- true
		SubscribeToRedisChannel()
	})

	startGoroutine(func(ch chan bool) {
		ch <- true
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				AddTasksFromDBWithTickers()
			}
		}
	})

	startGoroutine(func(ch chan bool) {
		ch <- true
		MonitorHeadNode()
	})

	for i := 0; i < 4; i++ {
		<-startedChan
	}
}

func addJobToRedis(taskDB *constants.TaskDB) {
	// generate update for cache:
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

// HandleNewTasks handles new tasks from API,  这里应该是入口函数。主要做创建任务的逻辑
func HandleNewTasks(name string, taskType int, cronExpression string,
	format int, script string, callBackURL string, retries int) (string, error) {
	// get update generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, format, script, retries)
	// insert update to database
	err := databaseClient.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, taskDB)
	if err != nil {
		return "nil", err
	}
	// TODO: Currently only add tasks that will be executed in 10 minute, change DURATION when necessary
	if data_structure_redis.CheckTasksInDuration(taskDB.ExecutionTime, DURATION) {
		// insert update to priority queue
		addJobToRedis(taskDB)
	}
	return taskDB.ID, nil
}

func HandleDeleteTasks(taskID string) error {
	// if the update is running, stop it
	// check the running_tasks_record board to see the job is running or not
	record, err := databaseClient.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, taskID)
	if err != nil {
		// apply the same logic as the following
		// if the update is in redis priority queue, delete it and delete it from database
		// if the update is in database, delete it from database
		data_structure_redis.RemoveJobByID(taskID)
		err := databaseClient.DeleteByID(context.Background(), constants.TASKS_FULL_RECORD, taskID)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		// record found, stop the job
		runTimeJob := record.(*constants.RunTimeTask)
		workerId, status, err := dispatch.StopRunningTask(runTimeJob)
		log.Printf("workerId: %s, status: %s, err: %v", workerId, status, err)
		return err
	}
}

func HandleUpdateTasks(taskID string, args map[string]interface{}) error {
	// do nothing when the update is running, report the user cannot update the update when it is running
	// if the update is in redis priority queue, update it and update it in database
	data_structure_redis.RemoveJobByID(taskID)
	var taskCache *constants.TaskCache
	taskCache.ID = taskID
	for key, value := range args {
		switch key {
		case "ExecutionTime":
			taskCache.ExecutionTime = value.(time.Time)
		case "JobType":
			taskCache.JobType = value.(int)
		case "execute_format":
			taskCache.Payload.Format = value.(int)
		case "execute_script":
			taskCache.Payload.Script = value.(string)
		case "RetriesLeft":
			taskCache.RetriesLeft = value.(int)
		case "CronExpr":
			taskCache.CronExpr = value.(string)
		}
	}
	data_structure_redis.AddJob(taskCache)
	// if the update is in database, update it in database
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, taskID, args)
	if err != nil {
		return fmt.Errorf("update update %s failed: %v", taskID, err)
	}
	return nil
}

func HandleGetTasks(taskID string) (*constants.TaskDB, error) {
	record, err := databaseClient.GetTaskByID(context.Background(), constants.TASKS_FULL_RECORD, taskID)
	if err != nil {
		return nil, err
	}
	return record.(*constants.TaskDB), nil
}

func HandleGetAllTasks() ([]*constants.TaskDB, error) {
	tasks, err := databaseClient.GetAllTasks()
	if err != nil {
		return nil, err
	}
	return tasks, nil
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
			taskProcessingLock.Lock()
			if !isTaskBeingProcessed {
				isTaskBeingProcessed = true
				fmt.Println("redis channel received: ", msg.Payload)
				// check redis priority queue and dispatch tasks
				executeMatureTasks()
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()

		}
	}
}

func executeMatureTasks() {
	matureTasks := data_structure_redis.PopJobsForDispatchWithBuffer()
	for _, task := range matureTasks {
		fmt.Println("dispatching task: ", task.ID)
		// TODO: need to update the lease time, the current duration is 2 seconds
		data_structure_redis.SetLeaseWithID(task.ID, 2*time.Second)
		// TODO: _ is the workerID, currently not used, will used for routing key in future
		_, status, err := HandoutTasksForExecuting(task)
		if err != nil {
			log.Printf("dispatch %s failed: %v", task.ID, err)
		}
		updateDatabaseWithDispatchResult(task, status)
	}
}

func HandoutTasksForExecuting(task *constants.TaskCache) (string, int, error) {
	workerService := os.Getenv("WORKER_SERVICE")
	if workerService == "" {
		workerService = WORKER_SERVICE
	}
	conn, err := grpc.Dial(workerService+":50051", grpc.WithInsecure())
	if err != nil {
		return task.ID, constants.JOBFAILED, fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTaskServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.GRPC_TIMEOUT)
	defer cancel()

	payload := &pb.Payload{
		Format: int32(task.Payload.Format),
		Script: task.Payload.Script,
	}

	r, err := client.ExecuteTask(ctx, &pb.TaskRequest{
		Id:            task.ID,
		Payload:       payload,
		ExecutionTime: timestamppb.New(task.ExecutionTime),
		MaxRetryCount: int32(task.RetriesLeft),
	})

	if err != nil {
		// over time, no response
		return task.ID, constants.JOBFAILED, fmt.Errorf("could not execute: %v", err)
	}
	log.Printf("Task %s executed with status %d", r.Id, r.Status)
	return r.Id, int(r.Status), nil
}

func MonitorHeadNode() {
	log.Println("Start monitoring head node")
	headNodeTicker := time.NewTicker(1 * time.Second)
	defer headNodeTicker.Stop()
	for {
		select {
		case <-headNodeTicker.C:
			taskProcessingLock.Lock()
			if !isTaskBeingProcessed {
				isTaskBeingProcessed = true
				//log.Println("Head node timer is triggered")
				taskCache := data_structure_redis.GetNextJob()
				if taskCache != nil && data_structure_redis.CheckWithinThreshold(taskCache.ExecutionTime) {
					executeMatureTasks()
				}
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()
		}
	}
}

func updateDatabaseWithDispatchResult(task *constants.TaskCache, jobStatusCode int) {
	switch jobStatusCode {
	case dispatch.JobDispatched:
		// designed for future use, with grpc we will get response in sync manner
		// update update status to 1
		// happens once the job is dispatched to worker
		// only update the task_db table in postgresql
		databaseClient.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID, map[string]interface{}{"job_status": jobStatusCode})
	case dispatch.JobSucceed:
		// the update should be deleted from ExecutionRecord
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
		// update update status to 3
		// task_db only update when the update is completely failed (retries == 0 or outdated)

		if checkJobCompletelyFailed(task) {
			// update completely failed need to update all information for report user.
			databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
			databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
				map[string]interface{}{"job_status": jobStatusCode})
		} else {
			// update not completely failed, update the retries left, time
			// and add to redis priority queue
			RescheduleFailedJobs(task)
		}
	}

}

// RescheduleFailedJobs reenter the update to queue and update the retries
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
			// update completely failed need to update all information for report user.
			return true
		}
	}
	// it will be either the update is recurring or update is one time but not outdated,
	// we check retries left. If retries left is <= 0, the update is completely failed
	return retriesLeft <= 0
}

// check if the update is outdated, current time is later than the execution time
func checkJobOutdated(executionTime time.Time) bool {
	return executionTime.Before(time.Now())
}

func dispatchTasksIJson(task *constants.TaskCache) {
	workerStatusCode, err := dispatch.SendTasksToWorker(task)
	if err != nil {
		log.Printf("dispatch update %s failed: %v", task.ID, err)
		return
	}
	updateDatabaseWithDispatchResult(task, workerStatusCode)
}

// HandleExpiryTasks handles the expiry tasks
func HandleExpiryTasks(msg *redis.Message) {
	// The key format is "lease:update:<task_id>"
	if strings.HasPrefix(msg.Payload, "lease:update:") {
		taskID := strings.TrimPrefix(msg.Payload, "lease:update:")
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
		// update completely failed need to update all information for report user.
		databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, runTimeTask.ID)
		databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
			runTimeTask.ID, map[string]interface{}{"job_status": constants.JOBFAILED})
	} else {
		// update not completely failed, update the retries left
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

func RegisterUser(userInfo *constants.UserInfo) error {
	err := databaseClient.InsertUser(context.Background(), userInfo)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func LoginUser(username string, password string) (bool, error) {
	isValid, err := databaseClient.IsValidCredential(context.Background(), username, password)
	return isValid, err
}
