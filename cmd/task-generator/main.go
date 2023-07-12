package main

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	r.POST("/generate", api.GenerateTaskHandler)

	r.Run() // listen and serve on 0.0.0.0:8080
}
