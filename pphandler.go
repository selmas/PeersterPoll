package pollparty

import (
	"time"
	"crypto/rand"
	"crypto/ecdsa"
)

type PoolPacketHandler func(PollKey, RunningPollReader)

func VoterHandler(g *Gossiper) func(PollKey, RunningPollReader) {
	return func(id PollKey, r RunningPollReader) {
		poll, ok := <-r.Poll
		if !ok {
			return
		}

		// TODO rumor poll

		// TODO GUI for now, force first choice
		assert(len(poll.VoteOptions) > 0)
		option := poll.VoteOptions[0]

		tmpKeyPair, ok := ecdsa.GenerateKey(curve, rand.Reader) // generates key pair
		if !ok {
			return
		}
		g.SendRegister(id, tmpKeyPair.PublicKey)

		// TODO return list of participants, type [][]*big.Int
		particiants, ok := <-r.tmpKeys
		if !ok {
			return
		}

		position := -1
		for index, key := range participants {
			if key[0].Cmp(tmpKeyPair.X) == 0 && key[1].Cmp(tmpKeyPair.Y) == 0 {
				position = index
			}
		}
		if position == -1 {
			// TODO suspect registery? I'm not included in poll
			return
		}

		commit := NewCommitment(option)
		g.SendCommitment(id, commit, participants, tmpKeyPair, position)

		commits, ok := <-r.PollCommitments
		if !ok {
			return
		}

		if !commits.Has(commit) {
			return // to avoid loading network, we abort here
		}

		g.SendVote(id, option, participants, tmpKeyPair, position)

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
