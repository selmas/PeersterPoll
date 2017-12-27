package main

import (
	"fmt"
	"net"
	"strconv"
)

func printRumor(gossiper *Gossiper, clientAddr *net.UDPAddr, msg *RumorMessage) {
	fmt.Println(
		"RUMOR origin", msg.Peer.Origin,
		"from", clientAddr.String(),
		"ID", msg.Peer.ID,
	)
	printPeers(gossiper)
}

func printStatus(gossiper *Gossiper, addr *net.UDPAddr, msg *StatusPacket) {
	var str string
	str += "STATUS from " + addr.String()

	for _, s := range msg.Want {
		str += " origin " + s.Identifier
		str += " nextID " + strconv.FormatUint(uint64(s.NextID), 10)
	}
	fmt.Println(str)
	printPeers(gossiper)
}

func printFlippedCoin(addr *net.UDPAddr, typeOfFlip string) {
	fmt.Println("FLIPPED COIN sending", typeOfFlip, "to", addr.String())
}

func printInSyncWith(addr *net.UDPAddr) {
	fmt.Println("IN SYNC WITH", addr.String())
}

func printPeers(gossiper *Gossiper) {
	var str string

	firstPrint := true
	gossiper.Peers.RLock()
	for peer := range gossiper.Peers.Set {
		if firstPrint {
			str += peer
			firstPrint = false
			continue
		}

		str += "," + peer
	}
	gossiper.Peers.RUnlock()

	fmt.Println(str)
}
