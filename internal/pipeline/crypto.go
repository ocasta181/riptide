package pipeline

import (
	"errors"

	"riptide/internal/checksum"
	"riptide/internal/cryptoutil"
)

type Encryptor struct {
	AEAD *cryptoutil.AEAD
	AAD  []byte
}

func Encrypt(enc *Encryptor) Transform {
	if enc == nil || enc.AEAD == nil {
		return func(d Descriptor) (Descriptor, error) { return Descriptor{}, errors.New("nil encryptor") }
	}
	return func(d Descriptor) (Descriptor, error) {
		ct, nonce := enc.AEAD.Seal(nil, d.Data, enc.AAD)
		out := make([]byte, 12+len(ct))
		copy(out[:12], nonce[:])
		copy(out[12:], ct)
		d.Data = out
		return d, nil
	}
}

func Decrypt(enc *Encryptor) Transform {
	if enc == nil || enc.AEAD == nil {
		return func(d Descriptor) (Descriptor, error) { return Descriptor{}, errors.New("nil encryptor") }
	}
	return func(d Descriptor) (Descriptor, error) {
		if len(d.Data) < 12 {
			return Descriptor{}, errors.New("ciphertext too short")
		}
		var nonce [12]byte
		copy(nonce[:], d.Data[:12])
		pt, err := enc.AEAD.Open(nil, d.Data[12:], enc.AAD, nonce)
		if err != nil {
			return Descriptor{}, err
		}
		d.Data = pt
		return d, nil
	}
}

func VerifyChecksum() Transform {
	return func(d Descriptor) (Descriptor, error) {
		sum := checksum.Compute128(d.Data)
		if !checksum.Equal(sum, d.Sum) {
			return Descriptor{}, errors.New("checksum mismatch")
		}
		return d, nil
	}
}
