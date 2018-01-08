package pollparty

import (
	"fmt"
	"net"
)

func (pkg PollPacket) Print(clientAddr net.UDPAddr) {
	fmt.Println(
		"Poll ", pkg.ID.String(),
		"from", clientAddr.String(),
		// TODO "Question", pkg.pollQuestion.Question,
		//"with options", strings.Join(msg.pollQuestion.VoteOptions, ","),
	)
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
}

func printFlippedCoin(addr *net.UDPAddr, typeOfFlip string) {
	fmt.Println("FLIPPED COIN sending", typeOfFlip, "to", addr.String())
}
