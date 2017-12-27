package main

import (
	"errors"
	set "github.com/deckarep/golang-set"
)

type RumorMessage struct {
	pollKey      PollKey
	pollQuestion *Poll
	pollVote     set.Set
}

// TODO check rumor message
func checkRumorMessage(msg RumorMessage) error {
	/*err := checkMessageID(msg.pollKey)
	if err != nil {
		return errors.New("RumorMessage: " + err.Error())
	}

	err = checkPollPacket(msg.pollQuestion)
	if err != nil {
		return errors.New("RumorMessage: " + err.Error())
	}
*/
	return nil
}

type PeerStatus struct {
	pollKey         *PollKey
	participantList set.Set
	pollVote     	set.Set
}

func checkPeerStatus(msg PeerStatus) error {
	if msg.pollKey.PollID < 0 && msg.pollKey.PollOrigin == ""{
		return errors.New("PeerStatus: illegal pollKey, ID: " + string(int(msg.pollKey.PollID)) + ", Origin: " + msg.pollKey.PollOrigin)
	}

	return nil
}

type StatusPacket struct {
	Want []PeerStatus
}

func checkStatusPacket(msg StatusPacket) error {
	for _, status := range msg.Want {
		err := checkPeerStatus(status)
		if err != nil {
			return errors.New("StatusPacket: " + err.Error())
		}
	}
	return nil
}

type GossipPacket struct {
	Rumor  *RumorMessage
	Status *StatusPacket
}

func CheckGossipPacket(pkg *GossipPacket) error {
	var nilCount uint = 0
	var err error = nil

	if pkg.Rumor != nil {
		nilCount++
		err = checkRumorMessage(*pkg.Rumor)
	}

	if pkg.Status != nil {
		nilCount++
		err = checkStatusPacket(*pkg.Status)
	}

	if err != nil {
		return err
	}

	if nilCount > 1 {
		return errors.New("too much fields defined")
	} else if nilCount == 0 {
		return errors.New("no field defined")
	}

	return nil
}
