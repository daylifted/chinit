package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	key2, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey second call failed: %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("Key derivation is not deterministic")
	}
}

func TestDeriveKeyFromSecret(t *testing.T) {
	key := DeriveKeyFromSecret("test-secret")
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	key2 := DeriveKeyFromSecret("test-secret")
	if !bytes.Equal(key, key2) {
		t.Error("DeriveKeyFromSecret is not deterministic")
	}

	key3 := DeriveKeyFromSecret("different-secret")
	if bytes.Equal(key, key3) {
		t.Error("Different secrets produced the same key")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"Empty", []byte{}},
		{"Short", []byte("Hello, World!")},
		{"Long", bytes.Repeat([]byte("A"), 10000)},
		{"Binary", []byte{0x00, 0xFF, 0x01, 0xFE}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tc.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if len(tc.plaintext) > 0 && bytes.Equal(ciphertext, tc.plaintext) {
				t.Error("Ciphertext equals plaintext")
			}

			decrypted, err := Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Error("Decrypted text doesn't match original")
			}
		})
	}
}

func TestEncryptRandomness(t *testing.T) {
	key, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	plaintext := []byte("Same plaintext")

	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("First encrypt failed: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Second encrypt failed: %v", err)
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Encrypting same plaintext twice produced identical ciphertext")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := DeriveKey()
	key2 := make([]byte, 32)
	copy(key2, key1)
	key2[0] ^= 0xFF

	plaintext := []byte("Secret message")

	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("Expected decryption with wrong key to fail")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	shortKey := []byte("too short")
	plaintext := []byte("test")

	_, err := Encrypt(plaintext, shortKey)
	if err == nil {
		t.Error("Expected error with short key")
	}

	_, err = Decrypt(plaintext, shortKey)
	if err == nil {
		t.Error("Expected error with short key")
	}
}
