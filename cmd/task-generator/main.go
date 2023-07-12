package main

import (
	"fmt"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskRequest struct {
	Name     string `json:"name"`
	JobType  string `json:"jobType"`
	Schedule string `json:"schedule"`
	Payload  string `json:"payload"`
}

func main() {
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/generate", func(c *gin.Context) {
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
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
