package main

import (
	"math/big"
	"crypto/sha256"
	"bytes"
	"encoding/binary"
	"fmt"
)

// msg contains hash of message to get signed
type LinkableRingSignature struct {
	msg []byte
	c1 []byte
	S []big.Int
}

/*func generateSig(msg []byte, L []crypto.PublicKey, gossiper Gossiper) LinkableRingSignature {

	return GossipPacket{&msg, nil}
}*/

/*func verifySig(sig LinkableRingSignature, L []crypto.PublicKey) bool{
	// todo: test this !!
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, L)

	Hx, Hy := mapToPoint()
}
*/

// hashes the input to a point on the curve
func mapToPoint(input []byte) (x, y *big.Int){
	i := 0
	p := curve.Params().P
	hash := sha256.New()

	var exp big.Int
	exp.Div(exp.Sub(p, big.NewInt(1)), big.NewInt(2)) // (p-1)/2

	for {
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, i)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}

		hash.Write(append(buf.Bytes(),input...))
		hashBytes := hash.Sum(nil)

		if x.SetBytes(hashBytes).Cmp(p) == -1 {
			var yPow2, xPow3, xTripled *big.Int
			// y^2 = x^3 + ax + b mod p
			// for NIST Prime Curves (incl P-256) a == p - 3
			// y^2 = x^3 - 3x + b mod p
			xPow3.Exp(x,big.NewInt(3),p) // x^3
			xTripled.Mul(x, big.NewInt(3))
			yPow2.Mod(yPow2.Add(yPow2.Sub(xPow3, xTripled),curve.Params().B),p)

			sqrtExist := y.ModSqrt(x, p) // returns nil if x not square mod p

			if sqrtExist != nil {
				return x,y
			}
		}
		i++
	}
}
