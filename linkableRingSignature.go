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
