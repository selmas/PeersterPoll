package main

import (
	crypto "crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/rand"
	"time"
	big "math/big"
)

var curve elliptic.Curve
//P       *big.Int // the order of the underlying field
//N       *big.Int // the order of the base point
//B       *big.Int // the constant of the curve equation
//Gx, Gy  *big.Int // (x,y) of the base point
//BitSize int      // the size of the underlying field
//Name    string   // the canonical name of the curve

func setup()  {
	curve = elliptic.P256()
	// Reader is a global, shared instance of a cryptographically strong pseudo-random generator.
	keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader)  // generates key pair


	if err != nil {
		errors.New("Elliptic Curve Generation: " + err.Error())
	}

	// TODO store keypair for gossiper
}

func commit(x int) {
	s1 := rand.NewSource(time.Now().UnixNano())
	var r big.Int
	// TODO very unsure how to pick a bigInt at random
	r.Rand(&rand.Rand{s1, nil, nil, nil}, curve.Params().N)
}


type CommitPedersen struct {

}

type OpenPedersen struct {

}





