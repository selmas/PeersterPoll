package pollparty

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"time"
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
		assert(len(poll.Options) > 0)
		option := poll.Options[0]

		tmpKeyPair, err := ecdsa.GenerateKey(Curve(), rand.Reader) // generates key pair
		if err != nil {
			return
		}
		//g.SendRegister(id, tmpKeyPair.PublicKey)

		// TODO return list of participants, type [][]*big.Int
		//participants, ok := <-r.tmpKeys
		var participants [][]*big.Int // remove again!!! just for testing
		if !ok {
			return
		}
		storeParticipants(g, id, participants)

		position, ok := containsKey(participants, tmpKeyPair.PublicKey)
		if !ok {
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

		// TODO Block until received all commits or timeout

		g.SendVote(id, Vote{
			Salt:   [20]byte{}, // TODO empty salt, nice
			Option: option,
		}, participants, tmpKeyPair, position)

		// TODO save to gossiper
		// TODO wait for timeout or to receive all votes
		// TODO locally compute all votes and display to user -> GUI
	}
}

func storeParticipants(g *Gossiper, id PollKey, participants [][]*big.Int) {
	g.Polls.Lock()
	pollInfos := g.Polls.m[id]
	pollInfos.Participants = participants
	g.Polls.m[id] = pollInfos
	g.Polls.Unlock()
}

func containsKey(keyArray [][]*big.Int, publicKey ecdsa.PublicKey) (int, bool) {
	for index, key := range keyArray {
		if key[0].Cmp(publicKey.X) == 0 && key[1].Cmp(publicKey.Y) == 0 {
			return index, true
		}
	}
	return -1, false
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
