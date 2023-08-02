package main

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "manager",
	Short: "MorFun_SuperNova Manager",
	Long: "Welcome to MorFun_SuperNova Manager, " +
		"Manager is designed to manager the tasks and " +
		"dispatch tasks to workers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("MorFun_SuperNova Manager")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Run MorFun_SuperNova Manager",
	Long:  "Start Run MorFun_SuperNova Manager, make manager to listen to tasks and dispatch them when needed",
	Run: func(cmd *cobra.Command, args []string) {
		task_manager.Init()
		fmt.Println("MorFun_SuperNova Manager Init")
		task_manager.Start()
		fmt.Println("MorFun_SuperNova Manager Start")
		go dispatch.InitManagerGRPC()
		time.Sleep(2 * time.Second)
		fmt.Println("***********************************************************")
		fmt.Println("***MorFun_SuperNova Manager All set, you are good to go***")
		fmt.Println("***********************************************************")
		r := gin.Default()

		r.LoadHTMLGlob("./app/frontend/*/*.html")

		r.POST("/api/generate", func(c *gin.Context) {
			api.GenerateTaskHandler(c)
		})
		r.DELETE("/api/task/:id/delete", func(c *gin.Context) {
			api.DeleteTaskHandler(c)
		})
		r.GET("/api/task/:id", func(c *gin.Context) {
			api.GetTaskHandler(c)
		})
		r.PUT("/api/task/:id/update", func(c *gin.Context) {
			api.UpdateTaskHandler(c)
		})
		r.GET("/api/dashboard", func(c *gin.Context) {
			api.DashboardHandler(c)
		})
		r.GET("/api/running_tasks", func(c *gin.Context) {
			api.RunningTasksHandler(c)
		})
		r.Run(":9090")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
		os.Exit(-1)
	}
}
