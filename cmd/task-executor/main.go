package main

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_executor "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-executor"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "MorFun_SuperNova Worker",
	Long: "Welcome to MorFun_SuperNova Worker, " +
		"Worker is designed to execute tasks dispatched by the manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("MorFun_SuperNova Worker")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Run MorFun_SuperNova Worker",
	Long:  "Start Run MorFun_SuperNova Worker, make worker listen to the tasks dispatched by the manager and execute them",
	Run: func(cmd *cobra.Command, args []string) {
		task_executor.Init()
		fmt.Println("MorFun_SuperNova Worker Init")
		go dispatch.InitWorkerGRPC()
		time.Sleep(2 * time.Second)
		fmt.Println("All set, Worker is ready to execute tasks")
		select {} // run forever
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
