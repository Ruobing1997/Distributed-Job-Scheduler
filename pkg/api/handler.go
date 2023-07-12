package api

import (
	"fmt"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskRequest struct {
	Name     string `json:"name"`
	JobType  string `json:"jobType"`
	Schedule string `json:"schedule"`
	Payload  string `json:"payload"`
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
	payload := taskRequest.Payload

	fmt.Printf("name: %s, taskType: %s, schedule: %s, payload: %s\n", name, taskType, schedule, payload)

	task := generator.GenerateTask(name, taskType, schedule, payload, "")

	c.JSON(200, gin.H{
		"task": task,
	})
}
