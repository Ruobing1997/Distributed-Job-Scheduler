package task_manager

import (
	"context"
	"fmt"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/postgreSQL"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
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
	logger               = logrus.New()
)

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

type ServerControlInterface interface {
	StartAPIServer()
	StopAPIServer() error
}

func InitConnection() {
	timeTracker = time.Now()
	logFile, err := os.OpenFile("./logs/managers/manager-"+managerID+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "InitConnection",
			"result":   "failed to open log file",
			"error":    err,
		}).Fatal("failed to open log file")
	}
	multiWrite := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(multiWrite)
	logger.SetFormatter(&logrus.JSONFormatter{})
	// TODO: make the user to choose database type
	databaseClient = postgreSQL.NewpostgreSQLClient()
	redisClient = data_structure_redis.Init()
}

func InitManagerGRPC() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "InitManagerGRPC",
			"result":   "failed to listen",
			"error":    err,
		}).Fatal("failed to listen")
	}
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &ServerImpl{})
	pb.RegisterLeaseServiceServer(s, &ServerImpl{})
	if err := s.Serve(lis); err != nil {
		logger.WithFields(logrus.Fields{
			"function": "InitManagerGRPC",
			"result":   "failed to serve",
			"error":    err,
		}).Fatal("failed to serve")
	}
}

func InitLeaderElection(control ServerControlInterface) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "InitLeaderElection",
			"result":   "failed to get k8s config",
			"error":    err,
		}).Fatal("failed to get k8s config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "InitLeaderElection",
			"result":   "failed to get k8s client",
			"error":    err,
		}).Fatal("failed to get k8s client")
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "manager-lock",
			Namespace: "supernova",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: managerID,
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   30 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				logger.WithFields(logrus.Fields{
					"function": "InitLeaderElection",
				}).Info(fmt.Sprintf("manager: %v started leading", managerID))
				InitConnection()
				PrometheusManagerInit()
				go InitManagerGRPC()
				Start()
				control.StartAPIServer()
			},
			OnStoppedLeading: func() {
				logger.WithFields(logrus.Fields{
					"function": "OnStoppedLeading",
					"manager":  managerID,
				}).Info("Manager stopped leading")

				if err := databaseClient.Close(); err != nil {
					logger.WithFields(logrus.Fields{
						"function": "OnStoppedLeading",
						"error":    err,
					}).Fatal("Error closing database client")
				}

				if err := redisClient.Close(); err != nil {
					logger.WithFields(logrus.Fields{
						"function": "OnStoppedLeading",
						"error":    err,
					}).Fatal("Error closing redis client")
				}

				if err := control.StopAPIServer(); err != nil {
					logger.WithFields(logrus.Fields{
						"function": "OnStoppedLeading",
						"error":    err,
					}).Fatal("Error stopping API server")
				}
			},
			OnNewLeader: func(identity string) {
				if identity != managerID {
					logger.WithFields(logrus.Fields{
						"function": "OnNewLeader",
						"leader":   identity,
					}).Info("New leader elected")
				}
			},
		},
	})
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

func addNewJobToRedis(taskDB *constants.TaskDB) {
	// generate update for cache:
	taskCache := generator.GenerateTaskCache(
		taskDB.ID,
		taskDB.JobType,
		taskDB.CronExpr,
		taskDB.ExecutionTime,
		taskDB.Retries,
		taskDB.Payload)
	taskCache.ExecutionID = generator.GenerateExecutionID()
	redisThroughput.Inc()
	data_structure_redis.AddJob(taskCache)
	logger.WithFields(logrus.Fields{
		"function": "addNewJobToRedis",
		"taskID":   taskCache.ID,
		"QLen":     data_structure_redis.GetQLength(),
	}).Info("added new job to redis")
}

// HandleNewTasks handles new tasks from API,  这里应该是入口函数。主要做创建任务的逻辑
func HandleNewTasks(name string, taskType int, cronExpression string,
	format int, script string, callBackURL string, retries int) (string, error) {
	tasksTotal.Inc()
	// get update generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, format, script, retries)
	logger.WithFields(logrus.Fields{
		"function":    "HandleNewTasks",
		"task_detail": taskDB,
	}).Info("new task created")
	// insert update to database
	err := databaseClient.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, taskDB)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "INSERT", "table": constants.TASKS_FULL_RECORD}).Inc()
	if err != nil {
		return "nil", err
	}

	redisThroughput.Inc()
	if data_structure_redis.CheckTasksInDuration(taskDB.ExecutionTime, DURATION) {
		// insert update to priority queue
		addNewJobToRedis(taskDB)
	}
	return taskDB.ID, nil
}

func HandleDeleteTasks(taskID string) error {
	logger.WithFields(logrus.Fields{
		"function": "HandleDeleteTasks",
		"taskID":   taskID,
	}).Info("delete task")
	tasksTotal.Dec()
	// if the update is running, stop it
	// check the running_tasks_record board to see the job is running or not
	record, err := databaseClient.CountRunningTasks(context.Background(), taskID)
	logger.WithFields(logrus.Fields{
		"function": "HandleDeleteTasks",
		"error":    err,
	}).Fatal("Error counting running tasks")
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()
	if record == 0 {
		logger.WithFields(logrus.Fields{
			"function": "HandleDeleteTasks",
			"taskID":   taskID,
		}).Info("task is not running, delete it from database")
		// apply the same logic as the following
		// if the update is in redis priority queue, delete it and delete it from database
		// if the update is in database, delete it from database
		data_structure_redis.RemoveJobByID(taskID)
		redisThroughput.Inc()
		err := databaseClient.DeleteByID(context.Background(), constants.TASKS_FULL_RECORD, taskID)
		err = databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, taskID)
		postgresqlOpsTotal.With(prometheus.Labels{"operation": "DELETE", "table": constants.TASKS_FULL_RECORD}).Inc()
		postgresqlOpsTotal.With(prometheus.Labels{"operation": "DELETE", "table": constants.RUNNING_JOBS_RECORD}).Inc()
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		logger.WithFields(logrus.Fields{
			"function": "HandleDeleteTasks",
			"taskID":   taskID,
		}).Info("task is running, wait it to finish the current turn")
		return fmt.Errorf("cannot delete the task when it is running")
	}
}

func HandleUpdateTasks(taskID string, args map[string]interface{}) error {
	logger.WithFields(logrus.Fields{
		"function": "HandleUpdateTasks",
		"taskID":   taskID,
	}).Info("update task")
	// if the task is in redis priority queue, update it and update it in database
	data_structure_redis.RemoveJobByID(taskID)
	redisThroughput.Inc()
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
		case "retries_left":
			taskCache.RetriesLeft = value.(int)
		case "cron_expression":
			taskCache.CronExpr = value.(string)
			taskCache.ExecutionTime = generator.DecryptCronExpress(value.(string))
		}
	}
	// if the task is in database, update it in database
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, taskID, args)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "HandleUpdateTasks",
			"taskID":   taskID,
		}).Errorf("update task %s failed: %v", taskID, err)
		return fmt.Errorf("update task %s failed: %v", taskID, err)
	}
	// database updated successfully then add job
	if data_structure_redis.CheckTasksInDuration(taskCache.ExecutionTime, DURATION) {
		// insert update to priority queue
		// 更新完按照新任务处理
		taskCache.ExecutionID = generator.GenerateExecutionID()
		redisThroughput.Inc()
		logger.WithFields(logrus.Fields{
			"function": "HandleUpdateTasks",
			"taskID":   taskID,
		}).Infof("Task in 10 mins, add to redis: %v", taskCache)
		data_structure_redis.AddJob(taskCache)
	}
	return nil
}

func HandleGetTasks(taskID string) (*constants.TaskDB, error) {
	logger.WithFields(logrus.Fields{
		"function": "HandleGetTasks",
		"taskID":   taskID,
	}).Info("get task")
	record, err := databaseClient.GetTaskByID(context.Background(), constants.TASKS_FULL_RECORD, taskID)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.TASKS_FULL_RECORD}).Inc()
	if err != nil {
		return nil, err
	}
	return record.(*constants.TaskDB), nil
}

func HandleGetAllTasks() ([]*constants.TaskDB, error) {
	logger.WithFields(logrus.Fields{
		"function": "HandleGetAllTasks",
	}).Info("get all tasks")
	tasks, err := databaseClient.GetAllTasks()
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.TASKS_FULL_RECORD}).Inc()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func HandleGetRunningTasks() ([]*constants.RunTimeTask, error) {
	logger.WithFields(logrus.Fields{
		"function": "HandleGetRunningTasks",
	}).Info("get all running tasks")
	tasks, err := databaseClient.GetAllRunningTasks()
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func HandleGetTaskHistory(taskID string) ([]*constants.RunTimeTask, error) {
	logger.WithFields(logrus.Fields{
		"function": "HandleGetTaskHistory",
	}).Info("get task history")
	records, err := databaseClient.GetTaskHistoryByID(context.Background(), taskID)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func AddTasksFromDBWithTickers() {
	logger.WithFields(logrus.Fields{
		"function": "AddTasksFromDBWithTickers",
	}).Info("add tasks from database with tickers")
	tasks, _ := databaseClient.GetTasksInInterval(time.Now(), time.Now().Add(DURATION), timeTracker)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.TASKS_FULL_RECORD}).Inc()
	timeTracker = time.Now()
	for _, task := range tasks {
		addNewJobToRedis(task)
	}
}

func SubscribeToRedisChannel() {
	logger.WithFields(logrus.Fields{
		"function": "SubscribeToRedisChannel",
	}).Info("subscribe to redis channel")
	pubsub := redisClient.Subscribe(context.Background(), data_structure_redis.REDIS_CHANNEL)
	_, err := pubsub.Receive(context.Background())
	redisThroughput.Inc()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "SubscribeToRedisChannel",
		}).Errorf("redis subscribe failed: %v", err)
	}
	ch := pubsub.Channel()
	for msg := range ch {
		if msg.Payload == data_structure_redis.TASK_AVAILABLE {
			taskProcessingLock.Lock()
			if !isTaskBeingProcessed {
				isTaskBeingProcessed = true
				// check redis priority queue and dispatch tasks
				err := executeMatureTasks()
				if err != nil {
					logger.WithFields(logrus.Fields{
						"function": "SubscribeToRedisChannel",
					}).Errorf("dispatch tasks failed: %v", err)
				}
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()

		} else if msg.Payload == data_structure_redis.RETRY_AVAILABLE {
			for {
				retryTask, err := data_structure_redis.PopRetry()
				redisThroughput.Inc()
				if err != nil {
					break
				}
				logger.WithFields(logrus.Fields{
					"function": "SubscribeToRedisChannel",
				}).Infof("retrying task: %s with job type %s", retryTask.ID, constants.TypeMap[retryTask.JobType])
				_, status, _ := HandoutTasksForExecuting(retryTask)
				if status == constants.JOBFAILED {
					handleJobFailed(retryTask, status)
				}
			}
		}
	}
}

func executeMatureTasks() error {
	dispatchTotal.Inc()
	matureTasks := data_structure_redis.PopJobsForDispatchWithBuffer()
	redisThroughput.Inc()
	for _, task := range matureTasks {
		logger.WithFields(logrus.Fields{
			"function": "executeMatureTasks",
		}).Infof("dispatching task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])
		runningTaskInfo := generator.GenerateRunTimeTaskThroughTaskCache(task,
			constants.JOBDISPATCHED, "nil")

		var err error
		err = databaseClient.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD,
			runningTaskInfo)

		postgresqlOpsTotal.With(prometheus.Labels{"operation": "INSERT", "table": constants.RUNNING_JOBS_RECORD}).Inc()

		if err != nil {
			return err
		}

		err = data_structure_redis.SetLeaseWithID(task.ID, task.ExecutionID, 2*time.Second)
		redisThroughput.Inc()
		if err != nil {
			return err
		}

		// TODO: 可能需要放在一个goroutine里面
		workerID, status, err := HandoutTasksForExecuting(task)
		grpcOpsTotal.With(prometheus.Labels{"operation": "EXECUTE_TASK", "sender": managerID}).Inc()

		logger.WithFields(logrus.Fields{
			"function": "executeMatureTasks",
		}).Infof("dispatch task: %s and executing by %s, found error: %v with status %s",
			task.ID, workerID, err, constants.StatusMap[status])

		if status == constants.JOBFAILED {
			handleJobFailed(task, status)
		}

		// check if the job is recurring job, if so, update the execution time
		//and add it to redis priority queue if in next 10 mins
		go handleRecurringJobAfterDispatch(task)
	}
	return nil
}

func handleRecurringJobAfterDispatch(task *constants.TaskCache) {
	logger.WithFields(logrus.Fields{
		"function": "handleRecurringJobAfterDispatch",
	}).Infof("handle recurring job after dispatching task: %s", task.ID)
	if task.JobType == constants.Recurring {
		cronExpr := task.CronExpr
		newExecutionTime := generator.DecryptCronExpress(cronExpr)
		task.ExecutionTime = newExecutionTime
		logger.WithFields(logrus.Fields{
			"function": "handleRecurringJobAfterDispatch",
		}).Infof("New Time Set %s for task: %s", newExecutionTime, task.ID)
		if data_structure_redis.CheckTasksInDuration(newExecutionTime, DURATION) {
			// Re dispatch the task
			task.ExecutionID = generator.GenerateExecutionID()
			redisThroughput.Inc()
			data_structure_redis.AddJob(task)
		}
	}
}

func HandoutTasksForExecuting(task *constants.TaskCache) (string, int, error) {
	logger.WithFields(logrus.Fields{
		"function": "HandoutTasksForExecuting",
	}).Infof("handout task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])

	workerService := os.Getenv("WORKER_SERVICE")
	if workerService == "" {
		workerService = WORKER_SERVICE
	}
	conn, err := grpc.Dial(workerService+":50051", grpc.WithInsecure(), grpc.WithBlock())
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
		ExecId:        task.ExecutionID,
	})

	if ctxErr := ctx.Err(); ctxErr != nil {
		logger.WithFields(logrus.Fields{
			"function": "HandoutTasksForExecuting",
			"error":    ctxErr,
		}).Info("grpc timeout")
		// grpc timeout
		if r != nil {
			return r.Id, int(r.Status), fmt.Errorf("grpc timeout: %v", ctxErr)
		}
		return "", constants.JOBFAILED, fmt.Errorf("grpc timeout: %v", ctxErr)
	}

	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "HandoutTasksForExecuting",
			"error":    err,
		}).Infof("dispatched task: %s failed: %v", task.ID, err)
		if r != nil {
			return r.Id, constants.JOBFAILED,
				fmt.Errorf("dispatched task: %s failed: %v", task.ID, err)
		}
		return "", constants.JOBFAILED,
			fmt.Errorf("dispatched task: %s failed: %v", task.ID, err)
	}

	logger.WithFields(logrus.Fields{
		"function": "HandoutTasksForExecuting",
	}).Infof("Worker: %s executed task: %s with status %s",
		r.Id, task.ID, constants.StatusMap[int(r.Status)])
	return r.Id, int(r.Status), nil
}

func MonitorHeadNode() {
	logger.WithFields(logrus.Fields{
		"function": "MonitorHeadNode",
	}).Info("Start monitoring head node")
	headNodeTicker := time.NewTicker(1 * time.Second)
	defer headNodeTicker.Stop()
	for {
		select {
		case <-headNodeTicker.C:
			taskProcessingLock.Lock()
			if !isTaskBeingProcessed {
				isTaskBeingProcessed = true
				taskCache := data_structure_redis.GetNextJob()
				redisThroughput.Inc()
				if taskCache != nil && data_structure_redis.CheckWithinThreshold(taskCache.ExecutionTime) {
					logger.WithFields(logrus.Fields{
						"function": "MonitorHeadNode",
					}).Infof("dispatch task: %s", taskCache.ID)
					err := executeMatureTasks()
					if err != nil {
						logger.WithFields(logrus.Fields{
							"function": "MonitorHeadNode",
							"error":    err,
						}).Fatal("execute mature tasks failed")
					}
				}
				isTaskBeingProcessed = false
			}
			taskProcessingLock.Unlock()
		}
	}
}

func handleJobFailed(task *constants.TaskCache, jobStatusCode int) error {
	logger.WithFields(logrus.Fields{
		"function": "handleJobFailed",
	}).Infof("handle job failed: %s Exec ID: %s", task.ID, task.ExecutionID)
	// update task status to 3
	// task_db only update when the update is completely failed (retries == 0 or outdated)
	if checkJobCompletelyFailed(task) {
		failedTasksTotal.Inc()
		return handleCompletelyFailedJob(task, jobStatusCode)
	}
	return handleRescheduleFailedJob(task)
}

func handleRescheduleFailedJob(task *constants.TaskCache) error {
	// update not completely failed, update the retries left, time
	// and add to redis priority queue
	logger.WithFields(logrus.Fields{
		"function": "handleRescheduleFailedJob",
	}).Infof("Eeschedule failed job: %s Exec ID: %s", task.ID, task.ExecutionID)

	err := data_structure_redis.RemoveLeaseWithID(task.ID, task.ExecutionID)
	redisThroughput.Inc()
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
	logger.WithFields(logrus.Fields{
		"function": "handleCompletelyFailedJob",
	}).Infof("Completely failed job: %s Exec ID: %s", task.ID, task.ExecutionID)

	// update completely failed need to update all information for report user.
	err := data_structure_redis.RemoveLeaseWithID(task.ID, task.ExecutionID)
	redisThroughput.Inc()
	// update the task according to task execution id.
	err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
		task.ID, map[string]interface{}{
			"job_status":              jobStatusCode,
			"update_time":             time.Now(),
			"previous_execution_time": task.ExecutionTime,
		},
	)

	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()

	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

// RescheduleFailedJobs reenter the task to queue and update the retries
func RescheduleFailedJobs(task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "RescheduleFailedJobs",
	}).Infof("Reschedule failed job: %s initial retries left: %d", task.ID, task.RetriesLeft)
	curRetries := task.RetriesLeft - 1
	err := databaseClient.UpdateByExecutionID(context.Background(), constants.RUNNING_JOBS_RECORD,
		task.ExecutionID, map[string]interface{}{
			"job_status":   constants.JOBRETRYING,
			"retries_left": curRetries,
			"worker_id":    ""})
	if err != nil {
		return err
	}

	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.RUNNING_JOBS_RECORD}).Inc()

	task.RetriesLeft = curRetries
	// 此时执行ID不应该发生变化，因为是同一个job的重试逻辑
	data_structure_redis.AddRetry(task)
	logger.WithFields(logrus.Fields{
		"function": "RescheduleFailedJobs",
	}).Infof("Add failed job: %s to retry list,"+
		" new retries left: %d ", task.ID, task.RetriesLeft)
	redisThroughput.Inc()
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
	// The expected key format is "lease:task:<taskID>execute:<execID>"
	prefix := "lease:task:"
	suffix := "execute:"

	if strings.HasPrefix(msg.Payload, prefix) {
		trimmedStr := strings.TrimPrefix(msg.Payload, prefix) // removing "lease:task:" prefix
		splitIndex := strings.Index(trimmedStr, suffix)

		if splitIndex == -1 {
			logger.WithFields(logrus.Fields{
				"function": "HandleExpiryTasks",
			}).Infof("Invalid message format received: %s", msg.Payload)
			return
		}

		execID := strings.TrimPrefix(trimmedStr[splitIndex:], suffix)

		err := HandleUnRenewLeaseJobThroughDB(execID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "HandleExpiryTasks",
			}).Errorf("Error handling expiry task: %s", err.Error())
		}
	}
}

func HandleUnRenewLeaseJobThroughDB(id string) error {
	logger.WithFields(logrus.Fields{
		"function": "HandleUnRenewLeaseJobThroughDB",
	}).Infof("Handle un renew lease job exec id: %s", id)
	// check job type (through runtime db table)
	// if job type is one time, directly return fail
	// if job type is recur, check retry times
	// if > 0, reenter queue, update redis, db; otherwise return fail
	record, err := databaseClient.GetTaskByID(context.Background(), constants.RUNNING_JOBS_RECORD, id)

	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()

	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "HandleUnRenewLeaseJobThroughDB",
		}).Errorf("Get task by exec id: %v failed", err)
		return err
	}
	runTimeTask := record.(*constants.RunTimeTask)
	if checkJobCompletelyFailed(runTimeTask) {
		logger.WithFields(logrus.Fields{
			"function": "HandleUnRenewLeaseJobThroughDB",
		}).Infof("Job completely failed: %s", runTimeTask.ID)

		failedTasksTotal.Inc()
		err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
			runTimeTask.ID, map[string]interface{}{"job_status": constants.JOBFAILED})

		postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()

		if err != nil {
			return err
		}
	} else {
		logger.WithFields(logrus.Fields{
			"function": "HandleUnRenewLeaseJobThroughDB",
		}).Infof("Reschedule since job: %s NOT completely failed", runTimeTask.ID)
		// update not completely failed, update the retries left
		task := generator.GenerateTaskCache(
			runTimeTask.ID,
			runTimeTask.JobType,
			runTimeTask.CronExpr,
			runTimeTask.ExecutionTime,
			runTimeTask.RetriesLeft,
			runTimeTask.Payload,
		)
		task.ExecutionID = runTimeTask.ExecutionID
		err := RescheduleFailedJobs(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func getJobInfo(data interface{}) (int, int, time.Time) {
	switch v := data.(type) {
	case *constants.TaskCache:
		return v.JobType, v.RetriesLeft, v.ExecutionTime
	case *constants.RunTimeTask:
		return v.JobType, v.RetriesLeft, v.ExecutionTime
	default:
		panic("unsupported data type")
	}
}

func ListenForExpiryNotifications() {
	logger.WithFields(logrus.Fields{
		"function": "ListenForExpiryNotifications",
	}).Info("Start listening for expiry notifications")
	redisThroughput.Inc()
	pubsub := redisClient.Subscribe(context.Background(), REDIS_EXPIRY_CHANNEL)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "ListenForExpiryNotifications",
			}).Errorf("Receive message failed: %v", err)
			continue
		}
		HandleExpiryTasks(msg)
	}
}

func (s *ServerImpl) NotifyTaskStatus(ctx context.Context,
	in *pb.NotifyMessageRequest) (*pb.NotifyMessageResponse, error) {
	logger.WithFields(logrus.Fields{
		"function": "NotifyTaskStatus",
	}).Infof("Received Task: %s with Status %s from worker %s with exec id: %s",
		in.TaskId, constants.StatusMap[int(in.Status)], in.WorkerId, in.ExecId)

	err := data_structure_redis.RemoveLeaseWithID(in.TaskId, in.ExecId)
	redisThroughput.Inc()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "NotifyTaskStatus",
		}).Errorf("Remove lease failed: %v", err)
	}

	record, err := databaseClient.GetTaskByID(ctx, constants.RUNNING_JOBS_RECORD, in.ExecId)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()

	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "NotifyTaskStatus",
		}).Errorf("Get task by id: %s failed: %v", in.ExecId, err)
		return &pb.NotifyMessageResponse{Success: false},
			fmt.Errorf("id %s not in DB: %v", in.ExecId, err)
	}

	runningTask := record.(*constants.RunTimeTask)

	logger.WithFields(logrus.Fields{
		"function": "NotifyTaskStatus",
	}).Infof("Running Task Got in DB: %v", runningTask)

	payload := generator.GeneratePayload(runningTask.Payload.Format, runningTask.Payload.Script)
	taskCache := generator.GenerateTaskCache(
		runningTask.ID, runningTask.JobType, runningTask.CronExpr,
		runningTask.ExecutionTime, runningTask.RetriesLeft, payload)
	taskCache.ExecutionID = in.ExecId

	if in.Status == int32(constants.JOBSUCCEED) {
		logger.WithFields(logrus.Fields{
			"function": "NotifyTaskStatus",
		}).Infof("Job SUCCEED with task id: %s and exeution id: %s from worker: %s",
			in.TaskId, in.ExecId, in.WorkerId)
		succeedTasksTotal.Inc()
		handleJobSucceed(taskCache)
	} else {
		logger.WithFields(logrus.Fields{
			"function": "NotifyTaskStatus",
		}).Infof("Job FAILED with task id: %s and exeution id: %s from worker: %s",
			in.TaskId, in.ExecId, in.WorkerId)
		handleJobFailed(taskCache, constants.JOBFAILED)
	}

	return &pb.NotifyMessageResponse{Success: true}, nil
}

func handleJobSucceed(task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleJobSucceed",
	}).Infof("handle job succeed for job: %s for job type: %s", task.ID, constants.TypeMap[task.JobType])
	if task.JobType == constants.Recurring {
		return handleRecurringJobSucceed(task)
	}
	return handleOneTimeJobSucceed(task)
}

func handleOneTimeJobSucceed(task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleOneTimeJobSucceed",
	}).Infof("task: %s is one time job, update task_db table and remove lease", task.ID)
	// if it is one time job, only update the task_db table
	updateVars := map[string]interface{}{
		"update_time":             time.Now(),
		"job_status":              constants.JOBSUCCEED,
		"previous_execution_time": task.ExecutionTime,
	}
	err := databaseClient.UpdateByID(context.Background(),
		constants.TASKS_FULL_RECORD, task.ID, updateVars)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()

	if err != nil {
		return err
	}
	err = data_structure_redis.RemoveLeaseWithID(task.ID, task.ExecutionID)
	redisThroughput.Inc()
	if err != nil {
		return err
	}
	return nil
}

func handleRecurringJobSucceed(task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleRecurringJobSucceed",
	}).Infof("task: %s is recurring job, update task_db table and remove lease", task.ID)
	// it is recur job, update execution time, update time, status, previous execution time
	cronExpr := task.CronExpr
	newExecutionTime := generator.DecryptCronExpress(cronExpr)
	updateVars := map[string]interface{}{
		"execution_time":          newExecutionTime,
		"update_time":             time.Now(),
		"job_status":              constants.JOBREENTERED,
		"previous_execution_time": task.ExecutionTime,
	}
	err := databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
	err = data_structure_redis.RemoveLeaseWithID(task.ID, task.ExecutionID)

	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()

	if err != nil {
		return err
	}
	return nil
}

func (s *ServerImpl) RenewLease(ctx context.Context, in *pb.RenewLeaseRequest) (*pb.RenewLeaseResponse, error) {
	// worker should ask for a new lease before the current lease expires
	err := data_structure_redis.SetLeaseWithID(in.Id, in.ExecId,
		in.LeaseDuration.AsDuration())
	redisThroughput.Inc()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "RenewLease",
		}).Errorf("Renew Lease failed %v", err)
		return &pb.RenewLeaseResponse{Success: false},
			fmt.Errorf("lease for update %s has expired", in.Id)
	}
	logger.WithFields(logrus.Fields{
		"function": "RenewLease",
	}).Infof("Renew Lease Request fraom Task: %s with execID: %s SUCCEED", in.Id, in.ExecId)
	return &pb.RenewLeaseResponse{Success: true}, nil
}
