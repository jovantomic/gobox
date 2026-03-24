# GoBox

container runtime from scratch in Go. why? sounded fun and as a extra project in OS class

## what works

- **namespaces**: PID, UTS, Mount, Network. container is fully isolated, sees only its own processes, has its own hostname, filesystem, network stack
- **cgroups v2:**  per-container PID and memory limits. two containers won't step on each other
- **overlayfs:** layered filesystem, base image stays clean, writes go to upper layer
- **networking:** veth pairs, container gets its own IP, can talk to host
- **OCI image pull:**  pulls images straight from Docker Hub. auth token, manifest resolution (handles multi-arch), layer download, tar.gz extraction. no docker needed
- **container lifecycle:**  run, ps, stop, rm, logs. state tracked in JSON files
```bash
sudo ./gobox pull alpine
sudo ./gobox run -i alpine /bin/sh
sudo ./gobox ps
sudo ./gobox checkpoint <id>
sudo ./gobox restore <id>
sudo ./gobox stop <id>
sudo ./gobox logs <id>
```

## stack

- Go 1.22
- Linux namespaces (CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNS, CLONE_NEWNET)
- cgroups v2 (pids + memory controllers)
- overlayfs for copy-on-write filesystem
- Docker Hub Registry API v2 for image pulling
- vishvananda/netlink for network setup
- Alpine Linux as default base image

## dev environment

multipass VM on mac — apple silicon can't run linux namespaces natively. vs code remote SSH into ubuntu VM.

[Developer diary](diary.md)
