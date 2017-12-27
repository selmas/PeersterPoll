package main

import (
	"errors"
)

type PeerMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func checkPeerMessage(msg PeerMessage) error {
	if msg.ID == 0 && msg.Text != "" {
		return errors.New("PeerMessage: text in routing msg")
	}

	return nil
}

type RumorMessage struct {
	PeerMessage
}

func checkRumorMessage(msg RumorMessage) error {
	err := checkPeerMessage(msg.PeerMessage)
	if err != nil {
		return errors.New("RumorMessage: " + err.Error())
	}

	return nil
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

func checkPeerStatus(msg PeerStatus) error {
	if msg.NextID == 0 {
		return errors.New("PeerStatus: want routing msg")
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

func checkPollPacket(msg PollPacket) error {
	var nilCount uint = 0
	var err error = nil

	// TODO check Votes field
	/*if msg.Vote != nil {
		nilCount++
		//err = checkVoteMessage(*msg.Vote)
	}*/

	if msg.Question != nil {
		nilCount++
		// TODO check poll message
		//err = checkPollMessage(*msg.Question)
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

type GossipPacket struct {
	Rumor  *RumorMessage
	Status *StatusPacket
	Poll   *PollPacket
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

	if pkg.Poll != nil {
		nilCount++
		err = checkPollPacket(*pkg.Poll)
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
