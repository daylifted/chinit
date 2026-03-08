# chinit

A static Go binary that replaces `/bin/sh` in minimal container images.

**Why?**
- Docker assumes `/bin/sh` exists and exec's everything through `/bin/sh -c`. chinit intercepts that.
- When used with a static app binary, only two files are needed in the image, which simplifies AppArmor profiles significantly.

## Modes

### Exec mode

Executes a command directly by replacing the current process (`syscall.Exec`). The binary is resolved via `PATH`.

```sh
chinit -c <command> [args...]
```

```sh
chinit -c /usr/bin/myapp
chinit -c myapp --port 8080
```

### HTTP health-check mode

Performs an HTTP GET request and exits non-zero on failure. Prints `chinit: OK` on success, `error` on failure.

```sh
chinit <url> [flags]
```

```sh
chinit https://service/health --grep "ready" --debug -k -ttl 5s
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--grep <pattern>` | — | Exit 1 if pattern not found in response body |
| `--debug` | — | Print response body to stdout |
| `-k` | — | Skip TLS certificate verification |
| `-ttl <duration>` | `3s` | Request timeout (`3s`, `500ms`, `1m`, …) |
| `--help` | — | Print usage |

## Install

```sh
go install github.com/daylift/chinit@latest
```

Or build a static binary locally:
```sh
make build GOARCH=arm64
```

## Docker example

Use chinit as both the init process and the health-check:

```dockerfile
FROM golang AS chinit
RUN go install github.com/daylift/chinit@latest

FROM scratch
COPY --from=chinit /go/bin/chinit /bin/sh
COPY myapp /usr/bin/myapp

ENTRYPOINT ["/bin/sh", "-c", "myapp"]

HEALTHCHECK --interval=10s --timeout=5s \
  CMD ["/bin/sh", "http://localhost:8080/health", "--grep", "ready"]
```

## TO DO

- Make a more flexible init-like process
  - Currently assumes the process stays in the foreground
  - Create PID file for backgrounded processes and periodically check they are alive.
