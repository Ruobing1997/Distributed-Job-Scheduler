package main

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
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
		fmt.Println("All set, you are good to go")
		select {} // start forever
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
