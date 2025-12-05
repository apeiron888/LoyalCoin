package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateDEK(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate DEK: %v", err)
	}

	if len(dek) != 32 {
		t.Errorf("Expected DEK length 32, got %d", len(dek))
	}

	// Generate another to ensure they're different
	dek2, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate second DEK: %v", err)
	}

	if bytes.Equal(dek, dek2) {
		t.Error("Two generated DEKs should not be identical")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("my-secret-private-key-12345")
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate DEK: %v", err)
	}

	// Encrypt
	blob, err := EncryptWithDEK(plaintext, dek)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Verify blob fields
	if blob.Version != 1 {
		t.Errorf("Expected version 1, got %d", blob.Version)
	}

	if blob.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm AES-256-GCM, got %s", blob.Algorithm)
	}

	if blob.Nonce == "" || blob.Ciphertext == "" || blob.Tag == "" {
		t.Error("Blob fields should not be empty")
	}

	// Decrypt
	decrypted, err := DecryptWithDEK(blob, dek)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted data doesn't match original.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

func TestEncryptedBlobJSON(t *testing.T) {
	plaintext := []byte("test-private-key")
	dek, _ := GenerateDEK()
	defer ZeroBytes(dek)

	blob, err := EncryptWithDEK(plaintext, dek)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Convert to JSON
	jsonStr, err := EncryptedBlobToJSON(blob)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Convert back from JSON
	blob2, err := JSONToEncryptedBlob(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify fields match
	if blob.Version != blob2.Version {
		t.Error("Version mismatch")
	}
	if blob.Nonce != blob2.Nonce {
		t.Error("Nonce mismatch")
	}
	if blob.Ciphertext != blob2.Ciphertext {
		t.Error("Ciphertext mismatch")
	}
}

func TestZeroBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("Byte at index %d should be 0, got %d", i, b)
		}
	}
}

func TestDecryptWrongDEK(t *testing.T) {
	plaintext := []byte("secret-data")
	dek1, _ := GenerateDEK()
	dek2, _ := GenerateDEK()

	blob, err := EncryptWithDEK(plaintext, dek1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with wrong DEK
	_, err = DecryptWithDEK(blob, dek2)
	if err == nil {
		t.Error("Decrypt should fail with wrong DEK")
	}
}
