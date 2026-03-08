package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	errBadArgs      = errors.New("bad arguments")
	errNon2xx       = errors.New("non-2xx response")
	errGrepMiss     = errors.New("grep pattern not found")
	errMissingValue = errors.New("flag missing value")
	errBadDuration  = errors.New("invalid duration")
)

const usage = `Usage:
  sh -c <command>                       exec a command directly
  sh <url> [flags]                      HTTP health-check

Flags (HTTP mode):
  --grep <pattern>   exit 1 if pattern not found in response body
  --debug            print response body to stdout
  -k                 skip TLS certificate verification
  -ttl <duration>    request timeout (default 3s)

  --help             show this help

Examples:
  sh -c /usr/bin/myapp
  sh http://service:9100/metrics
  sh https://service/health --grep "ready for work"
  sh https://service/health --debug -k -ttl 10s
`

func printHelp() {
	fmt.Print(usage)
}


func fail() {
	fmt.Println("error")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fail()
	}

	if os.Args[1] == "--help" {
		printHelp()
		return
	}

	switch {
	case os.Args[1] == "-c":
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
	case strings.HasPrefix(os.Args[1], "http://") || strings.HasPrefix(os.Args[1], "https://"):
		cfg, err := parseFlags(os.Args[2:])
		if err != nil {
			fail()
		}
		client := buildClient(cfg.skipTLS, cfg.timeout)
		if err := httpCheck(client, os.Args[1], cfg.grepPattern, cfg.debug); err != nil {
			fail()
		}
	default:
		fail()
	}
}

// buildArgv splits the command string and appends any extra args.
func buildArgv(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, errBadArgs
	}
	parts := strings.Fields(args[0])
	if len(parts) == 0 {
		return nil, errBadArgs
	}
	return append(parts, args[1:]...), nil
}

type httpConfig struct {
	grepPattern string
	skipTLS     bool
	debug       bool
	timeout     time.Duration
}

// parseFlags parses the optional flags that follow a URL argument.
func parseFlags(args []string) (httpConfig, error) {
	cfg := httpConfig{timeout: 3 * time.Second}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--debug":
			cfg.debug = true
		case "--grep":
			if i+1 >= len(args) {
				return cfg, errMissingValue
			}
			i++
			cfg.grepPattern = args[i]
		case "-k":
			cfg.skipTLS = true
		case "-ttl":
			if i+1 >= len(args) {
				return cfg, errMissingValue
			}
			i++
			d, err := time.ParseDuration(args[i])
			if err != nil {
				return cfg, errBadDuration
			}
			cfg.timeout = d
		default:
			return cfg, errBadArgs
		}
	}
	return cfg, nil
}

// buildClient constructs an http.Client with the given TLS and timeout settings.
func buildClient(skipTLS bool, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS}, //nolint:gosec
		},
	}
}

// httpCheck performs a GET request and validates the response status and optional grep pattern.
// When debug is true the full response body is printed to stdout.
func httpCheck(client *http.Client, url, grepPattern string, debug bool) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errNon2xx
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if debug {
		fmt.Print(string(body))
	}

	if grepPattern != "" && !strings.Contains(string(body), grepPattern) {
		return errGrepMiss
	}
	fmt.Println("chinit: OK")
	return nil
}
