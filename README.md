# GoBox

container runtime from scratch in Go. Why? It sounds fun and wanted to learn go

basically building what docker does under the hood 

namespaces, cgroups, overlayfs, networking.

## status

just started. can spawn a child process, next up is actual isolation.

```bash
sudo ./gobox run /bin/sh
```

## dev log

[diary.md](diary.md)
