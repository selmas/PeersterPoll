package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func printRumor(gossiper *Gossiper, clientAddr *net.UDPAddr, msg *RumorMessage) {
	fmt.Println(
		"Poll origin", msg.pollKey.PollOrigin,
		"from", clientAddr.String(),
		"ID", msg.pollKey.PollID,
		"Question", msg.pollQuestion.Question,
		"with options", strings.Join(msg.pollQuestion.VoteOptions, ","),
	)
	printPeers(gossiper)
}

func printStatus(gossiper *Gossiper, addr *net.UDPAddr, msg *StatusPacket) {
	var str string
	str += "STATUS from " + addr.String()

	for _, s := range msg.Want {
		str += " poll origin " + s.key.PollOrigin
		str += " ID " + strconv.FormatUint(uint64(s.key.PollID), 10)
		str += " participants " + s.participantList.String() // TODO what we want?
	}
	fmt.Println(str)
	printPeers(gossiper)
}

func printFlippedCoin(addr *net.UDPAddr, typeOfFlip string) {
	fmt.Println("FLIPPED COIN sending", typeOfFlip, "to", addr.String())
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
