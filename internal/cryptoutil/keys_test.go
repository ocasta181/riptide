package cryptoutil

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestX25519SharedSecretAndDeriveSession(t *testing.T) {
	privA, pubA, err := GenerateX25519()
	if err != nil {
		t.Fatalf("gen A: %v", err)
	}
	privB, pubB, err := GenerateX25519()
	if err != nil {
		t.Fatalf("gen B: %v", err)
	}
	secAB, err := SharedSecret(privA, pubB)
	if err != nil {
		t.Fatalf("secret AB: %v", err)
	}
	secBA, err := SharedSecret(privB, pubA)
	if err != nil {
		t.Fatalf("secret BA: %v", err)
	}
	if !bytes.Equal(secAB, secBA) {
		t.Fatalf("shared secrets mismatch")
	}
	salt := []byte("riptide-test-salt")
	keysA := DeriveSession(secAB, salt, true)
	keysB := DeriveSession(secBA, salt, false)
	if !bytes.Equal(keysA.TX[:], keysB.RX[:]) || !bytes.Equal(keysA.RX[:], keysB.TX[:]) {
		t.Fatalf("session key pairing mismatch")
	}
	if bytes.Equal(keysA.TX[:], keysA.RX[:]) {
		t.Fatalf("tx and rx should differ")
	}
}

func TestEd25519SignVerify(t *testing.T) {
	pub, priv, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	msg := []byte("hello")
	sig := Sign(priv, msg)
	if !Verify(pub, msg, sig) {
		t.Fatalf("verify failed")
	}
	bad := []byte("bye")
	if Verify(pub, bad, sig) {
		t.Fatalf("verify should fail on different msg")
	}
}

func TestAEADSealOpen(t *testing.T) {
	var key [32]byte
	_, _ = rand.Read(key[:])
	a, err := NewAEAD(key)
	if err != nil {
		t.Fatalf("aead: %v", err)
	}
	pt := []byte("plaintext")
	aad := []byte("aad")
	ct, nonce := a.Seal(nil, pt, aad)
	out, err := a.Open(nil, ct, aad, nonce)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !bytes.Equal(out, pt) {
		t.Fatalf("mismatch")
	}
	_, err = a.Open(nil, ct, []byte("aad2"), nonce)
	if err == nil {
		t.Fatalf("expected failure with wrong aad")
	}
}
