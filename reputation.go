package main

import (
	"net"
)

type ReputationTable map[*net.UDPAddr]*Reputation

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

//func ResetRepTable(repTable map[*net.UDPAddr]*Reputation) {}

func (repTable ReputationTable) NextRound() {
	for _, rep := range repTable {
		rep.IsOld = true
	}
}

func (repTable ReputationTable) Trust(peer *net.UDPAddr) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = newReputation()
	}

	if repTable[peer].IsOld {
		repTable[peer].Value += 1
		repTable[peer].IsOld = false
	}
}

func (repTable ReputationTable) Suspect(peer *net.UDPAddr) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = newReputation()
	}

	if repTable[peer].IsOld {
		repTable[peer].Value -= 1
		repTable[peer].IsOld = false
	}
}

func isBlacklisted(peer *net.UDPAddr) bool {
	return false
}

func blacklist(peer *net.UDPAddr) {

}

func updateReputations() {

}
