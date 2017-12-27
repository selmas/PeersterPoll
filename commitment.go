package main

import (
	"crypto/elliptic"
	"math/big"
	"math/rand"
	"time"
)

var curve elliptic.Curve

//P       *big.Int // the order of the underlying field
//N       *big.Int // the order of the base point
//B       *big.Int // the constant of the curve equation
//Gx, Gy  *big.Int // (x,y) of the base point
//BitSize int      // the size of the underlying field
//Name    string   // the canonical name of the curve
var random *rand.Rand

func setup() error {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))

	// TODO store keypair for gossiper

	return nil
}

func commit(x int) {
	var r big.Int
	max := big.NewInt(0xDEADBEEF)
	r.Rand(random, max)
}

type CommitPedersen struct {
}

type OpenPedersen struct {
}
