package pipeline

import (
	"riptide/internal/checksum"
)

func ComputeChecksum() Transform {
	return func(d Descriptor) (Descriptor, error) {
		d.Sum = checksum.Compute128(d.Data)
		return d, nil
	}
}

func Chunk(data []byte, chunkSize int) []Descriptor {
	if chunkSize <= 0 {
		chunkSize = 1
	}
	var out []Descriptor
	var off uint64
	for i := 0; i < len(data); {
		j := i + chunkSize
		if j > len(data) {
			j = len(data)
		}
		buf := make([]byte, j-i)
		copy(buf, data[i:j])
		out = append(out, Descriptor{
			Offset: off,
			Data:   buf,
		})
		off += uint64(len(buf))
		i = j
	}
	return out
}

func ApplyTransforms(ds []Descriptor, ts ...Transform) ([]Descriptor, error) {
	out := make([]Descriptor, len(ds))
	for i := range ds {
		cur := ds[i]
		var err error
		for _, t := range ts {
			cur, err = t(cur)
			if err != nil {
				return nil, err
			}
		}
		out[i] = cur
	}
	return out, nil
}
