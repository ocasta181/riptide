package checksum

import (
	"encoding/binary"

	"github.com/zeebo/blake3"
)

type Sum128 [16]byte

func Compute128(data []byte) Sum128 {
	h := blake3.New()
	_, _ = h.Write(data)
	var out Sum128
	copy(out[:], h.Sum(nil)[:16])
	return out
}

func Equal(a, b Sum128) bool {
	var v uint8
	for i := 0; i < 16; i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func ToUint64Pair(s Sum128) (uint64, uint64) {
	hi := binary.BigEndian.Uint64(s[:8])
	lo := binary.BigEndian.Uint64(s[8:])
	return hi, lo
}

func FromUint64Pair(hi, lo uint64) Sum128 {
	var s Sum128
	binary.BigEndian.PutUint64(s[:8], hi)
	binary.BigEndian.PutUint64(s[8:], lo)
	return s
}
