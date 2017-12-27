package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dedis/protobuf"
)

type Server struct {
	Addr *net.UDPAddr
	Conn *net.UDPConn
}

type PeerSet struct {
	sync.RWMutex
	Set map[string]bool
}

type PollKey struct {
	PollOrigin string
	PollID     uint32
}

type PollSet struct {
	sync.RWMutex
	Set map[PollKey]*PollPacket
}

type MessageSet struct {
	sync.RWMutex
	Set map[string][]MessageID
}

type Route struct {
	IsDirect bool
	Addr     net.UDPAddr
}

type RoutingTable struct {
	sync.RWMutex
	Table map[string]Route
}

type Gossiper struct {
	Name     string
	LastUid  uint32
	Peers    PeerSet
	Messages MessageSet
	Polls    PollSet
	Server   *Server
}

func (g *Gossiper) addPeer(addr net.UDPAddr) {
	g.Peers.Lock()
	defer g.Peers.Unlock()

	g.Peers.Set[addr.String()] = true
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
		Name:   name,
		Server: server,
		Peers: PeerSet{
			Set: make(map[string]bool),
		},
		Messages: MessageSet{
			Set: make(map[string][]MessageID),
		},
		Polls: PollSet{
			Set: make(map[PollKey]*PollPacket),
		},
	}
}

type Dispatcher func(*net.UDPAddr, *GossipPacket)

func runServer(gossiper *Gossiper, server *Server, dispatcher Dispatcher) {
	buf := make([]byte, 1024)

	for {
		bufSize, peerAddr, err := server.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("dropped connection:", err)
			continue
		}

		var msg GossipPacket
		err = protobuf.Decode(buf[:bufSize], &msg)
		if err != nil {
			log.Println("unable to decode msg:", err)
			continue
		}

		go dispatcher(peerAddr, &msg)
	}
}

func getRandomPeer(peers *PeerSet, butNotThisPeer *net.UDPAddr) *net.UDPAddr {
	peers.RLock()
	defer peers.RUnlock()

	var addr *net.UDPAddr = nil
	for peer := range peers.Set {
		if peer == butNotThisPeer.String() {
			continue
		}

		addr = parseAddr(peer)
	}

	return addr
}

func writeMsgToUDP(server *Server, peer *net.UDPAddr, rumor *RumorMessage, status *StatusPacket, poll *PollPacket) {
	toSend, err := protobuf.Encode(&GossipPacket{
		Rumor:  rumor,
		Status: status,
	})

	if err != nil {
		log.Printf("unable to encode answer: %s", err)
		return
	}

	server.Conn.WriteToUDP(toSend, peer)
}

func sendRumor(gossiper *Gossiper, msg *RumorMessage, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&gossiper.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(gossiper.Server, peer, msg, nil, nil)

		printFlippedCoin(peer, "rumor")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func peerIsAheadOfUs(gossiper *Gossiper, s *StatusPacket) bool {
	gossiper.Messages.RLock()
	defer gossiper.Messages.RUnlock()

	for _, status := range s.Want {
		msgs, found := gossiper.Messages.Set[status.Identifier]

		if !found {
			return true
		}

		if uint32(len(msgs)) < (status.NextID - 1) {
			return true
		}
	}

	return false
}

func getWantedRumor(gossiper *Gossiper, s *StatusPacket) *RumorMessage {
	gossiper.Messages.RLock()
	defer gossiper.Messages.RUnlock()

	for origin, msgs := range gossiper.Messages.Set {
		var status *PeerStatus = nil
		for _, current := range s.Want {
			if current.Identifier == origin {
				status = &current
			}
		}

		if status == nil || uint32(len(msgs)) >= status.NextID {
			var pos uint
			if status == nil {
				pos = 0
			} else {
				pos = uint(len(msgs)) - 1
			}

			msg := msgs[pos]
			return &RumorMessage{
				Peer: msg,
			}
		}
	}

	return nil
}

func getStatus(gossiper *Gossiper) *StatusPacket {
	gossiper.Messages.Lock()
	defer gossiper.Messages.Unlock()

	wanted := make([]PeerStatus, len(gossiper.Messages.Set))
	i := 0
	for origin, msgs := range gossiper.Messages.Set {
		wanted[i] = PeerStatus{
			Identifier: origin,
			NextID:     uint32(len(msgs)) + 1,
		}

		i++
	}

	return &StatusPacket{
		Want: wanted,
	}
}

func syncStatus(gossiper *Gossiper, peer *net.UDPAddr, msg *StatusPacket) {
	rumor := getWantedRumor(gossiper, msg)

	if rumor != nil {
		writeMsgToUDP(gossiper.Server, peer, rumor, nil, nil)
	} else if peerIsAheadOfUs(gossiper, msg) {
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper), nil)
	} else {
		printInSyncWith(peer)
	}
}

func storeRumor(gossiper *Gossiper, rumor *RumorMessage) bool {
	added := false

	gossiper.Messages.Lock()
	defer gossiper.Messages.Unlock()

	msgs := gossiper.Messages.Set[rumor.Peer.Origin]

	if rumor.Peer.ID == uint32(len(msgs))+1 {
		msgs = append(msgs, rumor.Peer)
		added = true
		gossiper.Messages.Set[rumor.Peer.Origin] = msgs
	}

	return added
}

func dispatcherPeersterMessage(gossiper *Gossiper) Dispatcher {
	return func(fromPeer *net.UDPAddr, pkg *GossipPacket) {
		err := CheckGossipPacket(pkg)
		if err != nil {
			log.Fatal("invalid GossipPacket received:", err)
			return
		}

		gossiper.addPeer(*fromPeer)

		if pkg.Rumor != nil {
			r := pkg.Rumor
			printRumor(gossiper, fromPeer, r)

			storeRumor(gossiper, r)
			sendRumor(gossiper, r, fromPeer)

			poll := r.Poll
			if poll.Question.StartTime.Add(poll.Question.Duration).Before(time.Now()) {
				handlePoll(gossiper, poll)
			} else {
				forwardPoll(gossiper, poll)
			}
		}

		if pkg.Status != nil {
			printStatus(gossiper, fromPeer, pkg.Status)
			syncStatus(gossiper, fromPeer, pkg.Status)
		}

	}
}

func parseAddr(str string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", str)
	if err != nil {
		log.Fatal("invalid peer \"", str, "\": ", err)
	}

	return addr
}

func antiEntropyGossip(gossiper *Gossiper) {
	ticker := time.NewTicker(time.Second)

	for {
		_ = <-ticker.C

		peer := getRandomPeer(&gossiper.Peers, nil)
		if peer == nil {
			continue
		}

		printFlippedCoin(peer, "status")
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper), nil)
	}
}

func main() {
	uiPort := flag.String("UIPort", "10000", "port for the client to connect")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port to connect the gossiper server")
	name := flag.String("name", "nodeA", "server identifier")
	peersStr := flag.String("peers", "127.0.0.1:5001_10.1.1.7:5002", "underscore separated list of peers")
	flag.Parse()

	gossiper := NewGossiper(*name, NewServer(*gossipAddr))
	defer gossiper.Server.Conn.Close()

	for _, peer := range strings.Split(*peersStr, "_") {
		if peer == "" {
			continue
		}

		gossiper.Peers.Set[peer] = true
	}

	// one should stay main thread'ed to avoid exiting
	go runServer(gossiper, gossiper.Server, dispatcherPeersterMessage(gossiper))
	go antiEntropyGossip(gossiper)
	apiStart(gossiper, *uiPort)
}
