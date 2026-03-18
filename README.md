# GoBox

container runtime from scratch in Go. why? sounded fun and wanted to learn go

basically building what docker does under the hood — namespaces, cgroups, networking.

## what works

- PID namespace — container only sees its own processes
- UTS namespace — container gets its own hostname
- Mount namespace — isolated /proc, no leaking into host
- Network namespace — veth pair, container has its own network stack
- Chroot — alpine rootfs, fully isolated filesystem
- Cgroups v2 — PID limit (20 processes), memory limit (100MB)
```bash
sudo ./gobox run /bin/sh
```

## stack

- Go 1.22
- Linux namespaces (CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNS, CLONE_NEWNET)
- cgroups v2 (pids + memory controllers)
- vishvananda/netlink for network setup
- Alpine Linux rootfs

## dev environment

multipass VM on mac — apple silicon can't run linux namespaces natively. vs code remote SSH into ubuntu VM.