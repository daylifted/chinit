package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	s := string(out)
	for _, want := range []string{"--grep", "-ttl", "--debug", "pack", "-input", "-output", "-key", "PACKED_KEY"} {
		if !strings.Contains(s, want) {
			t.Fatalf("help output missing %q", want)
		}
	}
}
