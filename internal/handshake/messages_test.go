package handshake

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestHelloEncodeDecode(t *testing.T) {
	h := NewHello(1, 0x1234)
	enc := h.Encode()
	out, err := DecodeHello(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.Version != h.Version || out.Caps != h.Caps || !bytes.Equal(out.Nonce[:], h.Nonce[:]) {
		t.Fatalf("mismatch")
	}
}

func TestKXEncodeDecode(t *testing.T) {
	pub := make([]byte, 32)
	_, _ = rand.Read(pub)
	k := KX{Public: pub}
	enc := k.Encode()
	out, err := DecodeKX(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if !bytes.Equal(out.Public, k.Public) {
		t.Fatalf("mismatch")
	}
}

func TestAuthEncodeDecode(t *testing.T) {
	pub := make([]byte, 32)
	sig := make([]byte, 64)
	_, _ = rand.Read(pub)
	_, _ = rand.Read(sig)
	a := Auth{Ed25519Pub: pub, Signature: sig}
	enc := a.Encode()
	out, err := DecodeAuth(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if !bytes.Equal(out.Ed25519Pub, a.Ed25519Pub) || !bytes.Equal(out.Signature, a.Signature) {
		t.Fatalf("mismatch")
	}
}

func TestSessionEncodeDecode(t *testing.T) {
	s := Session{MTU: 1400}
	enc := s.Encode()
	out, err := DecodeSession(enc)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if out.MTU != s.MTU {
		t.Fatalf("mismatch")
	}
}

func TestTranscriptDeterministic(t *testing.T) {
	p1 := []byte("hello")
	p2 := []byte("world")
	t1 := Transcript(p1, p2)
	t2 := Transcript(p1, p2)
	if !bytes.Equal(t1[:], t2[:]) {
		t.Fatalf("expected same")
	}
	t3 := Transcript(p2, p1)
	if bytes.Equal(t1[:], t3[:]) {
		t.Fatalf("expected different for different order")
	}
}
