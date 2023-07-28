package task_executor

import (
	"context"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/middleware"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"os"
	"os/exec"
	"time"
)

func Init() {
	go dispatch.InitWorkerGRPC()
	middleware.SetRenewLeaseFunction(dispatch.RenewLease)
	log.Printf("Task Executor and its grpc server are initialized")
}

func ExecuteTask(task *constants.TaskCache) error {
	log.Printf("Received update with payload: format: %d, content: %s",
		task.Payload.Format, task.Payload.Script)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go MonitorLease(ctx, task.ID)

	var err error
	switch task.Payload.Format {
	case constants.SHELL:
		err = executeScript(task.Payload.Script, constants.SHELL)
	case constants.PYTHON:
		err = executeScript(task.Payload.Script, constants.PYTHON)
	case constants.EMAIL:
		err = executeEmail(task.Payload.Script)
	}
	if err != nil {
		return fmt.Errorf("error executing update: %v", err)
	} else {
		return nil
	}
}

func MonitorLease(ctx context.Context, taskId string) {
	ticker := time.NewTicker(constants.LEASE_RENEW_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := WorkerRenewLease(taskId, time.Now().Add(constants.LEASE_DURATION))
			if err != nil {
				log.Printf("failed to renew lease for update %s: %v", taskId, err)
			}
		}
	}

}

func WorkerRenewLease(taskID string, newLeaseTime time.Time) error {
	success, err := middleware.RenewLeaseThroughMediator(taskID, newLeaseTime)
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
		cmd = exec.Command("python", "-c", scriptContent)
	default:
		return fmt.Errorf("unsupported script type: %d", scriptType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%d script execution failed: %v, output: %s", scriptType, err, output)
	}
	log.Printf("%d script executed successfully: %s", scriptType, output)
	return nil
}

func executeEmail(emailInfo string) error {
	// Here you can implement the logic to send an email based on the emailInfo
	// For example, you can use a package like "net/smtp" to send emails in Go
	// For simplicity, I'm just printing the emailInfo
	log.Printf("Sending email with info: %s", emailInfo)

	return fmt.Errorf("email not implemented yet")
}
