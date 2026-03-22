package overlay

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Magic is the 8-byte trailer that identifies a chinit overlay.
var Magic = [8]byte{'C', 'H', 'I', 'N', 'I', 'T', 0x00, 0x01}

const trailerSize = 16 // 8 bytes length + 8 bytes magic

// WriteOverlay writes a chinit binary with an appended encrypted payload.
// Format: [chinit ELF][encrypted payload][uint64 payload length BE][magic 8 bytes]
func WriteOverlay(w io.Writer, chinitBinary, encryptedPayload []byte) error {
	if _, err := w.Write(chinitBinary); err != nil {
		return fmt.Errorf("failed to write base binary: %w", err)
	}

	if _, err := w.Write(encryptedPayload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	lenBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBuf, uint64(len(encryptedPayload)))
	if _, err := w.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	if _, err := w.Write(Magic[:]); err != nil {
		return fmt.Errorf("failed to write magic: %w", err)
	}

	return nil
}

// ReadOverlay reads the trailer of a file and extracts the encrypted payload.
// It reads from an io.ReaderAt using the given file size.
func ReadOverlay(r io.ReaderAt, fileSize int64) ([]byte, error) {
	if fileSize < trailerSize {
		return nil, fmt.Errorf("file too small for overlay")
	}

	trailer := make([]byte, trailerSize)
	if _, err := r.ReadAt(trailer, fileSize-trailerSize); err != nil {
		return nil, fmt.Errorf("failed to read trailer: %w", err)
	}

	var magic [8]byte
	copy(magic[:], trailer[8:])
	if magic != Magic {
		return nil, nil // no overlay present
	}

	payloadLen := binary.BigEndian.Uint64(trailer[:8])
	if payloadLen == 0 {
		return nil, fmt.Errorf("overlay payload length is zero")
	}

	totalOverlay := int64(payloadLen) + trailerSize
	if totalOverlay > fileSize {
		return nil, fmt.Errorf("overlay size exceeds file size")
	}

	payload := make([]byte, payloadLen)
	offset := fileSize - totalOverlay
	if _, err := r.ReadAt(payload, offset); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	return payload, nil
}
