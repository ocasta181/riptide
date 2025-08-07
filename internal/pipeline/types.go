package pipeline

import (
	"errors"

	"riptide/internal/checksum"
)

type Descriptor struct {
	ChunkID uint64
	Offset  uint64
	Data    []byte
	Sum     checksum.Sum128
}

type Transform func(Descriptor) (Descriptor, error)

func Compose(ts ...Transform) Transform {
	return func(d Descriptor) (Descriptor, error) {
		cur := d
		for _, t := range ts {
			if t == nil {
				return Descriptor{}, errors.New("nil transform")
			}
			var err error
			cur, err = t(cur)
			if err != nil {
				return Descriptor{}, err
			}
		}
		return cur, nil
	}
}

type Queue[T any] interface {
	Enqueue(T) bool
	Dequeue() (T, bool)
	Len() int
	Cap() int
}
