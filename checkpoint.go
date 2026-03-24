package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

	cmd := exec.Command("criu", "dump",
		"-t", fmt.Sprint(state.Pid),
		"-D", dumpDir,
		"--shell-job",
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
	merged := filepath.Join(stateDir, id, "merged")

	if !isMountPoint(merged) {
		lower := inferLowerRootfs(id, state)
		setupOverlay(id, lower)
	}

	pidFile := filepath.Join(stateDir, id, "restore.pid")
	_ = os.Remove(pidFile)
	criuLog := filepath.Join(stateDir, id, "criu-restore.log")
	_ = os.Remove(criuLog)

	fmt.Printf("Restoring container %s from checkpoint...\n", id)
	cmd := exec.Command("criu", "restore",
		"-D", dumpDir,
		"--shell-job",
		"--root", merged,
		"--pidfile", pidFile,
		"-o", criuLog,
		"-v4",
		"--manage-cgroups",
		"--restore-detached",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if logData, readErr := os.ReadFile(criuLog); readErr == nil {
			logText := string(logData)
			if strings.Contains(logText, "killed by signal 11") {
				return fmt.Errorf("failed to restore container: criu restorer crashed (SIGSEGV). This is likely a CRIU 3.16.1 compatibility issue on this kernel/arch; upgrade CRIU and retry (see %s)", criuLog)
			}
		}
		return fmt.Errorf("failed to restore container: %v (see %s)", err, criuLog)
	}

	if pidData, err := os.ReadFile(pidFile); err == nil {
		if pid, convErr := strconv.Atoi(strings.TrimSpace(string(pidData))); convErr == nil {
			state.Pid = pid
		}
	}

	state.Status = "running"
	saveJSON(state)
	fmt.Printf("Container %s restored successfully\n", id)
	return nil
}

func isMountPoint(path string) bool {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 4 && fields[4] == path {
			return true
		}
	}

	return false
}

func inferLowerRootfs(id string, state *ContainerState) string {
	if state.Image != "" {
		imgPath := filepath.Join(imagesDir, state.Image, "rootfs")
		if _, err := os.Stat(imgPath); err == nil {
			return imgPath
		}
	}

	return rootfsPath
}
