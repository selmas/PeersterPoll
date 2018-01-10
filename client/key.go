package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	pkg "github.com/ValerianRousset/Peerster"
	"log"
	"math/big"
)

func key_new(s Settings, args []string) {
	origin := args[0]

	keys, err := pkg.KeyFileLoad()
	if err != nil {
		log.Fatal(err)
	}

	k, err := ecdsa.GenerateKey(pkg.Curve(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	var arr [2]big.Int
	arr[0], arr[1] = *k.PublicKey.X, *k.PublicKey.Y
	keys = append(keys, arr)

	err = pkg.PrivateKeySave(pkg.PrivateKeyFileName(origin), *k)
	if err != nil {
		log.Fatal(err)
	}

	err = pkg.KeyFileSave(keys)
	if err != nil {
		log.Fatal(err)
	}
}

func key(s Settings, args []string) {
	action := args[0]

	switch action {
	case "new":
		key_new(s, args[1:])
	default:
		panic("unkown key action: " + action)
	}
}
