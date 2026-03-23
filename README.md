# chinit

Swiss Army knife for scratch containers: `/bin/sh` replacement, health check, encrypting packer, and self-extracting loader in one static binary.

## Why

* Docker assumes `/bin/sh` exists and execs everything through `/bin/sh -c`. chinit intercepts that.
* With a static app binary, only two files are needed in the image, simplifying AppArmor profiles.
* The packer encrypts and embeds a binary into chinit itself, so the container image holds a single file.

## Modes

```
sh -c <command>                                          exec mode
sh <url> [--grep p] [--debug] [-k] [-ttl d]              health check
sh pack -input <binary> -output <file> [-key <secret>]   pack a binary
sh (no args, with embedded payload)                       run embedded payload
sh --help                                                 show help
```

## Build

```
make build
make test
```

## Health check flags

| Flag | Default | Description |
|------|---------|-------------|
| `--grep <pattern>` | — | Exit 1 if pattern not found in response body |
| `--debug` | — | Print response body to stdout |
| `-k` | — | Skip TLS certificate verification |
| `-ttl <duration>` | `3s` | Request timeout |

## Packing a binary

```
./sh pack -input ./myapp -output ./packed-sh -key "$SECRET"
```

At runtime, set `PACKED_KEY` to provide the decryption key. Without it, the key is derived from hostname + username + machine-id.

Health check and exec modes work even when a payload is embedded.

## Binary overlay format

```
[original chinit ELF binary]
[encrypted payload bytes]
[8 bytes: payload length as uint64 big-endian]
[8 bytes: magic marker "CHINIT\x00\x01"]
```

## Docker example

```dockerfile
FROM scratch
COPY packed-sh /bin/sh
ENTRYPOINT ["/bin/sh"]
HEALTHCHECK CMD ["/bin/sh", "http://localhost:8080/health", "--grep", "ok"]
```
