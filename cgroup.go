package main

import (
	"os"
	"path/filepath"
	"strconv"
)

func cg() {
	cgroup := cgroupPath
	err := os.Mkdir(cgroup, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+pids +memory"), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "pids.max"), []byte(pidsLimit), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "memory.max"), []byte(memoryLimit), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
