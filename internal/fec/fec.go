package fec

import (
	"errors"

	"github.com/klauspost/reedsolomon"
)

type Codec struct {
	dataShards   int
	parityShards int
	enc          reedsolomon.Encoder
}

func NewCodec(dataShards, parityShards int) (*Codec, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, errors.New("invalid shard counts")
	}
	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, err
	}
	return &Codec{
		dataShards:   dataShards,
		parityShards: parityShards,
		enc:          enc,
	}, nil
}

func (c *Codec) DataShards() int   { return c.dataShards }
func (c *Codec) ParityShards() int { return c.parityShards }

func (c *Codec) BuildShards(data [][]byte) ([][]byte, error) {
	if len(data) != c.dataShards {
		return nil, errors.New("wrong number of data shards")
	}
	if len(data) == 0 {
		return nil, errors.New("no shards")
	}
	size := len(data[0])
	for i := 1; i < len(data); i++ {
		if len(data[i]) != size {
			return nil, errors.New("unequal shard sizes")
		}
	}
	shards := make([][]byte, c.dataShards+c.parityShards)
	for i := 0; i < c.dataShards; i++ {
		cp := make([]byte, size)
		copy(cp, data[i])
		shards[i] = cp
	}
	for i := c.dataShards; i < len(shards); i++ {
		shards[i] = make([]byte, size)
	}
	if err := c.enc.Encode(shards); err != nil {
		return nil, err
	}
	return shards, nil
}

func (c *Codec) Reconstruct(shards [][]byte) error {
	if len(shards) != c.dataShards+c.parityShards {
		return errors.New("wrong total shard count")
	}
	if err := c.enc.Reconstruct(shards); err != nil {
		return err
	}
	ok, err := c.enc.Verify(shards)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("verification failed")
	}
	return nil
}

func SelectParity(lossRate float64, maxParity int) int {
	if maxParity <= 0 {
		return 0
	}
	switch {
	case lossRate <= 0.005:
		if maxParity >= 1 {
			return 1
		}
	case lossRate <= 0.02:
		if maxParity >= 2 {
			return 2
		}
	case lossRate <= 0.05:
		if maxParity >= 3 {
			return 3
		}
	case lossRate <= 0.10:
		if maxParity >= 4 {
			return 4
		}
	default:
		// High loss: cap to maxParity
		return maxParity
	}
	return maxParity
}
