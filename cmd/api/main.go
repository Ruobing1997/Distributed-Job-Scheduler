package main

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	r := gin.Default()
	task_manager.Init()
	fmt.Println("MorFun_SuperNova Manager Init")
	task_manager.Start()
	fmt.Println("MorFun_SuperNova Manager Start")
	go dispatch.Init()
	fmt.Println("Dispatch Strategy Start")
	time.Sleep(2 * time.Second)
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

	r.POST("/register", func(c *gin.Context) {
		api.RegisterUserHandler(c)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
