# diary

## day 1 - mar 16

set up dev environment. multipass VM on mac because apple silicon cant run linux namespaces natively. vs code remote ssh into ubuntu VM, go 1.22.

got basic process spawning working. `go run main.go run /bin/sh` starts a child sh. problem was spawned sh sees all host processes. need PID namespace so it only sees itself as PID 1.

fixed that, learned about /proc/self/exe re-exec pattern, why child process is needed etc.

UTS + PID + mount namespaces finished. chroot into alpine rootfs. /proc mount inside container so ps only shows container processes.

### first bug (1.5h)

mounted /proc inside the container but without a proper mount namespace, so it overwrote the host's /proc. suddenly nothing worked. go, ls, /proc/self/exe, all broken. fixed by mounting it back manually in terminal.

## day 2 - mar 17

cgroups v2. wrote a `cg()` function that creates `/sys/fs/cgroup/gobox`.

tested with fork bomb.

also added memory cgroup. memory.max set to 100MB. this was easy compared to day 1.

## day 3 - mar 18

networking was confusing at first but fun at the end.

 added CLONE_NEWNET so container starts with empty network namespace, only a dead loopback.

learned about veth pairs. virtual ethernet cable with two ends. one stays on host, other gets moved into container's namespace with `LinkSetNsPid()`. used vishvananda/netlink library.

assigned IPs. host side 10.10.10.1, container side 10.10.10.2. container can ping host. first tried 67.67.67.x but thats public IP space noooo. switched to 10.10.10.x private range.

had a race condition. child needs to wait for host to create and move the veth before it can configure its end. started with `time.Sleep(time.Second)` which felt wrong.

### evening session:

refactored single main.go into multiple files. main.go, network.go, cgroup.go, const.go. all still `package main`, just cleaner. dont see why i need folders yet, maybe later when theres more code.

set up cobra for CLI. first time making a real CLI tool. `./gobox run`, `./gobox ps`, `./gobox --help` all work now. cobra was surprisingly easy.

also learned difference between `net.ParseIP()` and `netlink.ParseAddr()` the hard way. one wants bare IP, other wants CIDR. spent too long on that.

added container state tracking. each container gets a random ID, saves a JSON file to `/var/lib/gobox/` with status, command, PID, and timestamp.

`gobox ps` now lists all containers for real. 

idk if i like oop in go still

each container gets its own cgroup path based on its ID. 
added CLI flags for memory and pids limits, `gobox run -m 200m -p 10 /bin/sh` works now.

added overlayfs. each container gets its own writable layer on top of a shared readonly alpine rootfs. files created in one container dont leak to others or to the base image. also restructured state to use per-container directories.

It was a BIG day, i think i have 6-7h today on this project, but i learned a lot!

## day 4 - mar 19

improved `gobox ps` so it reads real container state directories and prints clean columns for id/status/command/created.

added new CLI commands and implemented lifecycle helpers in state handling

stdout/stderr now goes both to terminal and per-container `log.txt` 

### evening session:
OCI IMAGE PULL!!!!!!!!!!!

this was actually so fun the whole pipiline and how they made id is impresive

special tnx to claude for generating getImageLayers function, i had bug and would get "found 0 layers:, because of the amd64 digest 

basically some images returns a manifest list, so wee need to check if we get layers of platfroms, then we can find platform digest, and send a new requst to get the actual layer list

## day 5 - mar 20

had a bug where container would start but immediately exit — `exec format error`. turned out i was pulling amd64 images on an arm64 VM. fixed by using `runtime.GOARCH` to auto-detect platform instead of hardcoding amd64.

learned how tar.gz extraction works in go

i think the first version is completed, still have to do things like exec, checkpoint and migrate and Goboxfile.

exec command added, now we can run `sudo ./gobox exec <id> command`, in some other bash for a running container