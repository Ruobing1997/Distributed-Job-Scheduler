package api

import (
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskRequest struct {
	Name       string  `json:"name"`
	JobType    int     `json:"jobType"`
	Schedule   int     `json:"schedule"`
	Importance float64 `json:"importance"`
	Payload    string  `json:"payload"`
}

func GenerateTaskHandler(c *gin.Context) {
	var taskRequest TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := taskRequest.Name
	taskType := taskRequest.JobType
	schedule := taskRequest.Schedule
	importance := taskRequest.Importance
	payload := taskRequest.Payload

	task := generator.GenerateTask(
		name,
		taskType,
		schedule,
		importance,
		payload,
		"")

	c.JSON(200, gin.H{
		"task": task,
	})
}
