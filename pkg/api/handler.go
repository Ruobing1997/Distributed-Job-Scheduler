package api

import (
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskRequest struct {
	JobName  string `json:"jobName"`
	JobType  int    `json:"jobType"`
	CronExpr string `json:"cronExpr"`
	Format   int    `json:"format"`
	Script   string `json:"script"`
	Retries  int    `json:"retries"`
}

func GenerateTaskHandler(c *gin.Context) {
	var taskRequest TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobName := taskRequest.JobName
	jobType := taskRequest.JobType
	cronExpr := taskRequest.CronExpr
	format := taskRequest.Format
	script := taskRequest.Script
	retries := taskRequest.Retries

	task := generator.GenerateTask(
		jobName,
		jobType,
		cronExpr,
		format,
		script,
		retries)

	c.JSON(200, gin.H{
		"task": task,
	})
}
