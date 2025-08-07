package delta

import (
	"encoding/hex"
)

type Weak uint32

type BlockSig struct {
	Weak   Weak
	Strong [32]byte
	Offset int
	Len    int
}

type FileSig struct {
	BlockSize int
	byKey     map[string]BlockSig
}

func keyFor(w Weak, s [32]byte) string {
	b := make([]byte, 0, 8+64)
	// weak as 8-hex chars + ':' + strong hex
	b = append(b, []byte(hexWeak(uint32(w)))...)
	b = append(b, ':')
	dst := make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(dst, s[:])
	return string(append(b, dst...))
}

func hexWeak(u uint32) string {
	const hexdigits = "0123456789abcdef"
	var b [8]byte
	for i := 7; i >= 0; i-- {
		b[i] = hexdigits[u&0xf]
		u >>= 4
	}
	return string(b[:])
}

func ComputeFileSig(data []byte, blockSize int) FileSig {
	if blockSize <= 0 {
		blockSize = 1
	}
	m := make(map[string]BlockSig)
	for off := 0; off < len(data); off += blockSize {
		end := off + blockSize
		if end > len(data) {
			end = len(data)
		}
		w := Weak(adlerLike(data[off:end]))
		s := Strong256(data[off:end])
		m[keyFor(w, s)] = BlockSig{
			Weak:   w,
			Strong: s,
			Offset: off,
			Len:    end - off,
		}
	}
	return FileSig{BlockSize: blockSize, byKey: m}
}

type Op uint8

const (
	OpCopy Op = iota + 1
	OpLiteral
)

type DeltaInstruction struct {
	Op     Op
	SrcOff int
	Len    int
	Data   []byte
}

func ComputeDelta(sig FileSig, newData []byte) []DeltaInstruction {
	if sig.BlockSize <= 0 {
		sig.BlockSize = 1
	}
	var out []DeltaInstruction
	for off := 0; off < len(newData); off += sig.BlockSize {
		end := off + sig.BlockSize
		if end > len(newData) {
			end = len(newData)
		}
		chunk := newData[off:end]
		w := Weak(adlerLike(chunk))
		s := Strong256(chunk)
		if bs, ok := sig.byKey[keyFor(w, s)]; ok {
			out = append(out, DeltaInstruction{
				Op:     OpCopy,
				SrcOff: bs.Offset,
				Len:    bs.Len,
			})
			continue
		}
		cp := make([]byte, len(chunk))
		copy(cp, chunk)
		out = append(out, DeltaInstruction{
			Op:   OpLiteral,
			Len:  len(cp),
			Data: cp,
		})
	}
	return out
}

func ApplyDelta(basis []byte, delta []DeltaInstruction) []byte {
	var out []byte
	for _, ins := range delta {
		switch ins.Op {
		case OpCopy:
			start := ins.SrcOff
			end := start + ins.Len
			if start < 0 {
				start = 0
			}
			if end > len(basis) {
				end = len(basis)
			}
			if start > end {
				start = end
			}
			out = append(out, basis[start:end]...)
		case OpLiteral:
			out = append(out, ins.Data...)
		}
	}
	return out
}

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
