package delta

import (
	"bytes"
	"testing"

	"github.com/zeebo/blake3"
)

func adlerLike(b []byte) uint32 {
	const modAdler = 65521
	var a, c uint32
	for i := 0; i < len(b); i++ {
		a += uint32(b[i])
		c += a
	}
	a %= modAdler
	c %= modAdler
	return (c << 16) | a
}

func TestRolling_InitAndRoll(t *testing.T) {
	data := []byte("abcdefg")
	r := NewRolling(3)
	if err := r.Init(data[:3]); err != nil {
		t.Fatalf("init err: %v", err)
	}
	if got, want := r.Sum(), adlerLike(data[:3]); got != want {
		t.Fatalf("sum0 got %d want %d", got, want)
	}
	sum1, err := r.Roll(data[3])
	if err != nil {
		t.Fatalf("roll1 err: %v", err)
	}
	if want := adlerLike([]byte("bcd")); sum1 != want {
		t.Fatalf("sum1 got %d want %d", sum1, want)
	}
	sum2, err := r.Roll(data[4])
	if err != nil {
		t.Fatalf("roll2 err: %v", err)
	}
	if want := adlerLike([]byte("cde")); sum2 != want {
		t.Fatalf("sum2 got %d want %d", sum2, want)
	}
	if _, err := NewRolling(3).Roll('x'); err == nil {
		t.Fatalf("expected error when rolling before init")
	}
}

func TestStrong256(t *testing.T) {
	b := []byte("hello world")
	got := Strong256(b)
	want := blake3.Sum256(b)
	if !bytes.Equal(got[:], want[:]) {
		t.Fatalf("strong256 mismatch")
	}
}
