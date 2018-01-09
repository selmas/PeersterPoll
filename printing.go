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
