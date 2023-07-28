package main

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
		api.GenerateTaskHandler(c)
	})
	r.DELETE("/task.:id", func(c *gin.Context) {
		api.DeleteTaskHandler(c)
	})
	r.GET("/task.:id", func(c *gin.Context) {
		api.GetTaskHandler(c)
	})
	r.PUT("/task.:id", func(c *gin.Context) {
		api.UpdateTaskHandler(c)
	})

	r.POST("/register", func(c *gin.Context) {
		api.RegisterUserHandler(c)
	})
	r.POST("/login", func(c *gin.Context) {
		api.LoginUserHandler(c)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
