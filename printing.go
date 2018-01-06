package main

import (
	"fmt"
	"net"
)

func printPollPacket(gossiper *Gossiper, clientAddr *net.UDPAddr, pkg PollPacket) {
	fmt.Println(
		"Poll ", pkg.ID.String(),
		"from", clientAddr.String(),
		// TODO "Question", pkg.pollQuestion.Question,
		//"with options", strings.Join(msg.pollQuestion.VoteOptions, ","),
	)
	printPeers(gossiper)
}

func printStatus(gossiper *Gossiper, addr *net.UDPAddr, msg *StatusPacket) {
	var str string
	str += "STATUS from " + addr.String()

	/* TODO for _, s := range msg.Want {
		str += " poll origin " + s.key.PollOrigin
		str += " ID " + strconv.FormatUint(uint64(s.key.PollID), 10)
		// todo: add meaningful output for status
	}*/
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
