package main

type ReputationTable map[string]*Reputation

type Reputation struct {
	Value int
	IsOld bool
}

func newReputation() *Reputation {
	return &Reputation{
		IsOld: true,
		Value: 0,
	}
}

func (repTable ReputationTable) NextRound() {
	for _, rep := range repTable {
		rep.IsOld = true
	}
}

func (repTable ReputationTable) Trust(peer string) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = newReputation()
	}

	if repTable[peer].IsOld {
		repTable[peer].Value += 1
		repTable[peer].IsOld = false
	}
}

func (repTable ReputationTable) Suspect(peer string) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = newReputation()
	}

	if repTable[peer].IsOld {
		repTable[peer].Value -= 1
		repTable[peer].IsOld = false
	}
}

type Blacklist map[string]bool

func (bl Blacklist) isBlacklisted(peer string) bool {
	return bl[peer]
}

func (bl Blacklist) blacklist(peer string) {
	bl[peer] = true
}

func (repTable ReputationTable) updateBlacklist(bList Blacklist) {
	//TODO threshold
	threshold := -12 //THIS IS NOT OKAY

	for peer, rep := range repTable {
		if rep.Value < threshold {
			bList.blacklist(peer)
		}
	}
}

func updateReputations() {

}
