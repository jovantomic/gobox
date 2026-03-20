package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	executeCLI()
}

func run(args []string, memory string, pids string, image string) {

	imgPath := rootfsPath
	if image != "" {
		imgPath = filepath.Join(imagesDir, image, "rootfs")
	}

	state := newContainerState(args[0])
	state.Status = "running"
	saveJSON(state)
	fmt.Printf("Container %s started\n", state.Id)

	fmt.Println("Running the application...", args, "PID:", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child", state.Id, memory, pids, imgPath}, args...)...)

	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	logFile, _ := os.Create(filepath.Join(stateDir, state.Id, "log.txt"))
	defer logFile.Close()
	cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
	cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	must(cmd.Start())
	state.Pid = cmd.Process.Pid
	saveJSON(state)

	setupHostNet(cmd.Process.Pid)

	cmd.Wait()
	cleanupOverlay(state.Id)

	state.Status = "stopped"
	saveJSON(state)
	cleanupNet()
	os.Remove("/sys/fs/cgroup/gobox_" + state.Id)
}

func child(args []string) {
	fmt.Println("Running the application...", args, "PID:", os.Getpid())
	id := args[0]
	memory := args[1]
	pids := args[2]
	imgPath := args[3]
	args = args[4:]

	cg(id, memory, pids)
	setupContainerNet()

	must(syscall.Sethostname([]byte(hostname)))
	//fmt.Println("DEBUG: imgPath =", imgPath)
	merged := setupOverlay(id, imgPath)
	//fmt.Println("DEBUG: merged =", merged)

	//fmt.Println("DEBUG: merged contents:", len(entries), "entries")

	must(syscall.Chroot(merged))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	//fmt.Println("DEBUG: cmd.Run error:", err)
}
