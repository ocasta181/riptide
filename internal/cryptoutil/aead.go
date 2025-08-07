package cryptoutil

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync/atomic"

	"golang.org/x/crypto/chacha20poly1305"
)

type AEAD struct {
	aead   cipher.AEAD
	prefix [4]byte
	ctr    atomic.Uint64
}

func NewAEAD(key [32]byte) (*AEAD, error) {
	a, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, err
	}
	var p [4]byte
	_, _ = rand.Read(p[:])
	return &AEAD{aead: a, prefix: p}, nil
}

func (a *AEAD) Nonce() [12]byte {
	var n [12]byte
	copy(n[:4], a.prefix[:])
	binary.BigEndian.PutUint64(n[4:], a.ctr.Add(1))
	return n
}

func (a *AEAD) Seal(dst, plaintext, aad []byte) ([]byte, [12]byte) {
	n := a.Nonce()
	out := a.aead.Seal(dst, n[:], plaintext, aad)
	return out, n
}

func (a *AEAD) Open(dst, ciphertext, aad []byte, nonce [12]byte) ([]byte, error) {
	if len(ciphertext) < a.aead.Overhead() {
		return nil, errors.New("ciphertext too short")
	}
	out, err := a.aead.Open(dst, nonce[:], ciphertext, aad)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Overhead() int {
	return chacha20poly1305.Overhead
}
