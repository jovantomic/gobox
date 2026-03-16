# diary

## day 1 — mar 16

set up dev environment — multipass VM on mac (apple silicon cant run linux namespaces natively), vs code remote ssh, go 1.22.

got basic process spawning working. `go run main.go run /bin/bash` starts a child bash.

```
$ ps
  11229 pts/1    00:00:00 bash
  11293 pts/1    00:00:00 go
  11329 pts/1    00:00:00 main
  11333 pts/1    00:00:00 bash
  11351 pts/1    00:00:00 ps
```

problem: spawned bash sees all host processes. need PID namespace so it only sees itself as PID 1. 

fixed that learned processes, why i need childetc. 

UTS + PID + mount namespaces finished. chroot into alpine rootfs. /proc mount inside container so ps only shows container processes.

### first bug (1.5h of hating go and my life)

mounted /proc inside the container but without a proper mount namespace, so it overwrote the host's /proc. suddenly nothing worked — go, ls /proc/self/exe, all broken

fixed by mouting it back to host in terminal

next: cgroups for resource limits, then networking.