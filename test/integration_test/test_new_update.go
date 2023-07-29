package integration_test

import (
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"log"
)

func main() {
	name := "TestTask"
	taskType := 1                 // 假设这是一个合法的任务类型
	cronExpression := "* * * * *" // 每分钟执行一次
	format := 0                   // 假设这是一个合法的格式
	script := "echo 'Hello, World!'"
	callBackURL := "http://example.com/callback"
	retries := 3

	callback, err := task_manager.HandleNewTasks(name, taskType, cronExpression, format, script, callBackURL, retries)
	if err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}
}
