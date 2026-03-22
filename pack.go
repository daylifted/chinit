package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/daylift/chinit/pkg/crypto"
	"github.com/daylift/chinit/pkg/overlay"
)

// packCmd runs the pack subcommand: encrypt a binary and append it to a copy of ourselves.
func packCmd(args []string) error {
	var inputPath, outputPath, keySecret string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-input":
			if i+1 >= len(args) {
				return errMissingValue
			}
			i++
			inputPath = args[i]
		case "-output":
			if i+1 >= len(args) {
				return errMissingValue
			}
			i++
			outputPath = args[i]
		case "-key":
			if i+1 >= len(args) {
				return errMissingValue
			}
			i++
			keySecret = args[i]
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("usage: sh pack -input <binary> -output <file> [-key <secret>]")
	}

	binaryData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	var key []byte
	if keySecret != "" {
		key = crypto.DeriveKeyFromSecret(keySecret)
	} else {
		key, err = crypto.DeriveKey()
		if err != nil {
			return fmt.Errorf("failed to derive key: %w", err)
		}
	}

	encryptedData, err := crypto.Encrypt(binaryData, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find own executable: %w", err)
	}

	selfBinary, err := os.ReadFile(selfPath)
	if err != nil {
		return fmt.Errorf("failed to read own executable: %w", err)
	}

	var buf bytes.Buffer
	if err := overlay.WriteOverlay(&buf, selfBinary, encryptedData); err != nil {
		return fmt.Errorf("failed to write overlay: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0555); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fingerprint, _ := crypto.GetKeyFingerprint()
	fmt.Printf("packed %s -> %s (key fingerprint: %s)\n", inputPath, outputPath, fingerprint)
	return nil
}
