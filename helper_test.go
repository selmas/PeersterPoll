package pollparty

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crypto "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestValidECSignature(t *testing.T) {
	g := DummyGossiper()
	pkg := PollPacket{
		ID:   PollKey{g.KeyPair.PublicKey, uint64(0)},
		Poll: DummyPoll(),
	}

	sig, err := ecSignature(&g, pkg)
	if err != nil {
		return
	}

	msg := GossipPacket{
		Poll:      &pkg,
		Signature: &sig,
		Status:    nil,
	}

	if !signatureValid(&g, msg) {
		t.Errorf("Cannot verify generated signature, \ns: %d\nr: %d", sig.ellipticCurveSig.s,
			sig.ellipticCurveSig.r)
	}
}

func TestValidLinkableRingSignature(t *testing.T) {
	g := DummyGossiper()
	poll := PollPacket{
		ID:         PollKey{g.KeyPair.PublicKey, uint64(0)},
		Commitment: &Commitment{},
	}

	input, err := json.Marshal(poll)
	if err != nil {
		log.Printf("unable to encode as json")
	}

	pos := 3
	numPubKey := 4
	L := DummyPublicKeyArray(g, pos, numPubKey)

	sig := linkableRingSignature(input, L, &g.KeyPair, pos)
	if err != nil {
		return
	}

	msg := GossipPacket{
		Poll:      &poll,
		Signature: &Signature{&sig, nil},
		Status:    nil,
	}

	if !signatureValid(&g, msg) {
		t.Errorf("Cannot verify generated linkable ring signature")
	}
}

func DummyGossiper() Gossiper {
	curve = elliptic.P256()
	key, err := ecdsa.GenerateKey(curve, crypto.Reader)
	if err != nil {
		fmt.Printf("error generating key pair")
	}

	return Gossiper{sync.RWMutex{},
		"name",
		uint64(0),
		*key,
		PeerSet{},
		RunningPollSet{},
		PollSet{},
		Server{},
		[]ecdsa.PublicKey{},
		RepOpinions{},
		Blacklist{},
	}
}

func DummyPoll() *Poll {
	return &Poll{
		Question:  "Do you like dogs?",
		Options:   []string{"Yes", "No"},
		StartTime: time.Now(),
		Duration:  time.Hour,
	}
}
