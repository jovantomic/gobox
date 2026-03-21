package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func checkpointContainer(id string) error {
	state := getContainerById(id)
	if state == nil {
		return fmt.Errorf("container not found")
	}
	if state.Status != "running" {
		return fmt.Errorf("container is not running")
	}

	dumpDir := filepath.Join(stateDir, id, "checkpoint")
	os.MkdirAll(dumpDir, 0755)

	merged := filepath.Join("/var/lib/gobox/containers", id, "merged")
	cmd := exec.Command("criu", "dump",
		"-t", fmt.Sprint(state.Pid),
		"-D", dumpDir,
		"--shell-job",
		"--root", merged,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("criu dump failed: %v", err)
	}

	state.Status = "checkpointed"
	logFile := filepath.Join(stateDir, id, "log.txt")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "Container %s checkpointed\n", id)
	}

	saveJSON(state)
	fmt.Printf("Container %s checkpointed\n", id)
	return nil
}

func restoreContainer(id string) error {
	state := getContainerById(id)
	if state == nil {
		return fmt.Errorf("container not found")
	}
	if state.Status != "checkpointed" {
		return fmt.Errorf("container is not checkpointed")
	}

	dumpDir := filepath.Join(stateDir, id, "checkpoint")
	merged := filepath.Join("/var/lib/gobox/containers", id, "merged")

	fmt.Printf("Restoring container %s from checkpoint...\n", id)
	cmd := exec.Command("criu", "restore",
		"-D", dumpDir,
		"--shell-job",
		"--root", merged,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore container: %v", err)
	}

	state.Status = "running"
	saveJSON(state)
	fmt.Printf("Container %s restored successfully\n", id)
	return nil
}
