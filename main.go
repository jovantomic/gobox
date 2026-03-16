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
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("wooooo")
	}
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
			syscall.CLONE_NEWNS,

		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())

}

func child() {
	fmt.Println("Running the application...", os.Args[2:], "PID:", os.Getpid())

	must(syscall.Sethostname([]byte("gobox")))
	must(syscall.Chroot("/home/ubuntu/gobox/rootfs"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

}
