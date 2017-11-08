package main

import (
	"log"
	"flag"
	"net"
	"strconv"

	"github.com/dedis/protobuf"

	"github.com/JohnDoe/Peerster/part1/proto"
)

func main() {
	port := flag.Uint64("UIPort", 10000, "port to connect to")
	msgStr := flag.String("msg", "Hello", "message to send")
	flag.Parse()

	msg := proto.Message{
		OPName:	nil,
		Relay:	nil,
		Text:	*msgStr,
	}
	to_send, err := protobuf.Encode(&msg)
	if err != nil {
		log.Fatal(err)
	}

	server_addr, err := net.ResolveUDPAddr("udp", ":" + strconv.FormatUint(*port, 10))
	if err != nil {
		log.Fatal(err)
	}

	client, err := net.DialUDP("udp", nil, server_addr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sent_size, err := client.Write(to_send)
	if err != nil {
		log.Fatal(err)
	}
	if sent_size != len(to_send) {
		log.Fatal("unable to send the whole message")
	}
}
