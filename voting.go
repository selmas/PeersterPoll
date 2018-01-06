package main

import (
	"net"
)

/*type Poll struct {
	Question    string
	VoteOptions []string
	//Participants set.Set // TODO set of public keys if we add support for linkable ring signatures
	StartTime time.Time
	Duration  time.Duration // After duration has passed, can no longer participate in votes
}

type Vote struct {
	VoteOrigin   string
	CommittedVal CommitPedersen
	OpenedVal    OpenPedersen
}*/

func handleClosedPoll(gossiper *Gossiper, rumor PollPacket, fromPeer *net.UDPAddr) {
	// TODO reach consensus on committed votes
	// TODO open commitment
	// TODO tally poll and display result
}

// this method processes the poll locally
// first the poll gets stored
// if the poll was already stored, check if we received a newer version
// the poll storage is updated if the received version is newer
// if the sender of the poll doesn't have the newest version of the poll
// reply to sender with the stored version of the poll
func handleOpenPoll(gossiper *Gossiper, msg PollPacket, fromPeer *net.UDPAddr) {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	/*storedPoll, isStored := gossiper.Polls.m[msg.pollKey]

	if !isStored {
		votes := set.NewSet()
		votes.Add(msg.pollVote)

		gossiper.Polls.m[msg.pollKey] = &VoteSet{msg.pollQuestion, votes}
		sendRumor(gossiper, &msg, fromPeer)
	} else {
		// Update stored participant list
		storedPoll.poll.Participants = storedPoll.poll.Participants.Union(msg.pollQuestion.
			Participants)

		// if sender is missing participants, send updated poll back
		if storedPoll.poll.Participants.Difference(msg.pollQuestion.Participants).Cardinality() != 0 {
			// Update participant list in rumor message
			msg.pollQuestion.Participants = storedPoll.poll.Participants
			writeMsgToUDP(gossiper.Server, fromPeer, &msg, nil)
		}

		sendRumor(gossiper, &msg, fromPeer)
	}*/
}

func handleClientVote(vote *Vote, key PollKey, gossiper Gossiper) {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	/*poll, succ := gossiper.Polls.m[key]

	if !succ {
		log.Fatal("Invalide client request received: Poll not found")
		return
	}
	gossiper.Polls.m[key].votes.Add(vote)
	sendRumor(&gossiper, &PollPacket{key, poll.poll, vote}, gossiper.Server.Addr)*/
}
