package pollparty

import (
	"testing"
	"crypto/ecdsa"
	"math/big"
	"fmt"
)

/*func TestMain(m *testing.M) {
	mySetupFunction()
	retCode := m.Run()
	myTearDownFunction()
	os.Exit(retCode)
}*/

func TestBlacklist(t *testing.T) {
	bl := make(Blacklist)
	peerA := "peerA"
	peerB := "peerB"

	bl.add(peerA)

	if !bl.IsBlacklisted(peerA) {
		t.Error("Didn't blacklist peer but should have")
	}

	if bl.IsBlacklisted(peerB) {
		t.Error("Blacklisted peer but shouldn't have")
	}

}

func TestTempUpdate(t *testing.T) {
	peerA := "peerA"
	peerB := "peerB"

	repTable := make(map[string]int)

	tempUpdateRep(peerA, +1, repTable)
	tempUpdateRep(peerA, +1, repTable)
	tempUpdateRep(peerB, -1, repTable)

	if repTable[peerA] != 2 || repTable[peerB] != -1 {
		t.Error("Wrong reputations")
	}
}

func TestReputationInfo_AddReputations(t *testing.T) {
	peerA := "peerA"
	peerB := "peerB"
	peerC := "peerC"

	repInfo := NewReputationInfo()

	pollKey := PollKey{
		Origin: ecdsa.PublicKey{},
		ID:     3,
	}

	repsA := make(RepOpinions)

	repsA.Trust(peerA)
	repsA.Suspect(peerB)

	pkgA := &ReputationPacket{
		Signer: ecdsa.PublicKey{
			Curve: Curve(),
			X:     &big.Int{},
			Y:     nil,
		},
		PollID:   pollKey,
		Opinions: repsA,
	}

	repsB := make(RepOpinions)

	repsB.Suspect(peerA)
	repsB.Suspect(peerB)
	repsB.Suspect(peerC)

	pkgB := &ReputationPacket{
		Signer: ecdsa.PublicKey{
			Curve: Curve(),
			Y:     &big.Int{},
			X:     nil,
		},
		PollID:   pollKey,
		Opinions: repsB,
	}

	repInfo.AddPeerOpinion(pkgA, pollKey)
	repInfo.AddPeerOpinion(pkgB, pollKey)

	if len(repInfo.PeersOpinions[pollKey]) != 2 {
		t.Error()
	}

	repInfo.AddReputations(pollKey)

	if repInfo.IsBlacklisted(peerA) || !repInfo.IsBlacklisted(peerB) || repInfo.IsBlacklisted(peerC) {
		t.Error()
	}

	fmt.Println(repInfo.Blacklist)

}
