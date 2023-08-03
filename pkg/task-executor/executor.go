package task_executor

import (
	"context"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/postgreSQL"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var (
	workerID       = os.Getenv("HOSTNAME")
	databaseClient *postgreSQL.Client
)

type Result struct {
	ID    string
	Error error
}

func Init() {
	logFile, err := os.OpenFile("./logs/workers/worker-"+workerID+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	multiWrite := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWrite)

	log.Printf("Task Executor and its grpc server are initialized")

	databaseClient = postgreSQL.NewpostgreSQLClient()
	log.Printf("database set up done")
}

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

func InitWorkerGRPC() {
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

func ExecuteTask(task *constants.TaskCache) (string, error) {
	log.Printf("-------------------------------------------------------")
	log.Printf("Received task with payload: format: %d, content: %s",
		task.Payload.Format, task.Payload.Script)
	var err error
	err = databaseClient.UpdateByExecutionID(context.Background(),
		constants.RUNNING_JOBS_RECORD, task.ExecutionID,
		map[string]interface{}{
			"job_status": constants.JOBRUNNING,
			"worker_id":  workerID,
		})
	if err != nil {
		log.Fatalf("insert task: %s with Execution ID: %s "+
			"into database failed: %v", task.ID, task.ExecutionID, err)
		return workerID, err
	}
	go func() {
		log.Printf("-------------------------------------------------------")
		log.Printf("Start Executing Task: %s", task.ID)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go MonitorLease(ctx, task.ID, task.ExecutionID)

		switch task.Payload.Format {
		case constants.SHELL:
			err = executeScript(task.Payload.Script, constants.SHELL)
		case constants.PYTHON:
			err = executeScript(task.Payload.Script, constants.PYTHON)
		case constants.EMAIL:
			err = executeEmail(task.Payload.Script)
		}
		if err != nil {
			log.Printf("Task: %s failed with error: %v", task.ID, err)
			err = databaseClient.UpdateByExecutionID(context.Background(),
				constants.RUNNING_JOBS_RECORD, task.ExecutionID,
				map[string]interface{}{"job_status": constants.JOBFAILED})
			if err != nil {
				notifyManagerTaskResult(task.ID, task.ExecutionID, constants.JOBFAILED)
				return
			}
			notifyManagerTaskResult(task.ID, task.ExecutionID, constants.JOBFAILED)
		} else {
			err = databaseClient.UpdateByExecutionID(context.Background(),
				constants.RUNNING_JOBS_RECORD, task.ExecutionID,
				map[string]interface{}{"job_status": constants.JOBSUCCEED})
			if err != nil {
				notifyManagerTaskResult(task.ID, task.ExecutionID, constants.JOBFAILED)
				log.Fatalf("update task execid: %s failed: %v",
					task.ExecutionID, err)
			} else {
				notifyManagerTaskResult(task.ID, task.ExecutionID, constants.JOBSUCCEED)
			}
		}
	}()
	return workerID, nil
}

func notifyManagerTaskResult(taskID string, execID string, jobStatus int) {
	managerService := os.Getenv("MANAGER_SERVICE")
	if managerService == "" {
		managerService = MANAGER_SERVICE
	}
	conn, err := grpc.Dial(managerService+":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTaskServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.EXECUTE_TASK_GRPC_TIMEOUT)
	defer cancel()

	success, err := client.NotifyTaskStatus(ctx, &pb.NotifyMessageRequest{
		TaskId:   taskID,
		WorkerId: workerID,
		ExecId:   execID,
		Status:   int32(jobStatus),
	})

	if err != nil {
		log.Printf("could not notify manager task result: %v", err)
		return
	}

	log.Printf("Task: %s notify manager task result: %v", taskID, success.Success)
}

func MonitorLease(ctx context.Context, taskId string, execID string) {
	log.Printf("-------------------------------------------------------")
	log.Printf("execID: %s is monitoring lease for update: %s", execID, taskId)
	ticker := time.NewTicker(constants.LEASE_RENEW_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("execID: %s is renewing lease for update: %s", execID, taskId)
			err := WorkerRenewLease(taskId, execID, constants.LEASE_DURATION)
			if err != nil {
				log.Printf("failed to renew lease for update execID %s: %v", execID, err)
			}
		}
	}

}

func RenewLease(taskID string, execID string, duration time.Duration) (bool, error) {
	managerService := os.Getenv("MANAGER_SERVICE")
	if managerService == "" {
		managerService = MANAGER_SERVICE
	}
	conn, err := grpc.Dial(managerService+":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewLeaseServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.RENEW_LEASE_GRPC_TIMEOUT)
	defer cancel()

	success, err := client.RenewLease(ctx, &pb.RenewLeaseRequest{
		Id:            taskID,
		ExecId:        execID,
		LeaseDuration: durationpb.New(duration),
	})

	if err != nil {
		return false, fmt.Errorf("could not renew lease: %v", err)
	}

	return success.Success, nil
}

func WorkerRenewLease(taskID string, execID string, duration time.Duration) error {
	success, err := RenewLease(taskID, execID, duration)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to renew lease for update %s", taskID)
	}

	return nil
}

func executeScript(scriptContent string, scriptType int) error {
	var cmd *exec.Cmd

	switch scriptType {
	case constants.SHELL:
		shellCommand := "/bin/sh"
		if os.Getenv("OS") == "Windows_NT" {
			shellCommand = "cmd.exe"
		}
		cmd = exec.Command(shellCommand, "-c", scriptContent)
	case constants.PYTHON:
		cmd = exec.Command("python3", "-c", scriptContent)
	default:
		return fmt.Errorf("unsupported script type: %d", scriptType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s script execution failed: %v, output: %s",
			constants.PayloadTypeMap[scriptType], err, output)
	}
	log.Printf("%s script executed successfully: %s", constants.PayloadTypeMap[scriptType], output)
	return nil
}

func executeEmail(emailInfo string) error {
	// Here you can implement the logic to send an email based on the emailInfo
	// For example, you can use a package like "net/smtp" to send emails in Go
	// For simplicity, I'm just printing the emailInfo,
	// and this can be used to test bad cases
	log.Printf("Sending email with info: %s", emailInfo)

	return fmt.Errorf("email not implemented yet")
}

func (s *ServerImpl) ExecuteTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	payload := generator.GeneratePayload(int(in.Payload.Format), in.Payload.Script)
	task := &constants.TaskCache{
		ID:            in.Id,
		Payload:       payload,
		ExecutionTime: in.ExecutionTime.AsTime(),
		RetriesLeft:   int(in.MaxRetryCount),
		ExecutionID:   in.ExecId,
	}

	workerID, err := ExecuteTask(task)
	if err != nil {
		return &pb.TaskResponse{Id: workerID, Status: constants.JOBFAILED}, err
	}
	return &pb.TaskResponse{Id: workerID, Status: constants.JOBDISPATCHED}, nil
}
