package main

import (
	"context"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"time"
)

type ServerControl struct{}

var server *http.Server

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
		fmt.Println("***********************************************************")
		fmt.Println("***MorFun_SuperNova Manager All set, you are good to go***")
		fmt.Println("***********************************************************")
		managerControl := &ServerControl{}
		task_manager.InitLeaderElection(managerControl)
		fmt.Println("MorFun_SuperNova Manager Start Leader Election")
		time.Sleep(2 * time.Second)
	},
}

func (*ServerControl) StartAPIServer() {
	r := gin.Default()
	r.Use(cors.Default())
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

	server = &http.Server{
		Addr:    ":9090",
		Handler: r,
	}

	server.ListenAndServe()
}

func (*ServerControl) StopAPIServer() error {
	if server != nil {
		err := server.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
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
