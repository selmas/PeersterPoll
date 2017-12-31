package main

import (
	"math/big"
	"crypto/sha256"
	"strconv"
	"fmt"
)

// msg contains hash of message to get signed
type LinkableRingSignature struct {
	msg []byte
	c0  []byte
	s   []*big.Int
	tag []*big.Int
}

func generateSig(msg []byte, L [][]*big.Int, gossiper Gossiper, pos int) LinkableRingSignature {
	if L[0][pos].Cmp(gossiper.KeyPair.X) != 0 && L[1][pos].Cmp(gossiper.KeyPair.Y) != 0{
		fmt.Println("Linkable ring signature generation failed: public key not in L")
		return LinkableRingSignature{}
	}

	var tag []*big.Int
	var pubKeys []byte
	for _, keyPair := range L {
		pubKeys = append(pubKeys, keyPair[0].Bytes()...)
		pubKeys = append(pubKeys, keyPair[1].Bytes()...)
	}

	Hx, Hy := mapToPoint(pubKeys)
	p := curve.Params().P

	tag[0], tag[1] = curve.ScalarMult(Hx, Hy, gossiper.KeyPair.D.Bytes())

	u := new(big.Int)
	u.Rand(random, p)

	commonPart := pubKeys
	commonPart = append(append(commonPart, tag[0].Bytes()...), tag[1].Bytes()...)
	commonPart = append(commonPart, msg...)

	uGx, uGy := curve.ScalarBaseMult(u.Bytes())
	uHx, uHy := curve.ScalarBaseMult(u.Bytes())

	hashInput := append(append(commonPart, uGx.Bytes()...), uGy.Bytes()...)
	hashInput = append(append(hashInput, uHx.Bytes()...), uHy.Bytes()...)

	hash := sha256.New()
	_, err := hash.Write(hashInput)
	if err != nil {
		fmt.Println("hash.Write failed:", err)
	}

	// c[pos] = hash(L, Tag, msg, uG, uH)
	c := make([][]byte, len(L))
	c[pos+1] = hash.Sum(nil)

	s := make([]*big.Int, len(L))

	// c[i] = hash(L, Tag, msg, siG + siYi, siH + ciTag)
	// c[pos+2] to c[len(L)-1], c[0]
	for i := pos+1; i < len(L); i++  {
		s[i].Rand(random, p)

		siGx, siGy := curve.ScalarBaseMult(s[i].Bytes())
		siYix, siYiy := curve.ScalarMult(L[i][0], L[i][1], s[i].Bytes())
		siGciYix, siGciYiy := curve.Add(siGx, siGy, siYix, siYiy)

		siHx, siHy := curve.ScalarMult(Hx, Hy, s[i].Bytes())
		ciTagx, ciTagy := curve.ScalarMult(tag[0], tag[1], c[i])
		siHciTagx, siHciTagy := curve.Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

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
	for i := 0; i < pos ; i++ {
		s[i].Rand(random, p)

		siGx, siGy := curve.ScalarBaseMult(s[i].Bytes())
		siYix, siYiy := curve.ScalarMult(L[i][0], L[i][1], s[i].Bytes())
		siGciYix, siGciYiy := curve.Add(siGx, siGy, siYix, siYiy)

		siHx, siHy := curve.ScalarMult(Hx, Hy, s[i].Bytes())
		ciTagx, ciTagy := curve.ScalarMult(tag[0], tag[1], c[i])
		siHciTagx, siHciTagy := curve.Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

		_, err := hash.Write(hashInput)
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}

		c[i+1] = hash.Sum(nil)
	}

	// s_pos = u - privKey * c[pos]
	s[pos].Sub(u,s[pos].Mul(gossiper.KeyPair.D, new(big.Int).SetBytes(c[pos])))
	return LinkableRingSignature{msg, c[0], s, tag}
}

func verifySig(sig LinkableRingSignature, L [][]*big.Int) bool{
	// todo: find better way to convert []crypto.Pubkeys to []byte
	var pubKeys []byte
	for _, keyPair := range L {
		pubKeys = append(pubKeys, keyPair[0].Bytes()...)
		pubKeys = append(pubKeys, keyPair[1].Bytes()...)
	}

	Hx, Hy := mapToPoint(pubKeys)

	// todo: 2-dim slices ??
	c := make([][]byte, len(L)+1)
	c[0] = sig.c0
	hash := sha256.New()

	// hash(L, Tag, msg, si*G + ci*Yi, si*H + ci*Tag)
	commonPart := pubKeys
	commonPart = append(append(commonPart, sig.tag[0].Bytes()...), sig.tag[1].Bytes()...)
	commonPart = append(commonPart, sig.msg...)
	for i:=0; i<len(L); i++ {
		siGx, siGy := curve.ScalarBaseMult(sig.s[i].Bytes())
		ciYix, ciYiy := curve.ScalarMult(L[i][0], L[i][1], c[i])
		siGciYix, siGciYiy := curve.Add(siGx, siGy, ciYix, ciYiy)

		siHx, siHy := curve.ScalarMult(Hx, Hy, sig.s[i].Bytes())
		ciTagx, ciTagy := curve.ScalarMult(sig.tag[0], sig.tag[1], c[i])
		siHciTagx, siHciTagy := curve.Add(siHx, siHy, ciTagx, ciTagy)

		hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
		hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

		_, err := hash.Write(hashInput)
		if err != nil {
			fmt.Println("hash.Write failed:", err)
		}
		c[i+1] = hash.Sum(nil)
	}

	if string(c[0]) == string(c[len(L)+1]) {
		return true
	}
	return false
}

// hashes the input to a point on the curve
func mapToPoint(input []byte) (x, y *big.Int){
	i := 0
	p := curve.Params().P
	hash := sha256.New()

	for {
		_, err := hash.Write(append([]byte(strconv.Itoa(i)),input...))
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
			beta.Add(beta, curve.Params().B)

			y2 := new(big.Int).Mod(beta, p)

			// returns nil if beta not square mod p
			y = new(big.Int)
			if y.ModSqrt(y2, p) != nil {
				return x,y
			}
		}
		i++
	}
}
