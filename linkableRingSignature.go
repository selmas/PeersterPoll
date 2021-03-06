package pollparty

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
)

// msg contains hash of message to get signed
type LinkableRingSignature struct {
	Message []byte
	C0      []byte
	S       []*big.Int
	Tag     [2]*big.Int
}

func linkableRingSignature(msg []byte, L [][2]big.Int, tmpKey *ecdsa.PrivateKey, pos int) LinkableRingSignature {
	if pos > len(L) || L[pos][0].Cmp(tmpKey.X) != 0 && L[pos][1].Cmp(tmpKey.Y) != 0 {
		fmt.Println("Linkable ring signature generation failed: public key not in L")
		return LinkableRingSignature{}
	}

	var tag [2]*big.Int
	var pubKeys []byte
	for _, keyPair := range L {
		pubKeys = append(pubKeys, keyPair[0].Bytes()...)
		pubKeys = append(pubKeys, keyPair[1].Bytes()...)
	}

	Hx, Hy := mapToPoint(pubKeys)
	n := Curve().Params().N

	tag[0], tag[1] = Curve().ScalarMult(Hx, Hy, tmpKey.D.Bytes())

	u, err := rand.Int(rand.Reader, n)
	if err != nil {
		fmt.Println("rand.Int failed:", err)
	}

	commonPart := pubKeys
	commonPart = append(append(commonPart, tag[0].Bytes()...), tag[1].Bytes()...)
	commonPart = append(commonPart, msg...)

	uGx, uGy := Curve().ScalarBaseMult(u.Bytes())
	uHx, uHy := Curve().ScalarMult(Hx, Hy, u.Bytes())

	hashInput := append(append(commonPart, uGx.Bytes()...), uGy.Bytes()...)
	hashInput = append(append(hashInput, uHx.Bytes()...), uHy.Bytes()...)

	hash := sha256.New()
	_, err = hash.Write(hashInput)
	if err != nil {
		fmt.Println("hash.Write failed:", err)
	}

	// c[pos+1] = hash(L, Tag, msg, uG, uH)
	c := make([][]byte, len(L))
	if pos == len(L)-1 {
		c[0] = hash.Sum(nil)
	} else {
		c[pos+1] = hash.Sum(nil)
	}

	s := make([]*big.Int, len(L))

	// c[i+1] = hash(L, Tag, msg, s[i]*G + s[i]*Yi, s[i]*H + c[i]*Tag)
	// c[pos+2] to c[len(L)-1], c[0]
	for i := pos + 1; i < len(L); i++ {
		s[i], err = rand.Int(rand.Reader, n)
		if err != nil {
			fmt.Println("rand.Int failed:", err)
		}

		siGx, siGy := Curve().ScalarBaseMult(s[i].Bytes())
		ciYix, ciYiy := Curve().ScalarMult(&L[i][0], &L[i][1], c[i])
		siGciYix, siGciYiy := Curve().Add(siGx, siGy, ciYix, ciYiy)

		siHx, siHy := Curve().ScalarMult(Hx, Hy, s[i].Bytes())
		ciTagx, ciTagy := Curve().ScalarMult(tag[0], tag[1], c[i])
		siHciTagx, siHciTagy := Curve().Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

		hash := sha256.New()
		_, err := hash.Write(hashInput)
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}

		if i == len(L)-1 {
			c[0] = hash.Sum(nil)
		} else {
			c[i+1] = hash.Sum(nil)
		}
	}

	// c[i] = hash(L, Tag, msg, siG + siYi, siH + ciTag)
	// c[1] to c[pos]
	for i := 0; i < pos; i++ {
		s[i], err = rand.Int(rand.Reader, n)
		if err != nil {
			fmt.Println("rand.Int failed:", err)
		}

		siGx, siGy := Curve().ScalarBaseMult(s[i].Bytes())
		ciYix, ciYiy := Curve().ScalarMult(&L[i][0], &L[i][1], c[i])
		siGciYix, siGciYiy := Curve().Add(siGx, siGy, ciYix, ciYiy)

		siHx, siHy := Curve().ScalarMult(Hx, Hy, s[i].Bytes())
		ciTagx, ciTagy := Curve().ScalarMult(tag[0], tag[1], c[i])
		siHciTagx, siHciTagy := Curve().Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

		hash := sha256.New()
		_, err := hash.Write(hashInput)
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}

		c[i+1] = hash.Sum(nil)
	}

	// s_pos = u - privKey * c[pos] mod n
	cPos := new(big.Int).SetBytes(c[pos])
	privKeyCpos := new(big.Int).Mul(tmpKey.D, cPos)
	privKeyCpos = new(big.Int).Mod(privKeyCpos, n)

	s[pos] = new(big.Int).Sub(u, privKeyCpos)
	s[pos] = new(big.Int).Add(s[pos], n)
	s[pos] = new(big.Int).Mod(s[pos], n)

	return LinkableRingSignature{msg, c[0], s, tag}
}

func verifySig(sig LinkableRingSignature, L [][2]big.Int) bool {
	var pubKeys []byte
	for _, keyPair := range L {
		pubKeys = append(pubKeys, keyPair[0].Bytes()...)
		pubKeys = append(pubKeys, keyPair[1].Bytes()...)
	}

	Hx, Hy := mapToPoint(pubKeys)

	c := make([][]byte, len(L)+1)
	c[0] = sig.C0

	// hash(L, Tag, msg, si*G + ci*Yi, si*H + ci*Tag)
	commonPart := pubKeys
	commonPart = append(append(commonPart, sig.Tag[0].Bytes()...), sig.Tag[1].Bytes()...)
	commonPart = append(commonPart, sig.Message...)

	for i := 0; i < len(L); i++ {
		siGx, siGy := Curve().ScalarBaseMult(sig.S[i].Bytes())
		ciYix, ciYiy := Curve().ScalarMult(&L[i][0], &L[i][1], c[i])
		siGciYix, siGciYiy := Curve().Add(siGx, siGy, ciYix, ciYiy)

		siHx, siHy := Curve().ScalarMult(Hx, Hy, sig.S[i].Bytes())
		ciTagx, ciTagy := Curve().ScalarMult(sig.Tag[0], sig.Tag[1], c[i])
		siHciTagx, siHciTagy := Curve().Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

		hash := sha256.New()
		_, err := hash.Write(hashInput)
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}
		c[i+1] = hash.Sum(nil)
	}

	if string(c[0]) == string(c[len(L)]) {
		return true
	}
	return false
}

// hashes the input to a point on the curve
func mapToPoint(input []byte) (x, y *big.Int) {
	i := 0
	p := Curve().Params().P

	for {
		hash := sha256.New()
		_, err := hash.Write(append([]byte(strconv.Itoa(i)), input...))
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}

		hashBytes := hash.Sum(nil)
		x = new(big.Int).SetBytes(hashBytes)

		if x.Cmp(p) == -1 {
			// y² = x³ - 3x + b
			x3 := new(big.Int).Mul(x, x)
			x3.Mul(x3, x)

			threeX := new(big.Int).Lsh(x, 1)
			threeX.Add(threeX, x)

			beta := new(big.Int).Sub(x3, threeX)
			beta.Add(beta, Curve().Params().B)

			y2 := new(big.Int).Mod(beta, p)

			// returns nil if beta not square mod p
			y = new(big.Int)
			if y.ModSqrt(y2, p) != nil {
				return x, y
			}
		}
		i++
	}
}
