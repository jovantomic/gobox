package main

import (
	"os"
	"path/filepath"
	"strconv"
)

func cg(id string, memory string, pids string) {
	cgroup := "/sys/fs/cgroup/gobox_" + id
	err := os.Mkdir(cgroup, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+pids +memory"), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "pids.max"), []byte(pids), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "memory.max"), []byte(memory), 0700))
	must(os.WriteFile(filepath.Join(cgroup, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
