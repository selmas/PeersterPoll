package pollparty

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"os"
)

const KeyFileName = "keys"

func KeyFileSave(keys []ecdsa.PublicKey) error {
	file, err := os.Create(KeyFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, k := range keys {
		bytes := elliptic.Marshal(k.Curve, k.X, k.Y)

		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func KeyFileLoad() ([]ecdsa.PublicKey, error) {
	ret := make([]ecdsa.PublicKey, 0)

	file, err := os.OpenFile(KeyFileName, os.O_CREATE, 0600)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	blockLen := 1 + 2*((Curve().Params().BitSize+7)>>3) // taken from stdlib == header + 2 * sizeof(big.Int)
	block := make([]byte, 0, blockLen)

	for {
		count, err := file.Read(block)
		if err != nil {
			return ret, err
		}

		if count == 0 {
			break
		}

		x, y := elliptic.Unmarshal(Curve(), block)
		if x == nil || y == nil {
			return ret, errors.New("unable to unmarshal point")
		}

		ret = append(ret, ecdsa.PublicKey{
			Curve: Curve(),
			X:     x,
			Y:     y,
		})
	}

	return ret, nil
}
