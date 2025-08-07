package handshake

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

type Hello struct {
	Version uint8
	Caps    uint16
	Nonce   [16]byte
}

func NewHello(version uint8, caps uint16) Hello {
	var n [16]byte
	_, _ = rand.Read(n[:])
	return Hello{Version: version, Caps: caps, Nonce: n}
}

func (h Hello) Encode() []byte {
	b := make([]byte, 1+2+16)
	b[0] = h.Version
	binary.BigEndian.PutUint16(b[1:3], h.Caps)
	copy(b[3:], h.Nonce[:])
	return b
}

func DecodeHello(b []byte) (Hello, error) {
	if len(b) < 19 {
		return Hello{}, errors.New("short hello")
	}
	var h Hello
	h.Version = b[0]
	h.Caps = binary.BigEndian.Uint16(b[1:3])
	copy(h.Nonce[:], b[3:19])
	return h, nil
}

type KX struct {
	Public []byte
}

func (k KX) Encode() []byte {
	l := len(k.Public)
	b := make([]byte, 2+l)
	binary.BigEndian.PutUint16(b[:2], uint16(l))
	copy(b[2:], k.Public)
	return b
}

func DecodeKX(b []byte) (KX, error) {
	if len(b) < 2 {
		return KX{}, errors.New("short kx")
	}
	l := int(binary.BigEndian.Uint16(b[:2]))
	if len(b) < 2+l {
		return KX{}, errors.New("short kx body")
	}
	pub := make([]byte, l)
	copy(pub, b[2:2+l])
	return KX{Public: pub}, nil
}

type Auth struct {
	Ed25519Pub []byte
	Signature  []byte
}

func (a Auth) Encode() []byte {
	lp := len(a.Ed25519Pub)
	ls := len(a.Signature)
	b := make([]byte, 2+lp+2+ls)
	binary.BigEndian.PutUint16(b[:2], uint16(lp))
	copy(b[2:2+lp], a.Ed25519Pub)
	off := 2 + lp
	binary.BigEndian.PutUint16(b[off:off+2], uint16(ls))
	copy(b[off+2:off+2+ls], a.Signature)
	return b
}

func DecodeAuth(b []byte) (Auth, error) {
	if len(b) < 2 {
		return Auth{}, errors.New("short auth")
	}
	lp := int(binary.BigEndian.Uint16(b[:2]))
	if len(b) < 2+lp+2 {
		return Auth{}, errors.New("short auth pub")
	}
	off := 2 + lp
	ls := int(binary.BigEndian.Uint16(b[off : off+2]))
	if len(b) < off+2+ls {
		return Auth{}, errors.New("short auth sig")
	}
	pub := make([]byte, lp)
	copy(pub, b[2:2+lp])
	sig := make([]byte, ls)
	copy(sig, b[off+2:off+2+ls])
	return Auth{Ed25519Pub: pub, Signature: sig}, nil
}

type Session struct {
	MTU uint16
}

func (s Session) Encode() []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b[:2], s.MTU)
	return b
}

func DecodeSession(b []byte) (Session, error) {
	if len(b) < 2 {
		return Session{}, errors.New("short session")
	}
	return Session{MTU: binary.BigEndian.Uint16(b[:2])}, nil
}

func Transcript(parts ...[]byte) [32]byte {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write(p)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}
