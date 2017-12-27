package main

import (
	"errors"
	"net"
)

type PeerMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func (m PeerMessage) IsRouting() bool {
	return m.ID == 0
}

func checkPeerMessage(msg PeerMessage) error {
	if msg.ID == 0 && msg.Text != "" {
		return errors.New("PeerMessage: text in routing msg")
	}

	return nil
}

type RumorMessage struct {
	PeerMessage
	LastIP   *net.IP
	LastPort *int
}

func checkRumorMessage(msg RumorMessage) error {
	err := checkPeerMessage(msg.PeerMessage)
	if err != nil {
		return errors.New("RumorMessage: " + err.Error())
	}

	// boolean xor
	if (msg.LastIP == nil) != (msg.LastPort == nil) {
		return errors.New("RumorMessage: half Last* defined")
	}

	return nil
}

func (r *RumorMessage) SetSender(sender net.UDPAddr) {
	r.LastIP = &sender.IP
	r.LastPort = &sender.Port
}

func (r RumorMessage) GetLastAddr() *net.UDPAddr {
	if r.IsDirect() {
		return nil
	}

	return &net.UDPAddr{
		IP:   *r.LastIP,
		Port: *r.LastPort,
	}
}

func (r RumorMessage) IsDirect() bool {
	return r.LastIP == nil && r.LastPort == nil
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

type PrivateMessage struct {
	PeerMessage
	Dest     string
	HopLimit uint32
}


func checkPollMessage(msg PollMessage) error {
	// TODO: add checks for PollMessage
}

func checkPrivateMessage(msg PrivateMessage) error {
	err := checkPeerMessage(msg.PeerMessage)
	if err != nil {
		return errors.New("Private: " + err.Error())
	}

	return nil
}

type GossipPacket struct {
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
	Poll	*PollMessage
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

	if pkg.Private != nil {
		nilCount++
		err = checkPrivateMessage(*pkg.Private)
	}

	if pkg.Poll != nil {
		nilCount++
		err = checkPollMessage(*pkg.Poll)
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
