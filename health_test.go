package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := parseFlags(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.grepPattern != "" || cfg.skipTLS || cfg.timeout != 3*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestParseFlags_Grep(t *testing.T) {
	cfg, err := parseFlags([]string{"--grep", "ready"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.grepPattern != "ready" {
		t.Fatalf("grepPattern = %q", cfg.grepPattern)
	}
}

func TestParseFlags_SkipTLS(t *testing.T) {
	cfg, err := parseFlags([]string{"-k"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.skipTLS {
		t.Fatal("expected skipTLS=true")
	}
}

func TestParseFlags_TTL(t *testing.T) {
	cfg, err := parseFlags([]string{"-ttl", "10s"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.timeout != 10*time.Second {
		t.Fatalf("timeout = %v", cfg.timeout)
	}
}

func TestParseFlags_Debug(t *testing.T) {
	cfg, err := parseFlags([]string{"--debug"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.debug {
		t.Fatal("expected debug=true")
	}
}

func TestParseFlags_AllFlags(t *testing.T) {
	cfg, err := parseFlags([]string{"--grep", "ok", "-k", "-ttl", "5s", "--debug"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.grepPattern != "ok" || !cfg.skipTLS || cfg.timeout != 5*time.Second || !cfg.debug {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	if _, err := parseFlags([]string{"--unknown"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestParseFlags_GrepMissingValue(t *testing.T) {
	if _, err := parseFlags([]string{"--grep"}); err == nil {
		t.Fatal("expected error when --grep has no value")
	}
}

func TestParseFlags_TTLMissingValue(t *testing.T) {
	if _, err := parseFlags([]string{"-ttl"}); err == nil {
		t.Fatal("expected error when -ttl has no value")
	}
}

func TestParseFlags_TTLInvalidDuration(t *testing.T) {
	if _, err := parseFlags([]string{"-ttl", "notaduration"}); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestHTTPCheck_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	if err := httpCheck(srv.Client(), srv.URL, "", false); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestHTTPCheck_Non2xx(t *testing.T) {
	for _, code := range []int{400, 404, 500, 503} {
		code := code
		t.Run(fmt.Sprintf("%d", code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()

			if err := httpCheck(srv.Client(), srv.URL, "", false); err != errNon2xx {
				t.Fatalf("expected errNon2xx, got: %v", err)
			}
		})
	}
}

func TestHTTPCheck_GrepFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ready for work")
	}))
	defer srv.Close()

	if err := httpCheck(srv.Client(), srv.URL, "ready for work", false); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestHTTPCheck_GrepNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not ready")
	}))
	defer srv.Close()

	if err := httpCheck(srv.Client(), srv.URL, "ready for work", false); err != errGrepMiss {
		t.Fatalf("expected errGrepMiss, got: %v", err)
	}
}

func TestHTTPCheck_DebugPrintsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "debug output")
	}))
	defer srv.Close()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := httpCheck(srv.Client(), srv.URL, "", true)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "debug outputchinit: OK\n" {
		t.Fatalf("stdout = %q, want %q", string(out), "debug outputchinit: OK\n")
	}
}

func TestHTTPCheck_NoDebugSilent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "silent")
	}))
	defer srv.Close()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := httpCheck(srv.Client(), srv.URL, "", false)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "chinit: OK\n" {
		t.Fatalf("stdout = %q, want %q", string(out), "chinit: OK\n")
	}
}

func TestHTTPCheck_NetworkError(t *testing.T) {
	client := &http.Client{Timeout: 100 * time.Millisecond}
	if err := httpCheck(client, "http://127.0.0.1:1", "", false); err == nil {
		t.Fatal("expected network error")
	}
}

func TestHTTPCheck_2xxRange(t *testing.T) {
	for _, code := range []int{200, 201, 204, 299} {
		code := code
		t.Run(fmt.Sprintf("%d", code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()

			if err := httpCheck(srv.Client(), srv.URL, "", false); err != nil {
				t.Fatalf("expected nil for %d, got: %v", code, err)
			}
		})
	}
}
