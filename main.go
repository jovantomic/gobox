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
	state.Image = image
	state.Status = "running"
	saveJSON(state)
	fmt.Printf("Container %s started\n", state.Id)

	fmt.Println("Running the application...", args, "PID:", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child", state.Id, memory, pids, imgPath}, args...)...)
	stdinReader, stdinWriter, err := os.Pipe()
	must(err)

	cmd.Stdin = stdinReader
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
	stdinReader.Close()
	go func() {
		defer stdinWriter.Close()
		io.Copy(stdinWriter, os.Stdin)
	}()
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
	merged := setupOverlay(id, imgPath)

	putOld := filepath.Join(merged, ".pivot_old")
	os.MkdirAll(putOld, 0755)
	must(syscall.PivotRoot(merged, putOld))
	must(syscall.Chdir("/"))
	must(syscall.Unmount("/.pivot_old", syscall.MNT_DETACH))
	os.Remove("/.pivot_old")
	os.WriteFile("/etc/resolv.conf", []byte("nameserver 8.8.8.8\n"), 0644)
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	os.MkdirAll("/dev/pts", 0755)
	must(syscall.Mount("devpts", "/dev/pts", "devpts", 0, ""))

	syscall.Setsid()
	syscall.Exec(args[0], args, os.Environ())
}
