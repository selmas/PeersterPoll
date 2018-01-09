package pollparty

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"time"
)

const NetworkConvergeDuration = time.Duration(10) * time.Second

type PoolPacketHandler func(PollKey, RunningPollReader)

func VoterHandler(g *Gossiper) func(PollKey, RunningPollReader) {
	return func(id PollKey, r RunningPollReader) {
		_ = <-r.Poll // TODO poll not used?

		tmpKey, err := ecdsa.GenerateKey(Curve(), rand.Reader) // generates tmpKey pair
		if err != nil {
			panic(err)
		}

		voteKey := VoteKey{
			g.KeyPair.PublicKey,
			tmpKey.PublicKey,
		}
		g.SendVoteKey(id, voteKey)

		keys := <-r.VoteKeys

		participants := keys.ToParticipants()
		g.storeParticipants(id, participants)

		position, ok := containsKey(participants, tmpKey.PublicKey)
		if !ok {
			return // we are not part of this vote
		}

		commits := make([]Commitment, 0)
		votes := make([]Vote, 0)

		salt := make(chan [SaltSize]byte)
		option := make(chan string)

		go func() {
			o := <-r.LocalVote
			commit, s := NewCommitment(o)
			salt <- s
			option <- o
			g.SendCommitment(id, commit, participants, tmpKey, position)
			close(salt)
			close(option)
		}()

	Timeout:
		for {
			select {
			case commit := <-r.Commitment:
				commits = append(commits, commit)
				if len(commits) == len(keys.Keys) {
					g.SendVote(id, Vote{
						Salt:   <-salt,
						Option: <-option,
					}, participants, tmpKey, position)
				}
			case vote := <-r.Vote:
				if len(commits) < len(keys.Keys) {
					// TODO ask for status
				}
				votes = append(votes, vote)
			case <-time.After(NetworkConvergeDuration):
				break Timeout
			}
		}

		// TODO locally compute all votes and display to user -> GUI
	}
}

func (g *Gossiper) storeParticipants(id PollKey, participants [][2]*big.Int) {
	g.Polls.Lock()
	defer g.Polls.Unlock()

	pollInfos := g.Polls.m[id.Pack()]
	pollInfos.Participants = participants
	g.Polls.m[id.Pack()] = pollInfos
}

func containsKey(keyArray [][2]*big.Int, publicKey ecdsa.PublicKey) (int, bool) {
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

		var keys []VoteKey
	Timeout:
		for {
			select {
			case k := <-r.VoteKey:
				keys = append(keys, k)
				// TODO check others commits -> bad rep
			case <-time.After(poll.Duration):
				break Timeout
			}
		}

		g.SendVoteKeys(id, VoteKeys{
			Keys: keys,
		})

		// TODO same handling as VoterHandler
	}
}
