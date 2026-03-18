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
	Short: "Run a command in a container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var childCmd = &cobra.Command{
	Use:    "child [command]",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		child()
	},
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running containers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ID\t\tSTATUS\t\tCOMMAND")
		// treba implementirati kada dodam vise od jendogh kontejnera
	},
}

func init() {
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
