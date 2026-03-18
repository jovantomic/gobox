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

func run() {
	fmt.Println("Running the application...", os.Args[2:], "PID:", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
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

	setupHostNet(cmd.Process.Pid)

	cmd.Wait()

	cleanupNet()
	os.Remove(cgroupPath)
}

func child() {
	fmt.Println("Running the application...", os.Args[2:], "PID:", os.Getpid())

	cg()

	setupContainerNet()

	must(syscall.Sethostname([]byte(hostname)))
	must(syscall.Chroot(rootfsPath))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
