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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	timeTracker    time.Time
	databaseClient *postgreSQL.Client
	redisClient    *redis.Client
	managerID      = os.Getenv("HOSTNAME")
	logger         = logrus.New()
	connPool       []clientWrapper
	poolMutex      sync.Mutex
)

type clientWrapper struct {
	Conn   *grpc.ClientConn
	Client pb.TaskServiceClient
}

type handoutResultStruct struct {
	workerID string
	status   int
	err      error
}

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

type ServerControlInterface interface {
	StartAPIServer()
	StopAPIServer() error
}

func InitConnection() {
	timeTracker = time.Now().UTC()
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

func getGRPCConnection(workerService string) (*clientWrapper, error) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	if len(connPool) > 0 {
		clientConnWrapper := connPool[0]
		connPool = connPool[1:]
		return &clientConnWrapper, nil
	}

	conn, err := grpc.Dial(workerService+":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	client := pb.NewTaskServiceClient(conn)
	return &clientWrapper{Conn: conn, Client: client}, nil
}

func releaseGRPCConnection(wrapper *clientWrapper) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	connPool = append(connPool, *wrapper)
}

//
//func InitLeaderElection(control ServerControlInterface) {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		logger.WithFields(logrus.Fields{
//			"function": "InitLeaderElection",
//			"result":   "failed to get k8s config",
//			"error":    err,
//		}).Fatal("failed to get k8s config")
//	}
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		logger.WithFields(logrus.Fields{
//			"function": "InitLeaderElection",
//			"result":   "failed to get k8s client",
//			"error":    err,
//		}).Fatal("failed to get k8s client")
//	}
//
//	lock := &resourcelock.LeaseLock{
//		LeaseMeta: metav1.ObjectMeta{
//			Name:      "manager-lock",
//			Namespace: "supernova",
//		},
//		Client: clientset.CoordinationV1(),
//		LockConfig: resourcelock.ResourceLockConfig{
//			Identity: managerID,
//		},
//	}
//
//	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
//		Lock:            lock,
//		ReleaseOnCancel: true,
//		LeaseDuration:   60 * time.Second,
//		RenewDeadline:   30 * time.Second,
//		RetryPeriod:     5 * time.Second,
//		Callbacks: leaderelection.LeaderCallbacks{
//			OnStartedLeading: func(ctx context.Context) {
//				logger.WithFields(logrus.Fields{
//					"function": "InitLeaderElection",
//					"manager":  managerID,
//					"PodIP":    os.Getenv("POD_IP"),
//				}).Info(fmt.Sprintf("manager: %v started leading", managerID))
//				updateEndpointsWithLeaderIP(os.Getenv("POD_IP"))
//				InitConnection()
//				PrometheusManagerInit()
//				go InitManagerGRPC()
//				Start()
//				control.StartAPIServer()
//			},
//			OnStoppedLeading: func() {
//				logger.WithFields(logrus.Fields{
//					"function": "OnStoppedLeading",
//					"manager":  managerID,
//				}).Info("Manager stopped leading")
//
//				if err := databaseClient.Close(); err != nil {
//					logger.WithFields(logrus.Fields{
//						"function": "OnStoppedLeading",
//						"error":    err,
//					}).Error("Error closing database client")
//				}
//
//				if err := redisClient.Close(); err != nil {
//					logger.WithFields(logrus.Fields{
//						"function": "OnStoppedLeading",
//						"error":    err,
//					}).Error("Error closing redis client")
//				}
//
//				if err := control.StopAPIServer(); err != nil {
//					logger.WithFields(logrus.Fields{
//						"function": "OnStoppedLeading",
//						"error":    err,
//					}).Error("Error stopping API server")
//				}
//			},
//			OnNewLeader: func(identity string) {
//				if identity != managerID {
//					logger.WithFields(logrus.Fields{
//						"function": "OnNewLeader",
//						"leader":   identity,
//					}).Info("New leader elected")
//				}
//			},
//		},
//	})
//}

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
		ticker := time.NewTicker(time.Second)
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

func updateEndpointsWithLeaderIP(podIP string) {
	// 创建 Kubernetes 客户端
	config, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(config)
	endpointsClient := clientset.CoreV1().Endpoints("supernova")
	currentEndpoints, err := endpointsClient.Get(context.TODO(), "manager", metav1.GetOptions{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "updateEndpointsWithLeaderIP",
			"error":    err,
		}).Errorf("failed to get endpoints")
		return
	}

	currentEndpoints.Subsets = []v1.EndpointSubset{
		{
			Addresses: []v1.EndpointAddress{
				{
					IP: podIP,
				},
			},
			Ports: []v1.EndpointPort{
				{
					Name: "9090",
					Port: 9090,
				},
				{
					Name: "grpc",
					Port: 50051,
				},
			},
		},
	}

	_, err = endpointsClient.Update(context.TODO(), currentEndpoints, metav1.UpdateOptions{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "updateEndpointsWithLeaderIP",
			"error":    err,
		}).Errorf("failed to update endpoints")
		return
	}
}

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

	// Using channels to communicate errors or results from goroutines
	dbErrCh := make(chan error)
	redisErrCh := make(chan error)

	// Goroutine for inserting into database
	go func() {
		err := databaseClient.InsertTask(context.Background(), constants.TASKS_FULL_RECORD, taskDB)
		postgresqlOpsTotal.With(prometheus.Labels{"operation": "INSERT", "table": constants.TASKS_FULL_RECORD}).Inc()
		dbErrCh <- err
	}()

	// Goroutine for inserting into redis
	go func() {
		if data_structure_redis.CheckTasksInDuration(taskDB.ExecutionTime, DURATION) {
			// insert update to priority queue
			addNewJobToRedis(taskDB)
			redisThroughput.Inc()
			redisErrCh <- nil
		} else {
			redisErrCh <- nil
		}
	}()

	// Waiting for results from both goroutines
	dbErr := <-dbErrCh
	redisErr := <-redisErrCh

	// Handle potential errors
	if dbErr != nil {
		return "nil", dbErr
	}
	if redisErr != nil {
		return "nil", redisErr
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
	}).Info("Error counting running tasks")
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()
	if record == 0 {
		logger.WithFields(logrus.Fields{
			"function": "HandleDeleteTasks",
			"taskID":   taskID,
		}).Info("task is not running, delete it from database")

		// Initialize a wait group
		var wg sync.WaitGroup

		// Delete from Redis
		wg.Add(1)
		go func() {
			defer wg.Done()
			data_structure_redis.RemoveJobByID(taskID)
			redisThroughput.Inc()
		}()

		// Delete from database
		wg.Add(2)
		go func() {
			defer wg.Done()
			databaseClient.DeleteByID(context.Background(), constants.TASKS_FULL_RECORD, taskID)
			postgresqlOpsTotal.With(prometheus.Labels{"operation": "DELETE", "table": constants.TASKS_FULL_RECORD}).Inc()
		}()
		go func() {
			defer wg.Done()
			databaseClient.DeleteByID(context.Background(), constants.RUNNING_JOBS_RECORD, taskID)
			postgresqlOpsTotal.With(prometheus.Labels{"operation": "DELETE", "table": constants.RUNNING_JOBS_RECORD}).Inc()
		}()

		// Wait for all delete operations to finish
		wg.Wait()

		return nil
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
		"function":       "AddTasksFromDBWithTickers",
		"Now: ":          time.Now().UTC(),
		"End: ":          time.Now().UTC().Add(DURATION),
		"Time Tracker: ": timeTracker,
	}).Info("add tasks from database with tickers")
	tasks, _ := databaseClient.GetTasksInInterval(time.Now().UTC(), time.Now().UTC().Add(DURATION), timeTracker)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.TASKS_FULL_RECORD}).Inc()
	timeTracker = time.Now().UTC()
	for _, task := range tasks {
		logger.WithFields(logrus.Fields{
			"function": "AddTasksFromDBWithTickers",
		}).Infof("add task %v", task)
		go addNewJobToRedis(task)
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
		if msg.Payload == data_structure_redis.RETRY_AVAILABLE {
			for {
				retryTask, err := data_structure_redis.PopRetry()
				redisThroughput.Inc()
				if err != nil {
					break
				}
				go func(task *constants.TaskCache) {
					logger.WithFields(logrus.Fields{
						"function": "SubscribeToRedisChannel",
					}).Infof("retrying task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])
					_, status, err := HandoutTasksForExecuting(task)
					if err != nil {
						logger.WithFields(logrus.Fields{
							"function": "SubscribeToRedisChannel",
							"error":    err,
						}).Errorf("Failed to handout task for execution: %s", task.ID)
					}
					if status == constants.JOBFAILED {
						handleJobFailed(context.Background(), task, status)
					}
				}(retryTask)
			}
		}
	}
}

// handleTask 处理任务的主要逻辑，比较复杂
// 包含三个功能，redis的插入和数据库的插入是并发的，
// HandoutTasksForExecuting调用了grpc，也并发，
// handlejobfailed收到HandoutTasksForExecuting返回的结果相应的处理fail的状态
// 不管结果如何, 需要安排下一次任务
func handleTask(ctx context.Context, task *constants.TaskCache) {
	logger.WithFields(logrus.Fields{
		"function": "handleTask",
	}).Infof("dispatching task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])
	runningTaskInfo := generator.GenerateRunTimeTaskThroughTaskCache(task,
		constants.JOBDISPATCHED, "nil")

	dbInsertDone := make(chan error)
	redisSetDone := make(chan error)
	handoutTaskDone := make(chan handoutResultStruct)

	// Concurrently insert into the database
	go func() {
		err := databaseClient.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD, runningTaskInfo)
		dbInsertDone <- err
	}()

	// Concurrently set lease in Redis
	go func() {
		err := data_structure_redis.SetLeaseWithID(task.ID, task.ExecutionID, 7*time.Second)
		redisSetDone <- err
	}()

	// Concurrently execute the task using gRPC
	go func() {
		workerID, status, err := HandoutTasksForExecuting(task)
		handoutTaskDone <- handoutResultStruct{workerID: workerID, status: status, err: err}
	}()

	// Handle DB insert result
	go func() {
		err := <-dbInsertDone
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "handleTask",
				"error":    err,
			}).Error("Failed to insert task into database")
		}
	}()

	// Handle Redis set lease result
	go func() {
		err := <-redisSetDone
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "handleTask",
				"error":    err,
			}).Error("Failed to set lease with ID in Redis")
		}
	}()

	// No need to wait for the result of the recurring job dispatch.
	go handleRecurringJobAfterDispatch(task)

	// Handle the handout task result asynchronously
	go func() {
		handoutResult := <-handoutTaskDone
		logger.WithFields(logrus.Fields{
			"function": "executeMatureTasks",
		}).Infof("dispatch task: %s and executing by %s, found error: %v with status %s",
			task.ID, handoutResult.workerID, handoutResult.err, constants.StatusMap[handoutResult.status])

		if handoutResult.status == constants.JOBFAILED {
			handleJobFailed(ctx, task, handoutResult.status)
		}
	}()
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
	dispatchTotal.Inc()
	logger.WithFields(logrus.Fields{
		"function": "HandoutTasksForExecuting",
	}).Infof("handout task: %s with job type %s", task.ID, constants.TypeMap[task.JobType])

	workerService := os.Getenv("WORKER_SERVICE")
	if workerService == "" {
		workerService = WORKER_SERVICE
	}
	wrapper, err := getGRPCConnection(workerService)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "HandoutTasksForExecuting",
			"error":    err,
		}).Error("Failed to get grpc connection")
		return "", constants.JOBFAILED, err
	}
	defer releaseGRPCConnection(wrapper)

	ctx, cancel := context.WithTimeout(context.Background(), constants.EXECUTE_TASK_GRPC_TIMEOUT)
	defer cancel()

	payload := &pb.Payload{
		Format: int32(task.Payload.Format),
		Script: task.Payload.Script,
	}

	grpcOpsTotal.With(prometheus.Labels{
		"operation": "Execute Task", "sender": managerID}).Inc()
	r, err := wrapper.Client.ExecuteTask(ctx, &pb.TaskRequest{
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

	for {
		headNode := data_structure_redis.GetNextJob()
		if headNode != nil && data_structure_redis.CheckWithinThreshold(headNode.ExecutionTime) {
			tasks := data_structure_redis.PopJobsForDispatchWithBuffer()
			redisThroughput.Add(float64(len(tasks)))
			logger.WithFields(logrus.Fields{
				"function": "MonitorHeadNode",
			}).Infof("Dispatching this number %d of tasks", len(tasks))
			for _, task := range tasks {
				go handleTask(context.Background(), task)
			}
		}
	}
}

func handleJobFailed(ctx context.Context, task *constants.TaskCache, jobStatusCode int) error {
	logger.WithFields(logrus.Fields{
		"function": "handleJobFailed",
	}).Infof("handle job failed: %s Exec ID: %s", task.ID, task.ExecutionID)
	// update task status to 3
	// task_db only update when the update is completely failed (retries == 0 or outdated)
	if checkJobCompletelyFailed(task) {
		failedTasksTotal.Inc()
		return handleCompletelyFailedJob(ctx, task, jobStatusCode)
	}
	return handleRescheduleFailedJob(ctx, task)
}

func handleRescheduleFailedJob(ctx context.Context, task *constants.TaskCache) error {
	// update not completely failed, update the retries left, time
	// and add to redis priority queue
	logger.WithFields(logrus.Fields{
		"function": "handleRescheduleFailedJob",
	}).Infof("Eeschedule failed job: %s Exec ID: %s", task.ID, task.ExecutionID)

	err := data_structure_redis.RemoveLeaseWithID(ctx, task.ID, task.ExecutionID)
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

func handleCompletelyFailedJob(ctx context.Context, task *constants.TaskCache, jobStatusCode int) error {
	logger.WithFields(logrus.Fields{
		"function": "handleCompletelyFailedJob",
	}).Infof("Completely failed job: %s Exec ID: %s", task.ID, task.ExecutionID)

	// update completely failed need to update all information for report user.
	err := data_structure_redis.RemoveLeaseWithID(ctx, task.ID, task.ExecutionID)
	redisThroughput.Inc()
	// update the task according to task execution id.
	err = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD,
		task.ID, map[string]interface{}{
			"job_status":              jobStatusCode,
			"update_time":             time.Now().UTC(),
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
	diff := time.Now().UTC().Sub(executionTime)
	return diff > constants.ONE_TIME_JOB_RETRY_TIME
}

func checkRecurringJobOutdated(executionTime time.Time) bool {
	diff := time.Now().UTC().Sub(executionTime)
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

	// Start goroutines to handle the database and Redis operations asynchronously
	go func() {
		// Create a new context for this goroutine
		localCtx, cancel := context.WithTimeout(context.Background(), constants.TIMEOUT)
		defer cancel()

		err := data_structure_redis.RemoveLeaseWithID(localCtx, in.TaskId, in.ExecId)
		redisThroughput.Inc()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "NotifyTaskStatus",
			}).Errorf("Remove lease failed: %v", err)
		}
	}()

	go func() {
		// Create a new context for this goroutine
		localCtx, cancel := context.WithTimeout(context.Background(), constants.TIMEOUT)
		defer cancel()

		record, err := databaseClient.GetTaskByID(localCtx, constants.RUNNING_JOBS_RECORD, in.ExecId)
		postgresqlOpsTotal.With(prometheus.Labels{"operation": "GET", "table": constants.RUNNING_JOBS_RECORD}).Inc()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"function": "NotifyTaskStatus",
			}).Errorf("Get task by id: %s failed: %v", in.ExecId, err)
			return
		}

		runningTask := record.(*constants.RunTimeTask)
		payload := generator.GeneratePayload(runningTask.Payload.Format, runningTask.Payload.Script)
		taskCache := generator.GenerateTaskCache(
			runningTask.ID, runningTask.JobType, runningTask.CronExpr, runningTask.ExecutionTime, runningTask.RetriesLeft, payload)
		taskCache.ExecutionID = in.ExecId

		// Process based on the job status
		if in.Status == int32(constants.JOBSUCCEED) {
			handleJobSucceed(localCtx, taskCache)
		} else {
			handleJobFailed(localCtx, taskCache, constants.JOBFAILED)
		}
	}()

	// Immediately acknowledge the reception of the notification to the Worker
	return &pb.NotifyMessageResponse{Success: true}, nil
}

func handleJobSucceed(ctx context.Context, task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleJobSucceed",
	}).Infof("handle job succeed for job: %s for job type: %s", task.ID, constants.TypeMap[task.JobType])
	if task.JobType == constants.Recurring {
		return handleRecurringJobSucceed(ctx, task)
	}
	return handleOneTimeJobSucceed(ctx, task)
}

func handleOneTimeJobSucceed(ctx context.Context, task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleOneTimeJobSucceed",
	}).Infof("task: %s is one time job, update task_db table and remove lease", task.ID)
	// if it is one time job, only update the task_db table
	updateVars := map[string]interface{}{
		"update_time":             time.Now().UTC(),
		"job_status":              constants.JOBSUCCEED,
		"previous_execution_time": task.ExecutionTime,
	}
	err := databaseClient.UpdateByID(context.Background(),
		constants.TASKS_FULL_RECORD, task.ID, updateVars)
	postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()

	if err != nil {
		return err
	}
	err = data_structure_redis.RemoveLeaseWithID(ctx, task.ID, task.ExecutionID)
	redisThroughput.Inc()
	if err != nil {
		return err
	}
	return nil
}

func handleRecurringJobSucceed(ctx context.Context, task *constants.TaskCache) error {
	logger.WithFields(logrus.Fields{
		"function": "handleRecurringJobSucceed",
	}).Infof("task: %s is recurring job, update task_db table and remove lease", task.ID)
	// it is recur job, update execution time, update time, status, previous execution time
	cronExpr := task.CronExpr
	newExecutionTime := generator.DecryptCronExpress(cronExpr)
	updateVars := map[string]interface{}{
		"execution_time":          newExecutionTime,
		"update_time":             time.Now().UTC(),
		"job_status":              constants.JOBREENTERED,
		"previous_execution_time": task.ExecutionTime,
	}

	var dbErr, redisErr error
	var wg sync.WaitGroup

	// Update the database asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbErr = databaseClient.UpdateByID(context.Background(), constants.TASKS_FULL_RECORD, task.ID, updateVars)
		postgresqlOpsTotal.With(prometheus.Labels{"operation": "UPDATE", "table": constants.TASKS_FULL_RECORD}).Inc()
	}()

	// Handle the Redis operation asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		redisErr = data_structure_redis.RemoveLeaseWithID(ctx, task.ID, task.ExecutionID)
	}()

	wg.Wait()

	// Check for errors and return
	if dbErr != nil {
		return dbErr
	}
	if redisErr != nil {
		return redisErr
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
