package delta

import (
	"errors"

	"github.com/zeebo/blake3"
)

const modAdler = 65521

type Rolling struct {
	n      int
	a      uint32
	b      uint32
	win    []byte
	cursor int
	init   bool
}

func NewRolling(window int) *Rolling {
	if window <= 0 {
		window = 1
	}
	return &Rolling{
		n:   window,
		win: make([]byte, window),
	}
}

func (r *Rolling) Init(data []byte) error {
	if len(data) != r.n {
		return errors.New("init size mismatch")
	}
	var a, b uint32
	for i := 0; i < r.n; i++ {
		a += uint32(data[i])
		b += a
	}
	a %= modAdler
	b %= modAdler
	copy(r.win, data)
	r.a = a
	r.b = b
	r.cursor = 0
	r.init = true
	return nil
}

func (r *Rolling) Roll(next byte) (uint32, error) {
	if !r.init {
		return 0, errors.New("not initialized")
	}
	old := r.win[r.cursor]
	r.win[r.cursor] = next
	r.cursor++
	if r.cursor == r.n {
		r.cursor = 0
	}
	a := r.a + uint32(next) + modAdler - uint32(old)
	a %= modAdler
	b := r.b + a + modAdler - (uint32(r.n) * uint32(old) % modAdler)
	b %= modAdler
	r.a = a
	r.b = b
	return r.Sum(), nil
}

func (r *Rolling) Sum() uint32 {
	return (r.b << 16) | r.a
}

func Strong256(b []byte) [32]byte {
	return blake3.Sum256(b)
}
