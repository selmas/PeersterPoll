package main

import (
	"errors"
	set "github.com/deckarep/golang-set"
)

type RumorMessage struct {
	pollKey      PollKey
	pollQuestion *Poll
	pollVote     *Vote
}

// TODO check rumor message
func checkRumorMessage(msg RumorMessage) error {
	var nilCount uint = 0
	var err error = nil

	/*if msg.pollQuestion != nil {
		nilCount++
		err = checkPoll(*msg.pollQuestion)
	}

	if msg.pollVote != nil {
		nilCount++
		err = checkVote(*msg.pollVote)
	}

	err = checkPollID(msg.pollKey)*/

	if err != nil {
		return errors.New("RumorMessage: " + err.Error())
	}
	if nilCount > 1 {
		return errors.New("too much fields defined")
	} else if nilCount == 0 {
		return errors.New("no field defined")
	}

	return nil
}

type PollStatus struct {
	key   *PollKey
	poll  *Poll
	votes set.Set
}

func checkPeerStatus(msg PollStatus) error {
	if msg.key.PollID < 0 && msg.key.PollOrigin == "" {
		return errors.New("PollStatus: illegal key, ID: " + string(int(msg.key.PollID)) + ", Origin: " + msg.key.PollOrigin)
	}

	return nil
}

type StatusPacket struct {
	Want []PollStatus
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
