package pollparty

// Reputation Opinions ---------------------------------------------------------------------------

type RepOpinions map[string]int

// TODO Call this in the beginning of any new vote
func NewRepOpinions(peers []string) RepOpinions {
	repOps := make(RepOpinions)
	for _, peer := range peers {
		repOps.Trust(peer)
	}

	return repOps
}

func (opinions RepOpinions) Suspect(peer string) {
	// also add peer to blacklist
	opinions[peer] = -1
}

// TODO trust all new peers to add them to the map
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

// TODO Do not forget to make sure that if two different received opinions are signed with the same key,
// none of those two should be taken into account

// Blacklist -------------------------------------------------------------------------------------

type Blacklist map[string]bool

func (bList Blacklist) IsBlacklisted(peer string) bool {
	return bList[peer]
}

func (bList Blacklist) add(peer string) {
	bList[peer] = true
}

func (bList Blacklist) Update(repTable map[string]int) {
	for peer, rep := range repTable {
		if rep < 0 {
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

// Reputation Info -------------------------------------------------------------------------------

type ReputationInfo struct {
	Opinions  RepOpinions
	Blacklist Blacklist

	PeersOpinions map[PollKey][]RepOpinions
}

func (repInfo ReputationInfo) AddPeerOpinion(opinion RepOpinions, pollID PollKey) {
	repInfo.PeersOpinions[pollID] = append(repInfo.PeersOpinions[pollID], opinion)
}

func (repInfo ReputationInfo) AddReputations(pollID PollKey) {

	repTable := make(map[string]int)

	for _, peerOpinions := range repInfo.PeersOpinions[pollID] {

		// If a peer has an invalid opinion of another peer
		// none of its opinions will be taken into account
		if peerOpinions.hasInvalidOpinion() {
			continue
		}

		for peer, rep := range peerOpinions {
			tempUpdateRep(peer, rep, repTable)
		}
	}

	repInfo.Blacklist.Update(repTable)
}

func tempUpdateRep(peer string, rep int, repTable map[string]int) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = 0
	}

	repTable[peer] += rep
}

func (repInfo ReputationInfo) IsBlacklisted(peer string) bool {
	return repInfo.Blacklist.IsBlacklisted(peer)
}

func (repInfo ReputationInfo) Suspect(peer string) {
	repInfo.Opinions.Suspect(peer)
	repInfo.Blacklist.add(peer)
}

