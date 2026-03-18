package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gobox",
	Short: "A simple container runtime written in Go",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to gobox! Use 'gobox run <command>' to run a command in a container.")
	},
}

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a new container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memory, _ := cmd.Flags().GetString("memory")
		pids, _ := cmd.Flags().GetString("pids")
		run(args, memory, pids)
	},
}
var childCmd = &cobra.Command{
	Use:    "child [command]",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		child(args)
	},
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running containers",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := os.ReadDir("/var/lib/gobox/")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%-20s %-20s %-20s %-20s\n", "ID", "STATUS", "COMMAND", "CREATED")
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			state := getContainerById(file.Name()[:len(file.Name())-5])
			if state != nil {
				fmt.Printf("%-20s %-20s %-20s %-20s\n", state.Id, state.Status, state.Command, state.Created.Format("2006-01-02 15:04:05"))
			}
		}
	},
}

func init() {
	runCmd.Flags().StringP("memory", "m", "100m", "Memory limit")
	runCmd.Flags().StringP("pids", "p", "20", "Max number of processes")
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(childCmd)
	rootCmd.AddCommand(psCmd)
}

func executeCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
