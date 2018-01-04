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

/*func TestSuspect(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")
	peer := addr.String()

	repTab.Suspect(peer)
	if repTab[peer].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Suspect(peer)
	if repTab[peer].Value != -1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestTrust(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")
	peer := addr.String()

	repTab.Trust(peer)
	if repTab[peer].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Trust(peer)
	if repTab[peer].Value != 1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestSuspectTrust(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")
	peer := addr.String()

	repTab.Suspect(peer)
	if repTab[peer].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Trust(peer)
	if repTab[peer].Value != -1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestTrustSuspect(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")
	peer := addr.String()

	repTab.Trust(peer)
	if repTab[peer].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Suspect(peer)
	if repTab[peer].Value != 1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestFiveRounds(t *testing.T) {
	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")
	peer := addr.String()

	repTab.Suspect(peer)
	if repTab[peer].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.NextRound()

	repTab.Trust(peer)
	if repTab[peer].Value != 0 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.NextRound()

	repTab.Trust(peer)
	if repTab[peer].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.NextRound()

	repTab.Trust(peer)
	if repTab[peer].Value != 2 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.NextRound()

	repTab.Suspect(peer)
	if repTab[peer].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[peer].IsOld {
		t.Error("Didn't update reputation status")
	}

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

	// round 1
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

	listOpinions := append(make([]RepOpinions, 0), opinionsA, opinionsB, opinionsC)

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
}
