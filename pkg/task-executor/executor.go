package task_executor

import (
	"context"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/postgreSQL"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/middleware"
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
	middleware.SetExecuteTaskFunc(ExecuteTask)
	log.Printf("ExecuteTask function set up done")
	databaseClient = postgreSQL.NewpostgreSQLClient()
	log.Printf("database set up done")
}

func ExecuteTask(task *constants.TaskCache) (string, error) {
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

	go MonitorLease(ctx, task.ID)

	switch task.Payload.Format {
	case constants.SHELL:
		err = executeScript(task.Payload.Script, constants.SHELL)
	case constants.PYTHON:
		err = executeScript(task.Payload.Script, constants.PYTHON)
	case constants.EMAIL:
		err = executeEmail(task.Payload.Script)
	}
	if err != nil {
		return "nil", fmt.Errorf("error executing task: %s with %v", task.ID, err)
	} else {
		return workerID, nil
	}
}

func MonitorLease(ctx context.Context, taskId string) {
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
			err := WorkerRenewLease(taskId, constants.LEASE_DURATION)
			if err != nil {
				log.Printf("failed to renew lease for update %s: %v", taskId, err)
			}
		}
	}

}

func RenewLease(taskID string, duration time.Duration) (bool, error) {
	managerService := os.Getenv("MANAGER_SERVICE")
	if managerService == "" {
		managerService = MANAGER_SERVICE
	}
	conn, err := grpc.Dial(managerService+":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewLeaseServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.RENEW_LEASE_GRPC_TIMEOUT)
	defer cancel()

	success, err := client.RenewLease(ctx, &pb.RenewLeaseRequest{
		Id:            taskID,
		LeaseDuration: durationpb.New(duration),
	})

	if err != nil {
		return false, fmt.Errorf("could not renew lease: %v", err)
	}

	return success.Success, nil
}

func WorkerRenewLease(taskID string, duration time.Duration) error {
	success, err := RenewLease(taskID, duration)
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
	// For simplicity, I'm just printing the emailInfo
	log.Printf("Sending email with info: %s", emailInfo)

	return fmt.Errorf("email not implemented yet")
}
