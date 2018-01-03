package main

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

func TestVerifiesSuccessfulGeneratedSignature(t *testing.T){
	gossiper := NewGossiper("NodeA", NewServer("127.0.0.1:5000"))
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

	lrs := generateSig(msg, L, *gossiper, pos)

	if !verifySig(lrs, L) {
		t.Errorf("Generated Signature cannot be verified")
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