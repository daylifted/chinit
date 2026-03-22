package main

import (
	"testing"
)

func TestPackCmd_MissingInput(t *testing.T) {
	err := packCmd([]string{"-output", "/tmp/out"})
	if err == nil {
		t.Fatal("expected error for missing -input")
	}
}

func TestPackCmd_MissingOutput(t *testing.T) {
	err := packCmd([]string{"-input", "/bin/ls"})
	if err == nil {
		t.Fatal("expected error for missing -output")
	}
}

func TestPackCmd_UnknownFlag(t *testing.T) {
	err := packCmd([]string{"-input", "/bin/ls", "-output", "/tmp/out", "--bad"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestPackCmd_MissingFlagValue(t *testing.T) {
	err := packCmd([]string{"-input"})
	if err == nil {
		t.Fatal("expected error for missing flag value")
	}
}
