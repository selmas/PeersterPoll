package main

import (
	"testing"
	"net"
)

/*func TestMain(m *testing.M) {
	mySetupFunction()
	retCode := m.Run()
	myTeardownFunction()
	os.Exit(retCode)
}*/

func TestSuspect(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")

	repTab.Suspect(addr)
	if repTab[addr].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Suspect(addr)
	if repTab[addr].Value != -1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestTrust(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")

	repTab.Trust(addr)
	if repTab[addr].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Trust(addr)
	if repTab[addr].Value != 1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestSuspectTrust(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")

	repTab.Suspect(addr)
	if repTab[addr].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Trust(addr)
	if repTab[addr].Value != -1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestTrustSuspect(t *testing.T) {

	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")

	repTab.Trust(addr)
	if repTab[addr].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}

	repTab.Suspect(addr)
	if repTab[addr].Value != 1 {
		t.Error("Wrongly updated reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Wrongly updated reputation status")
	}
}

func TestFiveRounds(t *testing.T) {
	repTab := make(ReputationTable)

	addr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8000")


	repTab.Suspect(addr)
	if repTab[addr].Value != -1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}


	repTab.NextRound()

	repTab.Trust(addr)
	if repTab[addr].Value != 0 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}


	repTab.NextRound()

	repTab.Trust(addr)
	if repTab[addr].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}


	repTab.NextRound()

	repTab.Trust(addr)
	if repTab[addr].Value != 2 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}


	repTab.NextRound()

	repTab.Suspect(addr)
	if repTab[addr].Value != 1 {
		t.Error("Didn't update reputation value")
	}
	if repTab[addr].IsOld {
		t.Error("Didn't update reputation status")
	}

}
