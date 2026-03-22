package main

import (
	"fmt"
	"os"

	"github.com/daylift/chinit/pkg/crypto"
	"github.com/daylift/chinit/pkg/loader"
	"github.com/daylift/chinit/pkg/overlay"
)

// detectPayload reads /proc/self/exe and checks for an appended overlay.
// Returns the encrypted payload bytes, or nil if no overlay is present.
func detectPayload() ([]byte, error) {
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to find own executable: %w", err)
	}

	f, err := os.Open(selfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open own executable: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat own executable: %w", err)
	}

	return overlay.ReadOverlay(f, info.Size())
}

// runPayload decrypts the embedded payload and executes it in memory.
func runPayload(encryptedPayload []byte) error {
	key, err := crypto.DeriveKey()
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	plaintext, err := crypto.Decrypt(encryptedPayload, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return loader.ExecuteInMemory(plaintext, os.Args)
}
