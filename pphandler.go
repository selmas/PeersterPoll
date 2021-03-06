package pollparty

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"
)

const NetworkConvergeDuration = time.Duration(3) * time.Second

type PoolPacketHandler func(PollKey, ecdsa.PrivateKey, RunningPollReader)

func VoterHandler(g *Gossiper) PoolPacketHandler {
	return func(id PollKey, key ecdsa.PrivateKey, r RunningPollReader) {
		_ = <-r.Poll
		log.Println("Voter: new poll:", id.String())

		voteKey := VoteKey{
			tmpKey:    key.PublicKey,
		}
		g.SendVoteKey(id, voteKey)
		log.Println("Voter: send back key")

		keys := <-r.VoteKeys
		log.Println("Voter: got keys")

		commonHandler("Voter", g, id, key, keys, r)
	}
}

func (g *Gossiper) storeParticipants(id PollKey, participants [][2]big.Int) {
	g.Polls.Lock()
	defer g.Polls.Unlock()

	pollInfo := g.Polls.m[id.Pack()]
	pollInfo.Participants = participants
	g.Polls.m[id.Pack()] = pollInfo
}

func containsKey(keyArray [][2]big.Int, tmpKey ecdsa.PublicKey) (int, bool) {
	for index, key := range keyArray {
		if key[0].Cmp(tmpKey.X) == 0 && key[1].Cmp(tmpKey.Y) == 0 {
			return index, true
		}
	}

	return -1, false
}

func MasterHandler(g *Gossiper) PoolPacketHandler {
	return func(id PollKey, key ecdsa.PrivateKey, r RunningPollReader) {
		poll := <-r.Poll
		log.Println("Master: new poll:", id.String())

		g.SendPoll(id, poll)

		keysMap := make(map[VoteKeyMap]bool)
		mapKey := VoteKey{
			tmpKey:    key.PublicKey,
		}.Pack()
		keysMap[mapKey] = true

	Timeout:
		for {
			select {
			case k := <-r.VoteKey:
				keysMap[k.Pack()] = true

			case <-time.After(poll.Duration):
				break Timeout
			}
		}

		var keys []VoteKey
		for k := range keysMap {
			keys = append(keys, k.Unpack())
		}

		voteKeys := VoteKeys{
			Keys: keys,
		}
		g.SendVoteKeys(id, voteKeys)
		log.Printf("Master: send %d keys", len(keys))

		commonHandler("Master", g, id, key, voteKeys, r)
	}
}

func commonHandler(logName string, g *Gossiper, id PollKey, key ecdsa.PrivateKey, keys VoteKeys, r RunningPollReader) {
	participants := keys.ToParticipants()
	g.storeParticipants(id, participants)

	position, ok := containsKey(participants, key.PublicKey)
	if !ok {
		log.Printf("%s: not considered for this vote, abort", logName)
		return
	}

	commits := make([]Commitment, 0)
	votes := make([]Vote, 0)

	salt := make(chan [SaltSize]byte)
	option := make(chan string)

	g.Reputations.AddTablesWait[id] = make(chan bool)

	go func() {
		o := <-r.LocalVote
		log.Printf("%s: got local vote for \"%s\"", logName, o)

		commit, s := NewCommitment(o)
		g.SendCommitment(id, commit, participants, key, position)
		log.Printf("%s: send commit for \"%s\"", logName, o)

		salt <- s
		option <- o

		close(salt)
		close(option)
	}()

	voteSent := false
	timedout := false
	timeout := time.After(NetworkConvergeDuration)
	for {
		select {
		case commit := <-r.Commitment:
			if !timedout {
				commits = append(commits, commit)
			} // do not accept commits after timeout, to prevent influencing
			if len(commits) == len(keys.Keys) {
				g.SendVote(id, Vote{
					Salt:   <-salt,
					Option: <-option,
				}, participants, key, position)
				log.Printf("%s: send vote", logName)
				voteSent = true
			}
		case vote := <-r.Vote:
			if len(commits) < len(keys.Keys) || timedout {
				myStatus := getStatus(g).toBase()
				writeMsgToUDP(g.Server, vote.Sender, nil, &myStatus, nil, nil)
				time.Sleep(time.Duration(250) * time.Millisecond)
				if len(commits) < len(keys.Keys) {
					g.Reputations.Suspect(vote.Sender.String())
				}
			}
			votes = append(votes, vote.Vote)
		case <-timeout:
			log.Printf("%s: timeout", logName)
			timedout = true
			if !voteSent {
				g.SendVote(id, Vote{
					Salt:   <-salt,
					Option: <-option,
				}, participants, key, position)
				log.Printf("%s: send vote at timeout", logName)
			}
		}

		if len(votes) == len(keys.Keys) && len(commits) == len(keys.Keys) {
			break
		}
	}

	log.Printf("%s: pool's closed", logName)

	UpdateReputations(g, id)
}
