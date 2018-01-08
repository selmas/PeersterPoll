package pollparty

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"io/ioutil"
	"math/big"
)

func PrivateKeyFileName(origin string) string {
	return origin + ".key"
}

func PrivateKeySave(filename string, k ecdsa.PrivateKey) error {
	public := elliptic.Marshal(Curve(), k.X, k.Y)
	private := append(public, k.D.Bytes()...)

	ioutil.WriteFile(filename, private, 0400)

	return nil
}

func PrivateKeyLoad(filename string) (ecdsa.PrivateKey, error) {
	var ret ecdsa.PrivateKey

	publicLen := len(elliptic.Marshal(Curve(), new(big.Int), new(big.Int)))

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return ret, err
	}

	x, y := elliptic.Unmarshal(Curve(), bytes[:publicLen])
	d := new(big.Int).SetBytes(bytes[publicLen:])

	ret.PublicKey = ecdsa.PublicKey{
		Curve: Curve(),
		X:     x,
		Y:     y,
	}
	ret.D = d

	return ret, nil
}
