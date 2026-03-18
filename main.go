package main

import (
	"fmt"
	"os"
	"os/exec"
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

func run(args []string, memory string, pids string) {
	state := newContainerState(args[0])
	state.Status = "running"
	saveJSON(state)
	fmt.Printf("Container %s started\n", state.Id)

	fmt.Println("Running the application...", args, "PID:", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child", state.Id, memory, pids}, args...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Start())
	state.Pid = cmd.Process.Pid
	saveJSON(state)

	setupHostNet(cmd.Process.Pid)

	cmd.Wait()

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
	args = args[3:]

	cg(id, memory, pids)
	setupContainerNet()

	must(syscall.Sethostname([]byte(hostname)))
	must(syscall.Chroot(rootfsPath))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
