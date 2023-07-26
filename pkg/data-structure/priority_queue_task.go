package data_structure_redis

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
)

type PriorityQueue []*constants.TaskCache

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].ExecutionTime.Before(pq[j].ExecutionTime) // 越早的排在前面
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	task := x.(*constants.TaskCache)
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	*pq = old[0 : n-1]
	return task
}
