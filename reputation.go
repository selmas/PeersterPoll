package main

import "strconv"

// Reputation Opinions ---------------------------------------------------------------------------

type RepOpinions map[string]int

func (opinions RepOpinions) Suspect(peer string) {
	opinions[peer] = -1
}

func (opinions RepOpinions) Trust(peer string) {
	// Opinion can only change from trusting to suspecting,
	// not the other way around
	if opinions[peer] != -1 {
		opinions[peer] = 1
	}
}

func (opinions RepOpinions) hasInvalidOpinion() bool {
	for _, rep := range opinions {
		if rep != 1 && rep != -1 {
			return true
		}
	}
	return false
}

// TODO remove the map because we won't be able to associate each opinion with its peer anymore
// to store a peer's opinions and associate it with the peer
type OpinionsMap map[string]RepOpinions

func (opinions OpinionsMap) Add(peer string, opn RepOpinions) {
	_, exists := opinions[peer]

	// if a peer tries to send its opinion two times to manipulate reputations
	// its opinion is not taken into account at all
	if exists {
		delete(opinions, peer)
	} else {
		opinions[peer] = opn
	}
}

// Reputation ------------------------------------------------------------------------------------

type Reputation struct {
	Value int
	Peer  string
}

type ReputationMap map[string]*Reputation

type ReputationTable struct {
	Reputations ReputationMap
	Threshold   int
}

func NewReputationTable(threshold int) ReputationTable {
	return ReputationTable{
		Reputations: make(ReputationMap),
		Threshold:   threshold,
	}
}

/*
Input an OpinionsMap with all the received opinions.
They will all be added and this will update the reputation table.
Reputations above 0 are not allowed so after adding everything,
all the reputations above 0 are set to 0.
*/
func (repTable ReputationTable) AddReputations(allOpinions OpinionsMap) {
	//TODO set threshold according to number of peers that gave opinion?

	for _, peerOpinions := range allOpinions {

		// If a peer has an invalid opinion of another peer
		// none of its opinions will be taken into account
		if peerOpinions.hasInvalidOpinion() {
			continue
		}

		for peer, rep := range peerOpinions {
			repTable.tempUpdateRep(peer, rep)
		}
	}

	repTable.regularize()
}

func (repTable ReputationTable) tempUpdateRep(peer string, rep int) {
	_, exists := repTable.Reputations[peer]

	if !exists {
		repTable.Reputations[peer] = &Reputation{
			Value: 0,
			Peer:  peer,
		}
	}

	repTable.Reputations[peer].Value += rep
}

func (repTable ReputationTable) regularize() {
	for peer, rep := range repTable.Reputations {
		if rep.Value > 0 {
			repTable.Reputations[peer].Value = 0
		}
	}
}

func (repTable ReputationTable) String() string {
	str := "Threshold: " + strconv.Itoa(repTable.Threshold) + "\n"

	for peer, rep := range repTable.Reputations {
		str += peer + "\t...\t" + strconv.Itoa(rep.Value) + "\n"
	}

	return str
}

// Blacklist -------------------------------------------------------------------------------------

type Blacklist map[string]bool

func (bList Blacklist) IsBlacklisted(peer string) bool {
	return bList[peer]
}

func (bList Blacklist) add(peer string) {
	bList[peer] = true
}

// This should be used after having a final reputation table considering the opinions of the other peers
func (bList Blacklist) UpdateBlacklist(repTable ReputationTable) {
	for peer, rep := range repTable.Reputations {
		if rep.Value < repTable.Threshold {
			bList.add(peer)
		}
	}
}

func (bList Blacklist) String() string {
	str := "Peer\t\tBlacklisted?\n"
	for peer, status := range bList {
		str += peer + "\t\t"
		if status {
			str += "yes\n"
		} else {
			str += "no\n"
		}
	}

	return str
}

// ... -------------------------------------------------------------------------------------------

type ReputationPacket struct {
	Opinions RepOpinions

	//TODO change
	PollID    uint64 //TODO change according to representation
	Signature []byte //TODO change according to representation
}

func updateReputations() {

}
