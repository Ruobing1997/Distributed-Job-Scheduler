package main

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	configFile string
	logDir     string
	serverIP   string
	serverPort string
)

var rootCmd = &cobra.Command{
	Use:     "MorFun_SuperNova Manager",
	Short:   "Start Run MorFun_SuperNova Manager",
	Long:    "Welcome to MorFun_SuperNova Manager",
	Example: "./task-manager -c ./config.toml -l ./logs -i 0.0.0.0 -p 5000",
	Run: func(cmd *cobra.Command, args []string) {
		task_manager.Init()
		task_manager.Start()
		go dispatch.Init()
	},
}

func init() {
	// 设置默认值
	rootCmd.Flags().StringVarP(&configFile, "config", "c",
		"./config.toml", "Config file, example: ./config.toml")
	rootCmd.Flags().StringVarP(&logDir, "log", "l",
		"./logs", "Log directory, example: ./logs")
	rootCmd.Flags().StringVarP(&serverIP, "ip", "i",
		"0.0.0.0", "Server IP")
	rootCmd.Flags().StringVarP(&serverPort, "port", "p",
		"5000", "Server port")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing root command: %v", err)
		os.Exit(-1)
	}
}
