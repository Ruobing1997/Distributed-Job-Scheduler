package Factory

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"log"
	"os"
	"os/exec"
)

type TaskExecutor interface {
	ExecuteTask(scriptContent string) error
}

type ShellExecutor struct{}

func (se *ShellExecutor) ExecuteTask(scriptContent string) error {
	shellCommand := "/bin/sh"
	if os.Getenv("OS") == "Windows_NT" {
		shellCommand = "cmd.exe"
	}
	cmd := exec.Command(shellCommand, "-c", scriptContent)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shell script execution failed: %v, output: %s", err, output)
	}
	log.Printf("Shell script executed successfully: %s", output)
	return nil
}

type PythonExecutor struct{}

func (pe *PythonExecutor) ExecuteTask(scriptContent string) error {
	cmd := exec.Command("python", "-c", scriptContent)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python script execution failed: %v, output: %s", err, output)
	}
	log.Printf("python script executed successfully: %s", output)
	return nil
}

func getExecutor(taskType int) TaskExecutor {
	switch taskType {
	case constants.SHELL:
		return &ShellExecutor{}
	case constants.PYTHON:
		return &PythonExecutor{}
	default:
		return nil
	}
}
