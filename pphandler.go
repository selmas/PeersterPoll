package pollparty

import (
	"time"
)

type PoolPacketHandler func(PollKey, RunningPollReader)

func VoterHandler(g *Gossiper) func(PollKey, RunningPollReader) {
	return func(id PollKey, r RunningPollReader) {
		poll, ok := <-r.Poll
		if !ok {
			return
		}

		// TODO for now, force first choice
		assert(len(poll.VoteOptions) > 0)
		option := poll.VoteOptions[0]

		commit := NewCommitment(option)
		g.SendCommitment(id, commit)

		commits, ok := <-r.PollCommitments
		if !ok {
			return
		}

		if !commits.Has(commit) {
			return // to avoid loading network, we abort here
		}

		g.SendVote(id, option)

		// TODO save to gossiper
	}
}

func MasterHandler(g *Gossiper) func(PollKey, RunningPollReader) {
	return func(id PollKey, r RunningPollReader) {
		poll, ok := <-r.Poll
		if !ok {
			return
		}

		g.SendPoll(id, poll)

		var commits []Commitment

	Timeout:
		for {
			select {
			case commit := <-r.Commitments:
				commits = append(commits, commit)
				// TODO check others commits -> bad rep
			case <-time.After(poll.Duration):
				break Timeout
			}
		}

		g.SendPollCommitments(id, PollCommitments{
			Commitments: commits,
		})

		//votes := <-r.Votes // TODO bad rep
	}
}
