package main

import (
	"testing"
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

func TestOpinions(t *testing.T) {
	repTabA := NewReputationTable(-3)
	repTabB := NewReputationTable(-3)
	repTabC := NewReputationTable(-3)

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

	listOpinions := make(OpinionsMap)
	listOpinions.Add(peerA, opinionsA)
	listOpinions.Add(peerB, opinionsB)
	listOpinions.Add(peerC, opinionsC)

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
}

func TestDoubleOpinion(t *testing.T) {
	repTabA := NewReputationTable(-3)
	repTabB := NewReputationTable(-3)
	repTabC := NewReputationTable(-3)

	peerA := "peerA"
	peerB := "peerB"
	peerC := "peerC"

	opinionsA := make(RepOpinions)
	opinionsB := make(RepOpinions)
	opinionsC1 := make(RepOpinions)
	opinionsC2 := make(RepOpinions)

	opinionsA.Trust(peerA)
	opinionsA.Trust(peerB)
	opinionsA.Suspect(peerC)

	opinionsB.Trust(peerA)
	opinionsB.Trust(peerB)
	opinionsB.Suspect(peerC)

	opinionsC1.Suspect(peerA)
	opinionsC1.Suspect(peerB)
	opinionsC1.Trust(peerC)
	opinionsC2.Suspect(peerA)
	opinionsC2.Suspect(peerB)
	opinionsC2.Trust(peerC)

	listOpinions := make(OpinionsMap)
	listOpinions.Add(peerA, opinionsA)
	listOpinions.Add(peerB, opinionsB)
	listOpinions.Add(peerC, opinionsC1)
	listOpinions.Add(peerC, opinionsC2)

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
		repTabA.Reputations[peerC].Value != -8 {

		t.Error("Wrong reputations")
	}
}
