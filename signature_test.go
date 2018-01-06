package main

import (
	"testing"
	"crypto/elliptic"
	"math/big"
	crypto "crypto/rand" // alias needed as we import two libraries with name "rand"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"crypto/sha256"
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

func TestSignatureGenerationDeterministic(t *testing.T){
	gossiper := setupGossiper()

	msg1 := []byte("Test input")
	msg2 := []byte("Test input")

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

	lrs1,_ := generateSig(msg1, L, *gossiper, pos)
	lrs2,_ := generateSig(msg2, L, *gossiper, pos)

	if string(lrs1.msg) != string(lrs2.msg) {
		t.Errorf("Messages not the same, message 1: %x, message 2: %x", lrs1.msg, lrs2.msg)
	} else if lrs1.tag[0].Cmp(lrs2.tag[0]) != 0 {
		t.Errorf("Tag X not the same, tag X1: %x, tag X2: %x", lrs1.tag[0], lrs2.tag[0])
	} else if lrs1.tag[1].Cmp(lrs2.tag[1]) != 0 {
		t.Errorf("Tag Y not the same, tag Y1: %x, tag Y2: %x", lrs1.tag[1], lrs2.tag[1])
	} else if string(lrs1.c0) != string(lrs2.c0){
		t.Errorf("C0 not the same, sig 1: %x, sig 2: %x", lrs1.c0, lrs2.c0)
	}

	for i:=0; i<len(lrs1.s);i++ {
		if lrs1.s[i].Cmp(lrs2.s[i]) != 0 {
			t.Errorf("s[%d] not the same, 1: %x, 2: %x", i, lrs1.s[i], lrs2.s[i])
		}
	}
}

func TestVerification(t *testing.T){
	setupGossiper()
	msg := []byte("")

	l1,_:=	new(big.Int).SetString("59782603363841408971056574526916208460707204720860703136421326544832538993649",10)
	l2,_:=	new(big.Int).SetString("71735795577836931504184728305099182139551642165338213542084614987201074903889",10)
	l3,_:=	new(big.Int).SetString("69335282256630887760677903716166487313183851179682156385266886019672416508733",10)
	l4,_:=	new(big.Int).SetString("15239595651785459382630357439953650859710601997942736997227958176259104951612",10)
	l5,_:=	new(big.Int).SetString("77274833420574580130132478443395727507413619626351012037733301842841317505833",10)
	l6,_:=	new(big.Int).SetString("51551585223004698217747352386577616234480509463306739272912485189671537534582",10)
	l7,_:=	new(big.Int).SetString("47895756667691830293283638688652806083753094959913434141099766406816837059492",10)
	l8,_:=	new(big.Int).SetString("49194694673295522391960597673276940265847757740059640982277684447121451502034",10)
	l9,_:=	new(big.Int).SetString("20160897872883339945649225656315246110358972075251985214437320275173408194572",10)
	l10,_:=	new(big.Int).SetString("59472953237846610054296281409003439548077091793313097348592502812800964772594",10)

	L := [][]*big.Int{
		{l1,l2,},
		{l3,l4,},
		{l5,l6,},
		{l7,l8,},
		{l9,l10,},
	}

	tag1,_:= new(big.Int).SetString("112470650631809901655156091493757479663798253097706365456643188004923782762561",10)
	tag2,_:= new(big.Int).SetString("85360625473191250059332399775805842194940748210306053685131669295957438558485",10)
	Tag := [2]*big.Int{tag1,tag2}

	c0,_ := new(big.Int).SetString("55216764086040345573354010399688228621254794779552051699702726589281858347183",10)

	s1,_:=	new(big.Int).SetString("105210217094425297789474396333731366807853810297724663345436223403976364857271",10)
	s2,_:=	new(big.Int).SetString("61241892716971742545702846389953971092292204825449600560551420782541483939924",10)
	s3,_:=	new(big.Int).SetString("643953056499445665129656823597160132992283685122302849315473277369323302465", 10)
	s4,_:=	new(big.Int).SetString("48617406641589855092720916430837353538204521464820374827581373047564655375368",10)
	s5,_:=	new(big.Int).SetString("26199623993143287956981664076527008920993357796979643621803367839123795711902",10)
	s := []*big.Int{s1,s2,s3,s4,s5}

	lrs := LinkableRingSignature{msg, c0.Bytes(), s, Tag}

	if !verifySig(lrs, L) {
		t.Errorf("Verification failed")
	}
}

func TestValidSignature(t *testing.T)  {
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

	sig, c := generateSig(msg, L, *gossiper, pos)

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
}

func TestVerifyGeneratedSignature(t *testing.T)  {
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

	lrs,_ := generateSig(msg, L, *gossiper, pos)

	if !verifySig(lrs, L) {
		t.Errorf("Unable to verify the generated signature")
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
	gossiper := NewGossiper("NodeA", NewServer("127.0.0.1:5000"))
	defer gossiper.Server.Conn.Close()

	return gossiper
}