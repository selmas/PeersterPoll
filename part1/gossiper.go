package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/dedis/protobuf"

	"github.com/JohnDoe/Peerster/part1/proto"
)

type Server struct {
	Addr	*net.UDPAddr
	Conn	*net.UDPConn
}

type Gossiper struct {
	Name	string
	Peers   PeerSet
	Server	*Server
}

type PeerSet struct {
	sync.RWMutex
	Set	map[string]bool
}

func NewServer(address string) *Server {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		Addr: addr,
		Conn: conn,
	}
}

func NewGossiper(name string, server *Server) *Gossiper {
	return &Gossiper{
		Name:	name,
		Peers:	PeerSet{
			Set:	make(map[string]bool),
		},
		Server:	server,
	}
}

func formatStringPtr(ptr *string) string {
	if ptr == nil {
		return "N/A"
	}
	return *ptr
}

func printMessage(gossiper *Gossiper, msg proto.Message) {
	fmt.Println(msg.Text, formatStringPtr(msg.OPName), formatStringPtr(msg.Relay))

	firstPrint := true
	gossiper.Peers.RLock()
	for peer, _ := range gossiper.Peers.Set {
		if firstPrint {
			fmt.Print(peer)
			firstPrint = false
			continue
		}

		fmt.Print(",", peer)
	}
	gossiper.Peers.RUnlock()
	fmt.Println()
}

func handleMessage(api *Server, gossiper *Gossiper, userFacing bool, clientAddr *net.UDPAddr, buf []byte) {
	var msg proto.Message

	err := protobuf.Decode(buf, &msg)
	if err != nil {
		// XXX unable to decode still print?
		log.Printf("unable to decode msg: ", err)
		return
	}

	if !userFacing {
		gossiper.Peers.Lock()
		gossiper.Peers.Set[clientAddr.String()] = true
		gossiper.Peers.Unlock()
	}

	printMessage(gossiper, msg)

	var msgToSend proto.Message
	var relay string = gossiper.Server.Addr.String()
	if userFacing {
		msgToSend = proto.Message{
			OPName:	&gossiper.Name,
			Relay:	&relay,
			Text:	msg.Text,
		}
	} else {
		msgToSend = proto.Message{
			OPName:	msg.OPName,
			Relay:	&relay,
			Text:	msg.Text,
		}
	}

	toSend, err := protobuf.Encode(&msgToSend)
	if err != nil {
		log.Printf("unable to encode answer: ", err)
		return
	}

	gossiper.Peers.RLock()
	if userFacing {
		for peer, _ := range gossiper.Peers.Set {
			gossiper.Server.Conn.WriteToUDP(toSend, parseAddr(peer))
		}
	} else {
		for peer, _ := range gossiper.Peers.Set {
			if (peer == gossiper.Server.Addr.String() || peer == clientAddr.String()) {
				continue
			}
			gossiper.Server.Conn.WriteToUDP(toSend, parseAddr(peer))
		}
	}
	gossiper.Peers.RUnlock()
}

func runServer(gossiper *Gossiper, api *Server, serverToRead *Server, userFacing bool) {
	buf := make([]byte, 64)

	for {
		bufSize, clientAddr, err := serverToRead.Conn.ReadFromUDP(buf)

		if err != nil {
			log.Printf("dropped connection: ", err)
			continue
		}

		handleMessage(api, gossiper, userFacing, clientAddr, buf[0:bufSize])
	}
}

func parseAddr(str string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", str)
	if err != nil {
		log.Fatal("invalide peer \"", str, "\": ", err)
	}

	return addr
}

func main() {
	apiPort := flag.String("UIPort", "10000", "port for the client to connect")
	gossipPort := flag.String("gossipPort", "127.0.0.1:5000", "port to connect the gossiper server")
	name := flag.String("name", "nodeA", "server identifier")
	peersStr := flag.String("peers", "127.0.0.1:5001,10.1.1.7:5002", "underscore separated list of peers")
	flag.Parse()

	gossiper := NewGossiper(*name, NewServer(*gossipPort))
	defer gossiper.Server.Conn.Close()

	for _, peer := range strings.Split(*peersStr, ",") {
		gossiper.Peers.Set[peer] = true
	}

	api := NewServer(":" + *apiPort)
	defer api.Conn.Close()

	go runServer(gossiper, api, gossiper.Server, false)
	runServer(gossiper, api, api, true)
}
