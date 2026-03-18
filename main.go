package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
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
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Start())

	setupHostNet(cmd.Process.Pid)

	cmd.Wait()

	cleanupNet()
	os.Remove("/sys/fs/cgroup/gobox")
}

func child() {
	fmt.Println("Running the application...", os.Args[2:], "PID:", os.Getpid())

	cg()

	setupContainerNet()

	must(syscall.Sethostname([]byte("gobox")))
	must(syscall.Chroot("/home/ubuntu/gobox/root"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func cg() {
	cgroup := "/sys/fs/cgroup/gobox"
	err := os.Mkdir(cgroup, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+pids +memory"), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "pids.max"), []byte("20"), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "memory.max"), []byte("100m"), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func setupHostNet(pid int) {
	la := netlink.NewLinkAttrs()
	la.Name = "host_veth"
	veth := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  "gobox_cont",
	}
	must(netlink.LinkAdd(veth))

	hostVeth, err := netlink.LinkByName("host_veth")
	if err != nil {
		panic(err)
	}
	//mora privatni IP da ne bi bilo konflikta sa stvarnim mrežama :(
	hostAddr, err := netlink.ParseAddr("10.10.10.1/24")
	if err != nil {
		panic(err)
	}
	must(netlink.AddrAdd(hostVeth, hostAddr))
	must(netlink.LinkSetUp(hostVeth))

	contVeth, err := netlink.LinkByName("gobox_cont")
	if err != nil {
		panic(err)
	}
	must(netlink.LinkSetNsPid(contVeth, pid))
}

func setupContainerNet() {
	time.Sleep(2 * time.Second)
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		panic(err)
	}
	must(netlink.LinkSetUp(lo))

	contVeth, err := netlink.LinkByName("gobox_cont")
	if err != nil {
		panic(err)
	}
	contAddr, err := netlink.ParseAddr("10.10.10.2/24")
	if err != nil {
		panic(err)
	}
	must(netlink.AddrAdd(contVeth, contAddr))
	must(netlink.LinkSetUp(contVeth))

	must(netlink.RouteAdd(&netlink.Route{
		Gw: net.ParseIP("10.10.10.1"),
	}))
}

func cleanupNet() {
	link, err := netlink.LinkByName("host_veth")
	if err == nil {
		netlink.LinkDel(link)
	}
}
