package main

import (
	"time"
)

// TODO: add option for origin node to sign / commit Question to guarantee integrity of it
type Poll struct {
	Question      string
	AnswerOptions []string
	Participants  []string // TODO needed (could hold public keys if we use linkable ring signatures)
	StartTime     time.Time
	Duration      time.Duration // After duration has passed, can no longer participate in vote
}

type Vote struct {
	VoteOrigin   string
	CommittedVal CommitPedersen
	OpenedVal    OpenPedersen
}

type PollPacket struct {
	PollOrigin string
	PollID     uint32
	Question   *Poll
	Votes      map[string]*Vote // mapping from origin to casted Vote (TODO if add anonymity, replace string=VoteOrigin with LRS Tag)
}

// this method forwards the poll packet without locally processing it
// called when the time to participate in vote has passed
func forwardPoll(gossiper *Gossiper, pollPkt PollPacket) {

}

// this method processes the poll locally
// first the poll gets stored
// if the poll was already stored, check if we received a newer version
// the poll storage is updated if the received version is newer
// if the sender of the poll doesn't have the newest version of the poll
// reply to sender with the stored version of the poll
func handlePoll(gossiper *Gossiper, pollPkt PollPacket) {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	/*storedPoll, isStored := gossiper.Polls.Set[PollKey{pollPkt.PollOrigin, pollPkt.PollID}]

	if !isStored {
		gossiper.Polls.Set[PollKey{pollPkt.PollOrigin, pollPkt.PollID}] = pollPkt
		sendRumor() // TODO forward
	}*/
}
func handleVote(vote *Vote) {

}
