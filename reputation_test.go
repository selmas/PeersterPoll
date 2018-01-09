package pollparty

import (
	"testing"
	"crypto/ecdsa"
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

/*func TestOpinions(t *testing.T) {
	repTabA := NewReputationTable()
	repTabB := NewReputationTable()
	repTabC := NewReputationTable()

	peerA := "peerA"
	peerB := "peerB"
	peerC := "peerC"

	opinionsA := make(RepOpinions)
	opinionsB := make(RepOpinions)
	opinionsC := make(RepOpinions)

	opinionsA.Trust(peerA)
	opinionsA.Trust(peerB)
	opinionsA.Suspect(peerC)

	opinionsB.Trust(peerA)
	opinionsB.Trust(peerB)
	opinionsB.Suspect(peerC)

	opinionsC.Suspect(peerA)
	opinionsC.Suspect(peerB)
	opinionsC.Trust(peerC)

	listOpinions := make([]RepOpinions, 0)
	listOpinions = append(listOpinions, opinionsA)
	listOpinions = append(listOpinions, opinionsB)
	listOpinions = append(listOpinions, opinionsC)

	repTabA.AddReputations(listOpinions)
	repTabB.AddReputations(listOpinions)
	repTabC.AddReputations(listOpinions)

	repTabA.AddReputations(listOpinions)
	repTabB.AddReputations(listOpinions)
	repTabC.AddReputations(listOpinions)

	repTabA.AddReputations(listOpinions)
	repTabB.AddReputations(listOpinions)
	repTabC.AddReputations(listOpinions)

	repTabA.AddReputations(listOpinions)
	repTabB.AddReputations(listOpinions)
	repTabC.AddReputations(listOpinions)

	//fmt.Println("Table A\n", repTabA)
	//fmt.Println("\nTable B\n", repTabB)
	//fmt.Println("\nTable C\n", repTabC)

	bl := make(Blacklist)

	bl.UpdateBlacklist(repTabA)

	//fmt.Println(bl)

	if bl.IsBlacklisted(peerA) || bl.IsBlacklisted(peerB) {
		t.Error("Blacklisted peer but shouldn't have")
	}

	if !bl.IsBlacklisted(peerC) {
		t.Error("Didn't blacklist peer but should have")
	}

	if repTabA.Reputations[peerA].Value != 0 ||
		repTabA.Reputations[peerB].Value != 0 ||
		repTabA.Reputations[peerC].Value != -4 {

		t.Error("Wrong reputations")
	}
}*/

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

	repInfo := ReputationInfo{
		Opinions:  make(RepOpinions),
		Blacklist: make(Blacklist),
		PeersOpinions: make(map[PollKey][]RepOpinions),
	}

	pollKey := PollKey{
		Origin: ecdsa.PublicKey{},
		ID:     3,
	}

	repsA := make(RepOpinions)
	repsB := make(RepOpinions)

	repsA[peerA] = +1
	repsA[peerB] = +1

	repsB[peerA] = +1
	repsB[peerB] = -1

	repInfo.AddPeerOpinion(repsA,pollKey)
	repInfo.AddPeerOpinion(repsB,pollKey)

	if len(repInfo.PeersOpinions[pollKey]) != 2 {
		t.Error()
	}

	repInfo.AddReputations(pollKey)

	if repInfo.IsBlacklisted(peerA) || repInfo.IsBlacklisted(peerB) {
		t.Error()
	}

	pollKey = PollKey{
		Origin: ecdsa.PublicKey{},
		ID:     4,
	}

	repsA[peerA] = +5
	repsA[peerB] = -1

	repsB[peerA] = -1
	repsB[peerB] = +1

	repsC := make(RepOpinions)
	repsC[peerA] = -1
	repsC[peerB] = -1

	repInfo.AddPeerOpinion(repsA,pollKey)
	repInfo.AddPeerOpinion(repsB,pollKey)
	repInfo.AddPeerOpinion(repsC,pollKey)

	repInfo.AddReputations(pollKey)

	if !repInfo.IsBlacklisted(peerA) || repInfo.IsBlacklisted(peerB) {
		t.Error()
	}

}
