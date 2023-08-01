package middleware

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
)

var executeTaskFunc func(task *constants.TaskCache) (string, error)

func SetExecuteTaskFunc(f func(task *constants.TaskCache) (string, error)) {
	executeTaskFunc = f
}

func ExecuteTaskFuncThroughMediator(task *constants.TaskCache) (string, error) {
	if executeTaskFunc != nil {
		return executeTaskFunc(task)
	}
	return "nil", fmt.Errorf("ExecuteTask function not set")
}
