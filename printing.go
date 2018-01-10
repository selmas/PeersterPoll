package pollparty

import (
	"fmt"
	"net"
	"strings"
)

func (pkg PollPacket) Print(clientAddr net.UDPAddr) {
	msg := fmt.Sprintf(
		"Poll %s from %s",
		pkg.ID.String(), clientAddr.String(),
	)

	if pkg.Poll != nil {
		msg = fmt.Sprintf(
			"%s asking \"%s\" with options [\"%s\"]",
			msg, pkg.Poll.Question, strings.Join(pkg.Poll.Options, "\",\""),
		)
	}

	fmt.Println(msg)
}

func (msg StatusPacket) Print(clientAddr net.UDPAddr) {
	var str string
	str += "STATUS from " + clientAddr.String()
	if msg.ReputationPkts != nil {
		str += ", sync REPUTATION table"
	}
	if msg.PollPkts != nil {
		str += ", sync POLLPACKET"
	}
	fmt.Println(str)
}


func printFlippedCoin(addr *net.UDPAddr, typeOfFlip string) {
	fmt.Println("FLIPPED COIN sending", typeOfFlip, "to", addr.String())
}
