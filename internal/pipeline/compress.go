package pipeline

import (
	"bytes"
	"io"

	"github.com/pierrec/lz4/v4"
)

func CompressLZ4() Transform {
	return func(d Descriptor) (Descriptor, error) {
		var buf bytes.Buffer
		w := lz4.NewWriter(&buf)
		if _, err := w.Write(d.Data); err != nil {
			return Descriptor{}, err
		}
		if err := w.Close(); err != nil {
			return Descriptor{}, err
		}
		d.Data = buf.Bytes()
		return d, nil
	}
}

func DecompressLZ4() Transform {
	return func(d Descriptor) (Descriptor, error) {
		r := lz4.NewReader(bytes.NewReader(d.Data))
		out, err := io.ReadAll(r)
		if err != nil {
			return Descriptor{}, err
		}
		d.Data = out
		return d, nil
	}
}
