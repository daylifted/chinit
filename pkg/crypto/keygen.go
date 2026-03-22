package crypto

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/user"
	"strings"
)

// DeriveKey generates a 32-byte AES-256 key.
// Priority: 1. PACKED_KEY env var, 2. System identifiers (hostname + username + machine-id)
func DeriveKey() ([]byte, error) {
	if customKey := os.Getenv("PACKED_KEY"); customKey != "" {
		hash := sha256.Sum256([]byte(customKey))
		return hash[:], nil
	}

	var components []string

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	components = append(components, hostname)

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	components = append(components, currentUser.Username)

	machineID, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		machineID, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err != nil {
			return nil, fmt.Errorf("failed to read machine-id: %w", err)
		}
	}
	components = append(components, strings.TrimSpace(string(machineID)))

	combined := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(combined))

	return hash[:], nil
}

// DeriveKeyFromSecret generates a 32-byte AES-256 key from a provided secret string.
func DeriveKeyFromSecret(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// GetKeyFingerprint returns a short fingerprint of the derived key for debugging.
func GetKeyFingerprint() (string, error) {
	key, err := DeriveKey()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key[:8]), nil
}
