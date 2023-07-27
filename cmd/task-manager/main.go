package main

import (
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	serverIP   string
	serverPort string
)

var rootCmd = &cobra.Command{
	Use:     "MorFun_SuperNova Manager",
	Short:   "Start Run MorFun_SuperNova Manager",
	Long:    "Welcome to MorFun_SuperNova Manager",
	Example: "./update-manager -i 0.0.0.0 -p 5000",
	Run: func(cmd *cobra.Command, args []string) {
		task_manager.Init()
		fmt.Println("MorFun_SuperNova Manager Init")
		task_manager.Start()
		fmt.Println("MorFun_SuperNova Manager Start")
		go dispatch.Init()
		fmt.Println("Dispatch Strategy Start")
	},
}

func init() {
	rootCmd.Flags().StringVarP(&serverIP, "ip", "i",
		"0.0.0.0", "Server IP")
	rootCmd.Flags().StringVarP(&serverPort, "port", "p",
		"5000", "Server port")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
		os.Exit(-1)
	}
}
