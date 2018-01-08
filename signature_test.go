package pollparty

import (
	"testing"
	"crypto/elliptic"
	"math/big"
	crypto "crypto/rand" // alias needed as we import two libraries with name "rand"
	"crypto/ecdsa"
	"errors"
)

func TestMapToPointDeterministic(t  *testing.T){
	curve = elliptic.P256()

	input := []byte("Test input")

	X1, Y1 := mapToPoint(input)
	X2, Y2 := mapToPoint(input)

	if X1.Cmp(X2) != 0 || Y1.Cmp(Y2) != 0 {
		t.Errorf("Mapped same input to different points, " +
			"\n X1 = %d \n X2 = %d \n Y1 = %d \n Y2 = %d", X1,
			X2, Y1, Y2)
	}
}

func TestDifferentInputMapsToDifferentPoints(t  *testing.T){
	curve = elliptic.P256()

	input1 := []byte("Test input 1")
	input2 := []byte("Test input 2")

	X1, Y1 := mapToPoint(input1)
	X2, Y2 := mapToPoint(input2)

	if X1.Cmp(X2) == 0 && Y1.Cmp(Y2) == 0 {
		t.Errorf("Mapped different input to same point, got:\n X = %d \n tag = %d ", X1,Y1)
	}
}

func TestMapToPointReturnsPointOnCurve(t  *testing.T)  {
	curve = elliptic.P256()

	input := []byte("Test input")
	X1, Y1 := mapToPoint(input)

	if !curve.IsOnCurve(X1,Y1) {
		t.Errorf("Point not on curve, got:\n X = %d \n tag = %d", X1, Y1)
	}
}

// Output slice c in method linkableRingSignature to run this test
/*func TestValidSignature(t *testing.T)  {
	gossiper := setupGossiper()
	msg := []byte("Test input")

	numPubKey := 4
	L := initTwoDimArray(2, numPubKey)
	for i:=0; i<numPubKey-1; i++ {
		keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader) // generates key pair
		if err != nil {
			errors.New("Elliptic Curve Generation: " + err.Error())
		}
		L[i][0] = keyPair.X
		L[i][1] = keyPair.Y
	}

	pos := numPubKey-1
	L[pos][0] = gossiper.KeyPair.X
	L[pos][1] = gossiper.KeyPair.Y

	sig, c := linkableRingSignature(msg, L, *gossiper, pos)

	i:= len(sig.s)-1

	var pubKeys []byte
	for _, keyPair := range L {
		pubKeys = append(pubKeys, keyPair[0].Bytes()...)
		pubKeys = append(pubKeys, keyPair[1].Bytes()...)
	}

	Hx, Hy := mapToPoint(pubKeys)

	commonPart := pubKeys
	commonPart = append(append(commonPart, sig.tag[0].Bytes()...), sig.tag[1].Bytes()...)
	commonPart = append(commonPart, sig.msg...)

	siGx, siGy := curve.ScalarBaseMult(sig.s[i].Bytes())
	ciYix, ciYiy := curve.ScalarMult(L[i][0], L[i][1], c[i])
	siGciYix, siGciYiy := curve.Add(siGx, siGy, ciYix, ciYiy)

	siHx, siHy := curve.ScalarMult(Hx, Hy, sig.s[i].Bytes())
	ciTagx, ciTagy := curve.ScalarMult(sig.tag[0], sig.tag[1], c[i])
	siHciTagx, siHciTagy := curve.Add(siHx, siHy, ciTagx, ciTagy)

	hashInput := append(append(commonPart, siGciYix.Bytes()...), siGciYiy.Bytes()...)
	hashInput = append(append(hashInput, siHciTagx.Bytes()...), siHciTagy.Bytes()...)

	hash := sha256.New()
	_, err := hash.Write(hashInput)
	if err != nil {
		fmt.Println("hash.Write failed:", err)
	}

	nextC := hash.Sum(nil)
	fmt.Printf("c[%d] %x\n", i+1,nextC)

	if string(sig.c0) != string(nextC){
		t.Errorf("Generated Signature invalide. \nExpected: %x\nGot: %x", sig.c0,nextC)
	}
}*/

func TestVerifyGeneratedSignature(t *testing.T)  {
	gossiper := setupGossiper()

	msg := []byte("Test input")

	numPubKey := 4
	L := initTwoDimArray(2, numPubKey)
	for pos := 0; pos < numPubKey; pos++  {
		for i:=0; i<numPubKey; i++ {
			if i == pos {
				L[pos][0] = gossiper.KeyPair.X
				L[pos][1] = gossiper.KeyPair.Y
			} else {
				keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader) // generates key pair
				if err != nil {
					errors.New("Elliptic Curve Generation: " + err.Error())
				}
				L[i][0] = keyPair.X
				L[i][1] = keyPair.Y
			}
		}

		lrs := linkableRingSignature(msg, L, *gossiper, pos)

		if !verifySig(lrs, L) {
			t.Errorf("Unable to verify the generated signature, public key at position %d",pos)
		}
	}
}

func TestVerifyInvalidSignature(t *testing.T)  {
	gossiper := setupGossiper()

	msg := []byte("Test input")

	pos := 3
	numPubKey := 4
	L := initTwoDimArray(2, numPubKey)
	for i:=0; i<numPubKey; i++ {
		if i == pos {
			L[pos][0] = gossiper.KeyPair.X
			L[pos][1] = gossiper.KeyPair.Y
		} else {
			keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader) // generates key pair
			if err != nil {
				errors.New("Elliptic Curve Generation: " + err.Error())
			}
			L[i][0] = keyPair.X
			L[i][1] = keyPair.Y
		}
	}

	lrs := linkableRingSignature(msg, L, *gossiper, pos)
	lrs.s[0] = lrs.s[1] // messing with some values

	if verifySig(lrs, L) {
		t.Errorf("Verified invalid signautre")
	}
}
// inspired by https://stackoverflow.com/questions/7703251/slice-of-slices-types
func initTwoDimArray(dx, dy int) [][]*big.Int {
	array := make([][]*big.Int, dy)
	for i := range array {
		array[i] = make([]*big.Int, dx)
	}
	return array
}

func setupGossiper() *Gossiper{
	gossiper, _ := NewGossiper("NodeA", NewServer("127.0.0.1:5000"))
	defer gossiper.Server.Conn.Close()

	return gossiper
}