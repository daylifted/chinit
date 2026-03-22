package main

import (
	"testing"
)

func TestBuildArgv_Simple(t *testing.T) {
	argv, err := buildArgv([]string{"/bin/echo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(argv) != 1 || argv[0] != "/bin/echo" {
		t.Fatalf("unexpected argv: %v", argv)
	}
}

func TestBuildArgv_WithArgs(t *testing.T) {
	argv, err := buildArgv([]string{"/bin/echo hello world"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/bin/echo", "hello", "world"}
	for i, v := range want {
		if argv[i] != v {
			t.Fatalf("argv[%d] = %q, want %q", i, argv[i], v)
		}
	}
}

func TestBuildArgv_ExtraArgs(t *testing.T) {
	argv, err := buildArgv([]string{"/bin/echo", "extra"})
	if err != nil {
		t.Fatal(err)
	}
	if len(argv) != 2 || argv[1] != "extra" {
		t.Fatalf("unexpected argv: %v", argv)
	}
}

func TestBuildArgv_Empty(t *testing.T) {
	if _, err := buildArgv(nil); err == nil {
		t.Fatal("expected error for nil args")
	}
	if _, err := buildArgv([]string{""}); err == nil {
		t.Fatal("expected error for empty command string")
	}
}
