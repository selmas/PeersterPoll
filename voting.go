package main

import (
	"time"
	"net"
	set "github.com/deckarep/golang-set"
)

// TODO: add option for origin node to sign / commit Question to guarantee integrity of it
type Poll struct {
	Question     string
	VoteOptions  []string
	Participants set.Set // TODO needed (could hold public keys if we use linkable ring signatures)
	StartTime    time.Time
	Duration     time.Duration // After duration has passed, can no longer participate in vote
}

type Vote struct {
	VoteOrigin   string
	CommittedVal CommitPedersen
	OpenedVal    OpenPedersen
}

// this method processes the poll locally
// first the poll gets stored
// if the poll was already stored, check if we received a newer version
// the poll storage is updated if the received version is newer
// if the sender of the poll doesn't have the newest version of the poll
// reply to sender with the stored version of the poll
func handlePoll(gossiper *Gossiper, msg RumorMessage, fromPeer *net.UDPAddr) {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	storedPoll, isStored := gossiper.Polls.Set[msg.pollKey]

	if !isStored {
		gossiper.Polls.Set[msg.pollKey] = &msg
		sendRumor(gossiper, &msg, fromPeer)
	} else {
		// Update stored participant list
		storedPoll.pollQuestion.Participants = storedPoll.pollQuestion.Participants.Union(msg.pollQuestion.Participants)
		// if sender is missing participants, send updated poll back
		if storedPoll.pollQuestion.Participants.Difference(msg.pollQuestion.Participants).Cardinality() != 0 {
			writeMsgToUDP(gossiper.Server, fromPeer, storedPoll, nil)
		}
		sendRumor(gossiper, storedPoll, fromPeer)
	}
}

func handleVote(vote *Vote) {

}
