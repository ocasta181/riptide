package cryptoutil

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

type SessionKeys struct {
	TX [32]byte
	RX [32]byte
}

func GenerateEd25519() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func GenerateX25519() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PublicKey(), nil
}

func SharedSecret(priv *ecdh.PrivateKey, pub *ecdh.PublicKey) ([]byte, error) {
	return priv.ECDH(pub)
}

func DeriveSession(shared, salt []byte, initiator bool) SessionKeys {
	r1 := hkdf.New(sha256.New, shared, salt, []byte("riptide/session/k1"))
	r2 := hkdf.New(sha256.New, shared, salt, []byte("riptide/session/k2"))
	var k1, k2 [32]byte
	_, _ = io.ReadFull(r1, k1[:])
	_, _ = io.ReadFull(r2, k2[:])
	if initiator {
		return SessionKeys{TX: k1, RX: k2}
	}
	return SessionKeys{TX: k2, RX: k1}
}

func Sign(priv ed25519.PrivateKey, msg []byte) []byte {
	return ed25519.Sign(priv, msg)
}

func Verify(pub ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(pub, msg, sig)
}
