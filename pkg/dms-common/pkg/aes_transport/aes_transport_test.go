package aes_transport

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()
	plain := "transport-aes-roundtrip"
	cipher, err := EncryptForTest(plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if cipher == "" || cipher == plain {
		t.Fatalf("expected non-empty ciphertext distinct from plaintext")
	}
	got, err := DecryptSecretPassword(cipher)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestDecryptEmpty(t *testing.T) {
	t.Parallel()
	if _, err := DecryptSecretPassword(""); err == nil {
		t.Fatal("expected incomplete ciphertext error")
	}
}

func TestDecryptGarbage(t *testing.T) {
	t.Parallel()
	if _, err := DecryptSecretPassword("not-valid-base64!!!"); err == nil {
		t.Fatal("expected decrypt failure")
	}
}
