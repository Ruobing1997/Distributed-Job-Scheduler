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
	"os"
	"os/exec"
	"time"
)

var (
	workerID       = os.Getenv("HOSTNAME")
	databaseClient *postgreSQL.Client
)

type ExecuteResult struct {
	WorkerID string
	TaskID   string
	Status   int32
}

func Init() {
	logFile, err := os.OpenFile("./logs/workers/worker-"+workerID+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	multiWrite := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWrite)

	if err != nil {
		log.Fatalf("error initializing worker bidirect grpc: %v", err)
	}

	log.Printf("Task Executor and its grpc server are initialized")

	databaseClient = postgreSQL.NewpostgreSQLClient()
	log.Printf("database set up done")
}

func InitWorkerBidirectGRPC() (pb.TaskManagerService_TaskStreamClient, error) {
	// MANAGER_SERVICE_IP in format of "manager-service.default.svc.cluster.local:50051"
	managerService := os.Getenv("MANAGER_SERVICE_IP")
	if managerService == "" {
		return nil, fmt.Errorf("MANAGER_SERVICE_IP is not set")
	}
	conn, err := grpc.Dial(managerService, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := pb.NewTaskManagerServiceClient(conn)

	stream, err := client.TaskStream(context.Background())

	if err != nil {
		return nil, fmt.Errorf("could not get stream: %v", err)
	}

	initialMsg := &pb.TaskStreamRequest{
		Request: &pb.TaskStreamRequest_WorkerIdentity{
			WorkerIdentity: &pb.WorkerIdentity{
				WorkerId: workerID,
			},
		},
	}

	if err := stream.Send(initialMsg); err != nil {
		return nil, fmt.Errorf("failed to send a note: %v", err)
	}

	return stream, nil
}

func HandleManagerMessages(stream pb.TaskManagerService_TaskStreamClient) {
	log.Printf("-------------------------------------------------------")
	log.Printf("Start handling messages from manager")

	resultChan := make(chan ExecuteResult)
	go func() {
		for result := range resultChan {
			response := &pb.TaskStreamRequest{
				Request: &pb.TaskStreamRequest_TaskResponse{
					TaskResponse: &pb.TaskResponse{
						WorkerID: result.WorkerID,
						TaskID:   result.TaskID,
						Status:   result.Status,
					},
				},
			}

			if err := stream.Send(response); err != nil {
				log.Printf("Error sending response to manager: %v", err)
			}
		}

	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Received EOF from manager")
			return
		}
		if err != nil {
			log.Printf("Error receiving task from manager: %v", err)
			continue
		}

		switch x := in.Response.(type) {
		case *pb.TaskStreamResponse_TaskRequest:
			task := x.TaskRequest
			payload := generator.GeneratePayload(int(task.Payload.Format), task.Payload.Script)
			taskCache := &constants.TaskCache{
				ID:            task.Id,
				Payload:       payload,
				ExecutionTime: task.ExecutionTime.AsTime(),
				RetriesLeft:   int(task.MaxRetryCount),
			}
			go ExecuteTask(taskCache, resultChan, stream)
		}
	}
}

func ExecuteTask(task *constants.TaskCache,
	resultChan chan<- ExecuteResult, stream pb.TaskManagerService_TaskStreamClient) {
	log.Printf("-------------------------------------------------------")
	log.Printf("Received update with payload: format: %d, content: %s",
		task.Payload.Format, task.Payload.Script)
	runningTaskInfo := generator.GenerateRunTimeTaskThroughTaskCache(task,
		constants.JOBRUNNING, workerID)
	var err error
	err = databaseClient.InsertTask(context.Background(), constants.RUNNING_JOBS_RECORD,
		runningTaskInfo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go MonitorLease(ctx, task.ID, stream)

	switch task.Payload.Format {
	case constants.SHELL:
		err = executeScript(task.Payload.Script, constants.SHELL)
	case constants.PYTHON:
		err = executeScript(task.Payload.Script, constants.PYTHON)
	case constants.EMAIL:
		err = executeEmail(task.Payload.Script)
	}

	var status int32
	if err != nil {
		status = constants.JOBFAILED
	} else {
		status = constants.JOBSUCCEED
	}

	resultChan <- ExecuteResult{
		WorkerID: workerID,
		TaskID:   task.ID,
		Status:   status,
	}
}

func MonitorLease(ctx context.Context, taskId string, stream pb.TaskManagerService_TaskStreamClient) {
	log.Printf("-------------------------------------------------------")
	log.Printf("Worker: %s is monitoring lease for update: %s", workerID, taskId)
	ticker := time.NewTicker(constants.LEASE_RENEW_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("Worker: %s is renewing lease for update: %s", workerID, taskId)
			renewRequest := &pb.TaskStreamRequest{
				Request: &pb.TaskStreamRequest_RenewLease{
					RenewLease: &pb.RenewLeaseRequest{
						Id:            taskId,
						LeaseDuration: durationpb.New(constants.LEASE_DURATION),
					},
				},
			}
			if err := stream.Send(renewRequest); err != nil {
				log.Printf("Error sending renew lease request to manager: %v", err)
			}
		}
	}

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
	// For simplicity, I'm just printing the emailInfo
	log.Printf("Sending email with info: %s", emailInfo)

	return fmt.Errorf("email not implemented yet")
}
