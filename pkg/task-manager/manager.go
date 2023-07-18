package task_manager

import (
	"container/heap"
	"context"
	databasehandler "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"time"
)

type PriorityQueue []*constants.TaskCache

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].ExecutionTime.Before(pq[j].ExecutionTime) // 越早的排在前面
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	task := x.(*constants.TaskCache)
	task.Index = n
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	task.Index = -1 // 从堆中移除
	*pq = old[0 : n-1]
	return task
}

var pq = &PriorityQueue{}

func GetPriorityQueue() *PriorityQueue {
	return pq
}

func storeTasksToDB(client databasehandler.DatabaseClient, taskDB *constants.TaskDB) error {
	err := client.InsertTask(context.Background(), taskDB)
	if err != nil {
		return err
	}
	return nil
}

func HandleTasks(client databasehandler.DatabaseClient,
	name string, taskType int, cronExpression string,
	payload string, callBackURL string, retries int) error {
	// get task generated form generator
	taskDB := generator.GenerateTask(name, taskType,
		cronExpression, payload, callBackURL, retries)
	// insert task to database
	err := storeTasksToDB(client, taskDB)
	if err != nil {
		return err
	}
	// generate task for cache:
	taskCache := generateTaskCache(taskDB.ID, taskDB.ExecutionTime, taskDB.NextExecutionTime)
	// insert task to priority queue
	heap.Push(pq, taskCache)
	return nil
}

func DispatchTasks() {
	
}

func generateTaskCache(id string, executionTime time.Time,
	nextExecutionTime time.Time) *constants.TaskCache {
	return &constants.TaskCache{
		ID:                id,
		ExecutionTime:     executionTime,
		NextExecutionTime: nextExecutionTime,
	}
}
