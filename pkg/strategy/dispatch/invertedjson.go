// Package dispatch
// /* Use the invertedjson method as the strategy for dispatching the request.
package dispatch

import (
	"bytes"
	"encoding/json"
	task_executor "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-executor"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"net/http"
)

func SendTasksToWorker(task *constants.TaskCache) (int, error) {
	taskData, err := json.Marshal(task)
	if err != nil {
		return JobFailed, err
	}
	req, err := http.NewRequest(POST, INVERTED_JSON_K8S_SERVICE_URL+TASK_CHANNEL,
		bytes.NewBuffer(taskData))
	if err != nil {
		return JobFailed, err
	}

	req.Header.Set("Content-Type", APPLICATION_JSON)
	req.Header.Set("type", "async")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return JobFailed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var workerStatus int
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&workerStatus)
		if err != nil {
			return JobFailed, err
		}
		return workerStatus, nil

	}
	return JobFailed, nil
}

func WorkerProcessTasks(w http.ResponseWriter, r *http.Request) (int, error) {

	// get task from redis and return received response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JobDispatched)

	var task *constants.TaskCache
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		return JobFailed, err
	}

	resultChannel := make(chan error)

	go func() {
		resultChannel <- task_executor.ExecuteTask(task)
	}()

	taskErr := <-resultChannel

	if taskErr != nil {
		return JobFailed, taskErr
	}

	return JobSucceed, nil
}
