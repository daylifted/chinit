package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	errBadArgs      = errors.New("bad arguments")
	errNon2xx       = errors.New("non-2xx response")
	errGrepMiss     = errors.New("grep pattern not found")
	errMissingValue = errors.New("flag missing value")
	errBadDuration  = errors.New("invalid duration")
)

const usage = `chinit — Swiss Army knife for scratch containers

Usage:
  sh -c <command>                                          exec mode
  sh <url> [--grep p] [--debug] [-k] [-ttl d]              health check
  sh pack -input <binary> -output <file> [-key <secret>]   pack a binary
  sh (no args, with embedded payload)                       run embedded payload
  sh --help                                                 show help

Flags (health check mode):
  --grep <pattern>   exit 1 if pattern not found in response body
  --debug            print response body to stdout
  -k                 skip TLS certificate verification
  -ttl <duration>    request timeout (default 3s)

Flags (pack mode):
  -input <binary>    input binary to encrypt and pack
  -output <file>     output packed binary
  -key <secret>      encryption key (default: PACKED_KEY env or system-derived)

At runtime, set PACKED_KEY env var to provide the decryption key.

Examples:
  sh -c /usr/bin/myapp
  sh http://service:9100/metrics
  sh https://service/health --grep "ready for work"
  sh pack -input ./myapp -output ./packed-sh -key "$SECRET"
`

func printHelp() {
	fmt.Print(usage)
}

func fail() {
	fmt.Println("error")
	os.Exit(1)
}

func main() {
	// --help anywhere
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		printHelp()
		return
	}

	// -c: exec mode
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		argv, err := buildArgv(os.Args[2:])
		if err != nil {
			fail()
		}
		path, err := exec.LookPath(argv[0])
		if err != nil {
			fail()
		}
		if err := syscall.Exec(path, argv, os.Environ()); err != nil {
			fail()
		}
		return
	}

	// URL prefix: health check mode
	if len(os.Args) > 1 && (strings.HasPrefix(os.Args[1], "http://") || strings.HasPrefix(os.Args[1], "https://")) {
		cfg, err := parseFlags(os.Args[2:])
		if err != nil {
			fail()
		}
		client := buildClient(cfg.skipTLS, cfg.timeout)
		if err := httpCheck(client, os.Args[1], cfg.grepPattern, cfg.debug); err != nil {
			fail()
		}
		return
	}

	// pack subcommand
	if len(os.Args) > 1 && os.Args[1] == "pack" {
		if err := packCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pack: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// no recognized flag: check for embedded payload
	encryptedPayload, err := detectPayload()
	if err != nil {
		fmt.Fprintf(os.Stderr, "payload detection: %v\n", err)
		os.Exit(1)
	}
	if encryptedPayload != nil {
		if err := runPayload(encryptedPayload); err != nil {
			fmt.Fprintf(os.Stderr, "payload: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// nothing matched
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}
	fail()
}
