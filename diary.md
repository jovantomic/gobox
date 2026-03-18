# diary

## day 1 — mar 16

set up dev environment — multipass VM on mac (apple silicon cant run linux namespaces natively), vs code remote ssh, go 1.22.

got basic process spawning working. `go run main.go run /bin/bash` starts a child bash.

problem: spawned bash sees all host processes. need PID namespace so it only sees itself as PID 1.

fixed that, learned about /proc/self/exe re-exec pattern, why child process is needed etc.

UTS + PID + mount namespaces finished. chroot into alpine rootfs. /proc mount inside container so ps only shows container processes.

### first bug (1.5h)

mounted /proc inside the container but without a proper mount namespace, so it overwrote the host's /proc. suddenly nothing worked — go, ls, /proc/self/exe, all broken.

fixed by mounting it back manually in terminal.

## day 2 — mar 17

cgroups v2. wrote a `cg()` function that creates `/sys/fs/cgroup/gobox` and sets pids.max to 20.

tested with fork bomb — `for i in $(seq 1 25); do sleep 100 & done` — got `can't fork: Resource temporarily unavailable` after hitting the limit. cgroup works, kernel enforces it even though container cant see the cgroup files (theyre on host side, before chroot).

verified from host terminal: `pids.max = 20`, `pids.current = 5`, `cgroup.procs` shows the right PIDs. 

this was easy

## day 3 — mar 18

this was confusing at first... but fun at the end

networking. added CLONE_NEWNET — container starts with empty network namespace, only a dead loopback.

learned about veth pairs — virtual ethernet cable with two ends. one stays on host, other gets moved into container's namespace with `LinkSetNsPid()`. used vishvananda/netlink library.

assigned IPs: host side 10.10.10.1, container side 10.10.10.2. container can ping host.

first tried 67.67.67.x but thats public IP space noooo. switched to 10.10.10.x (private range).

added `time.Sleep(time.Second)` in childbecuse child needs to wait for host to create and move the veth before it can configure its end. Learning race conditioning before helped, probably has a better approach

also added memory cgroup — `memory.max` set to 100MB.
