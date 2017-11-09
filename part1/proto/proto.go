package proto

import (
	"errors"
	"net"
)

type PeerMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func checkPeerMessage(msg *PeerMessage) error {
	if msg.Origin == "" {
		return errors.New("empty origin")
	}

	return nil
}

type RumorMessage struct {
	PeerMessage
	LastIP   *net.IP
	LastPort *int
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type PrivateMessage struct {
	PeerMessage
	Dest     string
	HopLimit uint32
}

type GossipPacket struct {
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
}

func CheckGossipPacket(pkg *GossipPacket) error {
	var nilCount uint = 0
	if pkg.Rumor != nil {
		nilCount++

		err := checkPeerMessage(&pkg.Rumor.PeerMessage)
		if err != nil {
			return errors.New("Rumor: " + err.Error())
		}
	}

	if pkg.Status != nil {
		nilCount++
	}

	if pkg.Private != nil {
		nilCount++

		err := checkPeerMessage(&pkg.Private.PeerMessage)
		if err != nil {
			return errors.New("Private: " + err.Error())
		}
	}

	if nilCount > 1 {
		return errors.New("too much fields defined")
	} else if nilCount == 0 {
		return errors.New("no field defined")
	}

	return nil
}
