package overlay

import (
	"bytes"
	"testing"
)

func TestWriteReadOverlay(t *testing.T) {
	baseBinary := []byte("FAKE-ELF-BINARY-CONTENT")
	payload := []byte("encrypted-payload-data-here")

	var buf bytes.Buffer
	if err := WriteOverlay(&buf, baseBinary, payload); err != nil {
		t.Fatalf("WriteOverlay failed: %v", err)
	}

	data := buf.Bytes()
	r := bytes.NewReader(data)

	got, err := ReadOverlay(r, int64(len(data)))
	if err != nil {
		t.Fatalf("ReadOverlay failed: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: got %q, want %q", got, payload)
	}
}

func TestReadOverlay_NoMagic(t *testing.T) {
	data := []byte("just a regular binary with no overlay at all here")
	r := bytes.NewReader(data)

	got, err := ReadOverlay(r, int64(len(data)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil payload, got %q", got)
	}
}

func TestReadOverlay_TooSmall(t *testing.T) {
	data := []byte("tiny")
	r := bytes.NewReader(data)

	_, err := ReadOverlay(r, int64(len(data)))
	if err == nil {
		t.Fatal("expected error for file too small")
	}
}

func TestWriteReadOverlay_LargePayload(t *testing.T) {
	baseBinary := bytes.Repeat([]byte{0x7f}, 1024)
	payload := bytes.Repeat([]byte{0xAB}, 65536)

	var buf bytes.Buffer
	if err := WriteOverlay(&buf, baseBinary, payload); err != nil {
		t.Fatalf("WriteOverlay failed: %v", err)
	}

	data := buf.Bytes()
	r := bytes.NewReader(data)

	got, err := ReadOverlay(r, int64(len(data)))
	if err != nil {
		t.Fatalf("ReadOverlay failed: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Fatal("large payload mismatch")
	}
}

func TestWriteReadOverlay_BaseBinaryPreserved(t *testing.T) {
	baseBinary := []byte("MY-BINARY")
	payload := []byte("secret")

	var buf bytes.Buffer
	if err := WriteOverlay(&buf, baseBinary, payload); err != nil {
		t.Fatalf("WriteOverlay failed: %v", err)
	}

	data := buf.Bytes()
	if !bytes.HasPrefix(data, baseBinary) {
		t.Fatal("base binary not preserved at start of output")
	}
}
