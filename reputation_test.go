package pollparty

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

	listOpinions := make([]RepOpinions,0)
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
}
