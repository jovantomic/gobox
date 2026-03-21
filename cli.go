package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	Use:                "run [command]",
	Short:              "Run a new container",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	TraverseChildren:   true,
	Run: func(cmd *cobra.Command, args []string) {
		memory, _ := cmd.Flags().GetString("memory")
		pids, _ := cmd.Flags().GetString("pids")
		imageName, _ := cmd.Flags().GetString("image")
		run(args, memory, pids, imageName)
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
		containers := getAllContainers()
		if len(containers) == 0 {
			fmt.Println("No containers found")
			return
		}
		fmt.Printf("%-12s %-10s %-20s %s\n", "ID", "STATUS", "COMMAND", "CREATED")
		for _, c := range containers {
			fmt.Printf("%-12s %-10s %-20s %s\n", c.Id, c.Status, c.Command, c.Created.Format("2006-01-02 15:04:05"))
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop a running container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stopContainer(args[0])
	},
}
var rmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a stopped container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeContainer(args[0])
	},
}

var logCmd = &cobra.Command{
	Use:   "logs [id]",
	Short: "Show logs of a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showLogs(args[0])
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull [image]",
	Short: "Pull an image from Docker Hub",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageName := args[0]
		fmt.Printf("Pulling image %s...\n", imageName)
		err := pullImage(imageName)
		if err != nil {
			fmt.Printf("Error pulling image: %v\n", err)
			return
		}
		fmt.Printf("Image %s pulled successfully!\n", imageName)

	},
}

var execCmd = &cobra.Command{
	Use:   "exec [id] [command]",
	Short: "Execute a command in a running container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		command := args[1:]
		execContainer(id, command)
	},
}

var portForwardCmd = &cobra.Command{
	Use:   "port [id] [host_port]:[container_port]",
	Short: "Forward a port from host to container",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		state := getContainerById(id)
		if state == nil {
			fmt.Printf("Container with ID %s not found\n", id)
			return
		}

		var hostPort, contPort int
		fmt.Sscanf(args[1], "%d:%d", &hostPort, &contPort)

		forwardPort(hostPort, contPort, state.IP)

		fmt.Printf("Forwarded %d -> %s:%d\n", hostPort, id, contPort)
	},
}

var cpCmd = &cobra.Command{
	Use:   "cp [id] [src] [dest]",
	Short: "Copy file from host to container",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		src := args[1]
		dest := args[2]

		state := getContainerById(id)
		if state == nil {
			fmt.Printf("Container %s not found\n", id)
			return
		}

		merged := filepath.Join("/var/lib/gobox/overlay", id, "merged")
		fullDest := filepath.Join(merged, dest)

		os.MkdirAll(filepath.Dir(fullDest), 0755)
		data, err := os.ReadFile(src)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", src, err)
			return
		}
		err = os.WriteFile(fullDest, data, 0644)
		if err != nil {
			fmt.Printf("Error writing: %v\n", err)
			return
		}
		fmt.Printf("Copied %s -> %s:%s\n", src, id, dest)
	},
}

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint [id]",
	Short: "Checkpoint a running container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		err := checkpointContainer(id)
		if err != nil {
			fmt.Printf("Error checkpointing container: %v\n", err)
			return
		}
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore [id]",
	Short: "Restore a checkpointed container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		err := restoreContainer(id)
		if err != nil {
			fmt.Printf("Error restoring container: %v\n", err)
			return
		}
	},
}

func init() {
	runCmd.Flags().StringP("memory", "m", "100m", "Memory limit")
	runCmd.Flags().StringP("pids", "p", "20", "Max number of processes")
	runCmd.Flags().StringP("image", "i", "", "Image to use")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(childCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(restoreCmd)

}

func executeCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
