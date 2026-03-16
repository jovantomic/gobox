# GoBox

container runtime from scratch in Go. Why? It sounds fun and wanted to learn go

basically building what docker does under the hood 

namespaces, cgroups, overlayfs, networking.

## status

can spawn isolated processes with their own hostname, PID namespace, and mount namespace. chroot into alpine rootfs works. 

```bash
sudo ./gobox run /bin/sh
```

## dev log

[diary.md](diary.md)
