package main

import (
	"testing"
)

func TestDetectPayload_NonePresent(t *testing.T) {
	// The test binary itself has no overlay appended.
	payload, err := detectPayload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != nil {
		t.Fatalf("expected nil payload on unpacked binary, got %d bytes", len(payload))
	}
}
