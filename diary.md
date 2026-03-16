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