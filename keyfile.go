package pollparty

import (
	"crypto/elliptic"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
)

const KeyFileName = "keys"

func KeyFileSave(keys [][2]big.Int) error {
	file, err := os.Create(KeyFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, k := range keys {
		bytes := elliptic.Marshal(Curve(), &k[0], &k[1])

		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func KeyFileLoad() ([][2]big.Int, error) {
	ret := make([][2]big.Int, 0)

	file, err := os.OpenFile(KeyFileName, os.O_CREATE, 0600)
	if err != nil {
		return ret, err
	}
	file.Close()

	content, err := ioutil.ReadFile(KeyFileName)
	if err != nil {
		return ret, err
	}

	blockLen := 1 + 2*((Curve().Params().BitSize+7)>>3) // taken from stdlib == header + 2 * sizeof(big.Int)

	for i := 0; i < len(content)/blockLen; i++ {
		block := content[i*blockLen : (i+1)*blockLen]

		x, y := elliptic.Unmarshal(Curve(), block)
		if x == nil || y == nil {
			return ret, errors.New("unable to unmarshal point")
		}

		ret = append(ret, [2]big.Int{*x, *y})
	}

	return ret, nil
}
