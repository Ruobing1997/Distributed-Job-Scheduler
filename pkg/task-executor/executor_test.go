package task_executor

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"testing"
)

func TestExecuteScript(t *testing.T) {
	// 测试 shell 脚本
	shellScript := "echo 'Hello, World!'"
	err := executeScript(shellScript, constants.SHELL)
	if err != nil {
		t.Errorf("Shell script execution failed: %v", err)
	}

	// 测试 Python 脚本
	pythonScript := "print('Hello, World!')"
	err = executeScript(pythonScript, constants.PYTHON)
	if err != nil {
		t.Errorf("Python script execution failed: %v", err)
	}

	// 测试不支持的脚本类型
	unsupportedScriptType := 999
	err = executeScript(shellScript, unsupportedScriptType)
	if err == nil {
		t.Errorf("Expected error for unsupported script type, but got no error")
	}
}
