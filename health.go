package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
