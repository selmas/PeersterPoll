package main

import (
	"testing"
	"net"
)

/*func TestMain(m *testing.M) {
	mySetupFunction()
	retCode := m.Run()
	myTearDownFunction()
	os.Exit(retCode)
}*/

func TestSuspect(t *testing.T) {

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

}

func TestBlacklist(t *testing.T) {
	bl := make(Blacklist)
	peerA := "peerA"
	peerB := "peerB"

	bl.blacklist(peerA)

	if !bl.isBlacklisted(peerA) {
		t.Error("Didn't blacklist peer but should have")
	}

	if bl.isBlacklisted(peerB) {
		t.Error("Blacklisted peer but shouldn't have")
	}

}

