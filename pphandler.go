package main

import (
	"log"
)

type PoolPacketHandler func(PollKey, RunningPollReader)

func VoterHandler(g *Gossiper) func(PollKey, RunningPollReader) {
	return func(id PollKey, r RunningPollReader) {
		poll := <-r.Poll

		if poll.IsTooLate() {
			log.Println("poll came in too late")
			return
		}

		// TODO for now, force first choice
		assert(len(poll.VoteOptions) > 0)
		option := poll.VoteOptions[0]

		commit := NewCommitment(option)
		g.SendCommitment(id, commit)

		commits := <-r.PollCommitments
		if !commits.Has(commit) {
			return // to avoid loading network, we abort here
		}

		g.SendVote(id, option)
	}
}
