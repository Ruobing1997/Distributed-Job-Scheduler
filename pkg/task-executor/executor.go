package task_executor

import (
	"context"
	"encoding/json"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"os/exec"
	"time"
)

func ExecuteTask(task *constants.TaskCache) error {
	log.Printf("Received task with payload: %s", task.Payload)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go MonitorLease(ctx, task.ID)

	var payloadJson constants.PayloadJson
	err := json.Unmarshal([]byte(task.Payload), &payloadJson)
	if err != nil {
		return fmt.Errorf("unmarshal payload failed: %v", err)
	}
	switch payloadJson.Type {
	case constants.SHELL:
		err = executeScript(payloadJson.Content, constants.SHELL)
	case constants.PYTHON:
		err = executeScript(payloadJson.Content, constants.PYTHON)
	case constants.EMAIL:
		err = executeEmail(payloadJson.Content)
	}
	if err != nil {
		return fmt.Errorf("error executing task: %v", err)
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
				log.Printf("failed to renew lease for task %s: %v", taskId, err)
			}
		}
	}

}

func WorkerRenewLease(taskID string, newLeaseTime time.Time) error {
	success, err := dispatch.RenewLease(taskID, newLeaseTime)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to renew lease for task %s", taskID)
	}

	return nil
}

func executeScript(scriptContent string, scriptType int) error {
	var cmd *exec.Cmd

	switch scriptType {
	case constants.SHELL:
		cmd = exec.Command("/bin/sh", "-c", scriptContent)
	case constants.PYTHON:
		cmd = exec.Command("python3", "-c", scriptContent)
	default:
		return fmt.Errorf("unsupported script type: %s", scriptType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s script execution failed: %v, output: %s", scriptType, err, output)
	}
	log.Printf("%s script executed successfully: %s", scriptType, output)
	return nil
}

func executeEmail(emailInfo string) error {
	// Here you can implement the logic to send an email based on the emailInfo
	// For example, you can use a package like "net/smtp" to send emails in Go
	// For simplicity, I'm just printing the emailInfo
	log.Printf("Sending email with info: %s", emailInfo)

	return fmt.Errorf("email not implemented yet")
}
