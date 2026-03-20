package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// status could be created, running, stopped
// (we can make enum but i dont know how to do that in go :D)

type ContainerState struct {
	Id      string    `json:"id"`
	Status  string    `json:"status"`
	Command string    `json:"command"`
	Created time.Time `json:"created"`
	Pid     int       `json:"pid,omitempty"`
}

// charset is in const.go
func generateId() string {
	for {
		b := make([]byte, 8)
		for i := range b {
			b[i] = charset[rand.Int63()%int64(len(charset))]
		}
		id := string(b)
		if getContainerById(id) == nil {
			return id
		}
	}
}

func newContainerState(command string) *ContainerState {
	return &ContainerState{
		Id:      generateId(),
		Status:  "created",
		Command: command,
		Created: time.Now(),
	}
}

func getContainerById(id string) *ContainerState {
	data, err := os.ReadFile(filepath.Join(stateDir, id, "state.json"))
	if err != nil {
		return nil
	}
	var state ContainerState
	json.Unmarshal(data, &state)
	return &state
}

func saveJSON(state *ContainerState) {
	dir := filepath.Join(stateDir, state.Id)
	os.MkdirAll(dir, 0755)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		panic(err)
	}
	must(os.WriteFile(filepath.Join(dir, "state.json"), data, 0644))
}

func deleteContainerState(id string) {

	os.RemoveAll(filepath.Join(stateDir, id))
}

func showLogs(id string) {
	state := getContainerById(id)
	if state == nil {
		fmt.Printf("Container with ID %s not found\n", id)
		return
	}
	logData, err := os.ReadFile(filepath.Join(stateDir, id, "log.txt"))
	if err != nil {
		fmt.Printf("No logs found for container %s\n", id)
		return
	}
	fmt.Printf("Logs for container %s:\n%s\n", id, string(logData))
}

func stopContainer(id string) {
	state := getContainerById(id)
	if state == nil {
		fmt.Printf("Container with ID %s not found\n", id)
		return
	}
	if state.Status != "running" {
		fmt.Printf("Container with ID %s is not running\n", id)
		return
	}
	syscall.Kill(state.Pid, syscall.SIGKILL)
	state.Status = "stopped"
	state.Pid = 0
	saveJSON(state)
}

func removeContainer(id string) {
	state := getContainerById(id)
	if state == nil {
		fmt.Printf("Container with ID %s not found\n", id)
		return
	}
	if state.Status == "running" {
		fmt.Printf("Container with ID %s is running, stop it before removing\n", id)
		return
	}
	deleteContainerState(id)
	fmt.Printf("Container %s removed\n", id)

}

func getAllContainers() []*ContainerState {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return nil
	}
	var containers []*ContainerState
	for _, entry := range entries {
		if entry.IsDir() {
			state := getContainerById(entry.Name())
			if state != nil {
				containers = append(containers, state)
			}
		}
	}
	return containers
}

func execContainer(id string, command []string) {
	state := getContainerById(id)
	if state == nil {
		fmt.Printf("Container with ID %s not found\n", id)
		return
	}
	pid := fmt.Sprint(state.Pid)
	args := []string{"-t", pid, "-m", "-u", "-p", "-n", "--"}
	args = append(args, command...)

	cmd := exec.Command("nsenter", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
