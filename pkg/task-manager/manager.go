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
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"log"
	"net"
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
	managerID            = os.Getenv("HOSTNAME")
)

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

func Init() {
	timeTracker = time.Now()
	logFile, err := os.OpenFile("./logs/managers/manager-"+managerID+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	multiWrite := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWrite)
	// TODO: make the user to choose database type
	databaseClient = postgreSQL.NewpostgreSQLClient()
	redisClient = data_structure_redis.Init()
}

func InitManagerGRPC() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &ServerImpl{})
	pb.RegisterLeaseServiceServer(s, &ServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
	log.Printf("Task %v added to priority queue. Now the Q length is: %d",
		taskCache, data_structure_redis.GetQLength())
}

// HandleNewTasks handles new tasks from API,  这里应该是入口函数。主要做创建任务的逻辑
func HandleNewTasks(name string, taskType int, cronExpression string,
	format int, script string, callBackURL string, retries int) (string, error) {
	log.Printf("----------------------------------")
	// get update generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, format, script, retries)
	log.Printf("Handle New Task")
	// insert update to database
	log.Printf("Inserting task to database: %s", taskDB.ID)
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
	log.Printf("----------------------------------")
	log.Printf("Handle Delete Task")
	// if the update is running, stop it
	// check the running_tasks_record board to see the job is running or not
	record, err := databaseClient.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, taskID)
	if record == nil || err != nil {
		log.Printf("Task %s is not running, delete it from database, see err: %v", taskID, err)
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
		log.Printf("Task %s is running, wait it to finish the current turn, see err: %v", taskID, err)
		return fmt.Errorf("cannot delete the task when it is running")
	}
}

func HandleUpdateTasks(taskID string, args map[string]interface{}) error {
	// do nothing when the update is running, report the user cannot update the task when it is running
	// if the update is in redis priority queue, update it and update it in database
	data_structure_redis.RemoveJobByID(taskID)
	taskCache := new(constants.TaskCache)
	taskCache.ID = taskID
	taskCache.Payload = &constants.Payload{}
	for key, value := range args {
		switch key {
		case "job_type":
			taskCache.JobType = value.(int)
		case "execute_format":
			taskCache.Payload.Format = value.(int)
		case "execute_script":
			taskCache.Payload.Script = value.(string)
		case "retries":
			taskCache.RetriesLeft = value.(int)
		case "cron_expr":
			taskCache.CronExpr = value.(string)
			taskCache.ExecutionTime = generator.DecryptCronExpress(value.(string))
		}
	}
	// if the update is in database, update it in database
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, taskID, args)
	// database updated successfully then add job
	data_structure_redis.AddJob(taskCache)
	if err != nil {
		return fmt.Errorf("update task %s failed: %v", taskID, err)
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

func HandleGetRunningTasks() ([]*constants.RunTimeTask, error) {
	tasks, err := databaseClient.GetAllRunningTasks()
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
				err := executeMatureTasks()
				if err != nil {
					log.Printf("dispatch tasks failed: %v", err)
				}
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()

		}
	}
}

func executeMatureTasks() error {
	matureTasks := data_structure_redis.PopJobsForDispatchWithBuffer()
	for _, task := range matureTasks {
		log.Printf("dispatching task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])
		runningTaskInfo := generator.GenerateRunTimeTaskThroughTaskCache(task,
			constants.JOBDISPATCHED, "nil")

		var err error
		err = databaseClient.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD,
			runningTaskInfo)
		if err != nil {
			return err
		}
		// TODO: need to update the lease time, the current duration is 2 seconds
		err = data_structure_redis.SetLeaseWithID(task.ID, 2*time.Second)
		if err != nil {
			return err
		}
		// TODO: _ is the workerID, currently not used, will used for routing key in future
		workerID, status, err := HandoutTasksForExecuting(task)

		log.Printf("dispatch task: %s and executing by %s, found error: %v", task.ID, workerID, err)

		err = updateWithDispatchResult(task, status)
		if err != nil {
			return err
		}
	}
	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), constants.EXECUTE_TASK_GRPC_TIMEOUT)
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

	if ctxErr := ctx.Err(); ctxErr != nil {
		// grpc timeout
		if r != nil {
			return r.Id, int(r.Status), fmt.Errorf("grpc timeout: %v", ctxErr)
		}
		return "", constants.JOBFAILED, fmt.Errorf("grpc timeout: %v", ctxErr)
	}

	if err != nil {
		if r != nil {
			return r.Id, constants.JOBFAILED,
				fmt.Errorf("could not execute task: %s %v", task.ID, err)
		}
		return "", constants.JOBFAILED,
			fmt.Errorf("could not execute task: %s %v", task.ID, err)
	}
	log.Printf("Worker: %s executed task: %s with status %s", r.Id, task.ID, constants.StatusMap[int(r.Status)])
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
					log.Printf("dispatching task: %s through Monitor Trigger",
						taskCache.ID)
					err := executeMatureTasks()
					if err != nil {
						log.Printf("dispatch tasks failed: %v", err)
					}
				}
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()
		}
	}
}

func updateWithDispatchResult(task *constants.TaskCache, jobStatusCode int) error {
	log.Printf("------------------------")
	log.Printf("update task: %s with status: %s", task.ID, constants.StatusMap[jobStatusCode])
	switch jobStatusCode {
	case dispatch.JobDispatched:
		return handleJobDispatched(task, jobStatusCode)
	case dispatch.JobSucceed:
		return handleJobSucceed(task)
	case dispatch.JobFailed:
		return handleJobFailed(task, jobStatusCode)
	}
	return nil
}

func handleJobFailed(task *constants.TaskCache, jobStatusCode int) error {
	// update task status to 3
	// task_db only update when the update is completely failed (retries == 0 or outdated)
	if checkJobCompletelyFailed(task) {
		return handleCompletelyFailedJob(task, jobStatusCode)
	}
	return handleRescheduleFailedJob(task)
}

func handleRescheduleFailedJob(task *constants.TaskCache) error {
	// update not completely failed, update the retries left, time
	// and add to redis priority queue
	err := data_structure_redis.RemoveLeaseWithID(task.ID)
	if err != nil {
		return err
	}

	err = RescheduleFailedJobs(task)
	if err != nil {
		return err
	}
	return nil
}

func handleCompletelyFailedJob(task *constants.TaskCache, jobStatusCode int) error {
	// update completely failed need to update all information for report user.
	err := databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
	if err != nil {
		return err
	}
	err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
		map[string]interface{}{"job_status": jobStatusCode})
	if err != nil {
		return err
	}
	err = data_structure_redis.RemoveLeaseWithID(task.ID)
	if err != nil {
		return err
	}
	return nil
}

func handleJobSucceed(task *constants.TaskCache) error {
	log.Printf("---------------------------------")
	log.Printf("handle job succeed for job: %s for job type: %s", task.ID,
		constants.TypeMap[task.JobType])
	// the task should be deleted from ExecutionRecord
	err := databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID)
	if err != nil {
		return err
	}
	if task.JobType == constants.Recurring {
		return handleRecurringJobSucceed(task)
	}
	return handleOneTimeJobSucceed(task)
}

func handleOneTimeJobSucceed(task *constants.TaskCache) error {
	// if it is one time job, only update the task_db table
	updateVars := map[string]interface{}{
		"update_time":             time.Now(),
		"status":                  constants.JOBSUCCEED,
		"previous_execution_time": task.ExecutionTime,
	}
	err := databaseClient.UpdateByID(context.Background(),
		constants.TASKS_FULL_RECORD, task.ID, updateVars)
	if err != nil {
		return err
	}
	err = data_structure_redis.RemoveLeaseWithID(task.ID)
	if err != nil {
		return err
	}
	return nil
}

func handleRecurringJobSucceed(task *constants.TaskCache) error {
	// it is recur job, update execution time, update time, status, previous execution time
	cronExpr := task.CronExpr
	newExecutionTime := generator.DecryptCronExpress(cronExpr)
	log.Printf("old time: %v, new time: %v", task.ExecutionTime, newExecutionTime)
	updateVars := map[string]interface{}{
		"execution_time":          newExecutionTime,
		"update_time":             time.Now(),
		"status":                  constants.JOBREENTERED,
		"previous_execution_time": task.ExecutionTime,
	}
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
	if err != nil {
		return err
	}
	task.ExecutionTime = newExecutionTime
	err = data_structure_redis.RemoveLeaseWithID(task.ID)
	if err != nil {
		return err
	}
	if data_structure_redis.CheckTasksInDuration(newExecutionTime, DURATION) {
		data_structure_redis.AddJob(task)
	}
	return nil
}

func handleJobDispatched(task *constants.TaskCache, jobStatusCode int) error {
	// update task status to 1
	// happens once the job is dispatched to worker
	// only update the task_db table in postgresql
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
		map[string]interface{}{"status": jobStatusCode})
	if err != nil {
		return err
	}
	return nil
}

// RescheduleFailedJobs reenter the update to queue and update the retries
func RescheduleFailedJobs(task *constants.TaskCache) error {
	curRetries := task.RetriesLeft - 1
	err := databaseClient.UpdateByID(context.Background(), constants.RUNNING_JOBS_RECORD, task.ID,
		map[string]interface{}{"job_status": constants.JOBRETRYING, "retries_left": curRetries, "worker_id": ""})
	if err != nil {
		return err
	}

	err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID,
		map[string]interface{}{
			"retries_left":            curRetries,
			"previous_execution_time": task.ExecutionTime,
			"update_time":             time.Now(),
			"status":                  constants.JOBRETRYING})
	if err != nil {
		return err
	}
	task.RetriesLeft = curRetries
	data_structure_redis.AddJob(task)
	return nil
}

func checkJobCompletelyFailed(data interface{}) bool {
	jobType, retriesLeft, executionTime := getJobInfo(data)
	if jobType == constants.OneTime {
		if checkOneTimeJobOutdated(executionTime) {
			// update completely failed need to update all information for report user.
			return true
		} else {
			return retriesLeft <= 0
		}
	} else {
		if checkRecurringJobOutdated(executionTime) {
			return true
		} else {
			return retriesLeft <= 0
		}
	}
}

// check if the update is outdated, current time is later than the execution time
func checkOneTimeJobOutdated(executionTime time.Time) bool {
	diff := time.Now().Sub(executionTime)
	return diff > constants.ONE_TIME_JOB_RETRY_TIME
}

func checkRecurringJobOutdated(executionTime time.Time) bool {
	diff := time.Now().Sub(executionTime)
	return diff > constants.RECURRING_JOB_RETRY_TIME
}

// HandleExpiryTasks handles the expiry tasks
func HandleExpiryTasks(msg *redis.Message) {
	// The key format is "lease:update:<task_id>"
	if strings.HasPrefix(msg.Payload, "lease:update:") {
		taskID := strings.TrimPrefix(msg.Payload, "lease:update:")
		log.Printf("Task %s is expired", taskID)
		err := HandleUnRenewLeaseJobThroughDB(taskID)
		if err != nil {
			log.Printf("Lease renew failed: %v", err)
		}
	}
}

func HandleUnRenewLeaseJobThroughDB(id string) error {
	log.Printf("-----------------------------")
	log.Printf("Handle un renew lease job: %s", id)
	// check job type (through runtime db table)
	// if job type is one time, directly return fail
	// if job type is recur, check retry times
	// if > 0, reenter queue, update redis, db; otherwise return fail
	record, err := databaseClient.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, id)
	if err != nil {
		log.Printf("Get task by id failed when handle renew lease jobs: %v", err)
		return err
	}
	runTimeTask := record.(*constants.RunTimeTask)
	if checkJobCompletelyFailed(runTimeTask) {
		// update completely failed need to update all information for report user.
		err := databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, runTimeTask.ID)
		if err != nil {
			return err
		}
		err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
			runTimeTask.ID, map[string]interface{}{"job_status": constants.JOBFAILED})
		if err != nil {
			return err
		}
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
		err := RescheduleFailedJobs(task)
		if err != nil {
			return err
		}
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

func (s *ServerImpl) NotifyTaskStatus(ctx context.Context,
	in *pb.NotifyMessageRequest) (*pb.NotifyMessageResponse, error) {
	log.Printf("Received Task: %s with Status %s from worker %s",
		in.TaskId, constants.StatusMap[int(in.Status)], in.WorkerId)

	record, err := databaseClient.GetTaskByID(ctx, constants.RUNNING_JOBS_RECORD, in.TaskId)
	if err != nil {
		return &pb.NotifyMessageResponse{Success: false},
			fmt.Errorf("failed to get task by id from DB: %v", err)
	}

	runningTask := record.(*constants.RunTimeTask)

	log.Printf("Running Task: %v", runningTask)

	payload := generator.GeneratePayload(runningTask.Payload.Format, runningTask.Payload.Script)
	taskCache := generator.GenerateTaskCache(
		runningTask.ID, runningTask.JobType, runningTask.CronExpr,
		runningTask.ExecutionTime, runningTask.RetriesLeft, payload)

	err = updateWithDispatchResult(taskCache, int(in.Status))
	if err != nil {
		return &pb.NotifyMessageResponse{Success: false},
			fmt.Errorf("failed to get response back to worker: %v", err)
	}

	return &pb.NotifyMessageResponse{Success: true}, nil
}

func (s *ServerImpl) RenewLease(ctx context.Context, in *pb.RenewLeaseRequest) (*pb.RenewLeaseResponse, error) {
	// worker should ask for a new lease before the current lease expires
	err := data_structure_redis.SetLeaseWithID(in.Id, in.LeaseDuration.AsDuration())
	if err != nil {
		return &pb.RenewLeaseResponse{Success: false},
			fmt.Errorf("lease for update %s has expired", in.Id)
	}
	return &pb.RenewLeaseResponse{Success: true}, nil
}
