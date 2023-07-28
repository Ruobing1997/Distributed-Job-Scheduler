package middleware

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
)

var executeTaskFunc func(task *constants.TaskCache) error

func SetExecuteTaskFunc(f func(task *constants.TaskCache) error) {
	executeTaskFunc = f
}

func ExecuteTaskFuncThroughMediator(task *constants.TaskCache) error {
	if executeTaskFunc != nil {
		return executeTaskFunc(task)
	}
	return fmt.Errorf("ExecuteTask function not set")
}
