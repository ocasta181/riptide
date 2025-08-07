package pipeline

import (
	"errors"

	"riptide/internal/fec"
)

func FECGroupEncode(ds []Descriptor, dataShards, parityShards int) ([]Descriptor, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, errors.New("invalid shard counts")
	}
	if len(ds) != dataShards {
		return nil, errors.New("descriptor count must equal dataShards")
	}
	maxLen := 0
	for i := range ds {
		if l := len(ds[i].Data); l > maxLen {
			maxLen = l
		}
	}
	data := make([][]byte, dataShards)
	for i := 0; i < dataShards; i++ {
		if len(ds[i].Data) == maxLen {
			cp := make([]byte, maxLen)
			copy(cp, ds[i].Data)
			data[i] = cp
			continue
		}
		p := make([]byte, maxLen)
		copy(p, ds[i].Data)
		data[i] = p
	}
	codec, err := fec.NewCodec(dataShards, parityShards)
	if err != nil {
		return nil, err
	}
	shards, err := codec.BuildShards(data)
	if err != nil {
		return nil, err
	}
	out := make([]Descriptor, 0, len(ds)+parityShards)
	out = append(out, ds...)
	for i := dataShards; i < len(shards); i++ {
		out = append(out, Descriptor{
			Offset: ds[0].Offset,
			Data:   shards[i],
		})
	}
	return out, nil
}

func FECGroupReconstruct(shards []Descriptor, dataShards, parityShards int, lostIdxs []int) ([]Descriptor, error) {
	if len(shards) != dataShards+parityShards {
		return nil, errors.New("wrong shard count")
	}
	arr := make([][]byte, len(shards))
	for i := range shards {
		arr[i] = shards[i].Data
	}
	for _, idx := range lostIdxs {
		if idx < 0 || idx >= len(arr) {
			return nil, errors.New("lost index out of range")
		}
		arr[idx] = nil
	}
	codec, err := fec.NewCodec(dataShards, parityShards)
	if err != nil {
		return nil, err
	}
	if err := codec.Reconstruct(arr); err != nil {
		return nil, err
	}
	out := make([]Descriptor, len(shards))
	for i := range shards {
		out[i] = shards[i]
		out[i].Data = arr[i]
	}
	return out, nil
}
