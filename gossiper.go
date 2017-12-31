package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"github.com/dedis/protobuf"

	set "github.com/deckarep/golang-set"
	crypto "crypto/rand" // alias needed as we import two libraries with name "rand"
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
	m map[PollKey]*VoteSet
}

type VoteSet struct {
	poll 	*Poll
	votes	set.Set
}

type Route struct {
	IsDirect bool
	Addr     net.UDPAddr
}

// might be needed for consensus, do not remove
type RoutingTable struct {
	sync.RWMutex
	Table map[string]Route
}

type Gossiper struct {
	Name       string
	KeyPair    *ecdsa.PrivateKey
	lastPollID uint32
	Peers      PeerSet
	Polls      PollSet
	Server     *Server
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
	curve = elliptic.P256()
	// Reader is a global, shared instance of a cryptographically strong pseudo-random generator.
	keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader) // generates key pair

	if err != nil {
		errors.New("Elliptic Curve Generation: " + err.Error())
	}

	return &Gossiper{
		Name:    name,
		KeyPair: keyPair,
		Server:  server,
		Peers: PeerSet{
			Set: make(map[string]bool),
		},
		Polls: PollSet{
			m: make(map[PollKey]*VoteSet),
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

func writeMsgToUDP(server *Server, peer *net.UDPAddr, rumor *RumorMessage, status *StatusPacket) {
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

		writeMsgToUDP(gossiper.Server, peer, msg, nil)

		printFlippedCoin(peer, "rumor")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func getStatus(gossiper *Gossiper) *StatusPacket {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	wanted := make([]PollStatus, len(gossiper.Polls.m))
	i := 0
	for key, msg := range gossiper.Polls.m {
		wanted[i] = PollStatus{
			&key,
			msg.poll,
			msg.votes,
		}
		i++
	}

	return &StatusPacket{
		Want: wanted,
	}
}

// Todo: how to properly use mapset.Set ???
func syncStatus(gossiper *Gossiper, peer *net.UDPAddr, s *StatusPacket) {
	/*gossiper.Polls.RLock()
	defer gossiper.Polls.RUnlock()

	// store vector clocks as sets
	receivedPolls := set.NewSetFromSlice([]interface{}{
		s.Want,
	})
	storedPolls := set.NewSetFromSlice([]interface{}{
		getStatus(gossiper),
	})

	// Update my Poll storage
	for _, status := range s.Want {
		msg, found := gossiper.Polls.m[*status.key]

		// if I don't have it, add it
		if !found {
			gossiper.Polls.m[*status.key] = &VoteSet{status.poll, &status.votes}
			break
		}

		voteDiff := status.votes.Difference(msg.votes)
		if voteDiff.Cardinality() != 0 {
			gossiper.Polls.m[*status.key].votes = msg.votes.Union(status.votes)
		}

		participantDiff := status.poll.Participants.Difference(msg.poll.Participants)
		if participantDiff.Cardinality() != 0 {
			gossiper.Polls.m[*status.key].poll.Participants = msg.poll.Participants.Union(status.poll.Participants)
		}
	}

	// Compare vector clocks and send my updated vc if peer misses information
	if storedPolls.Difference(receivedPolls).Cardinality() != 0 { // Todo: test this difference !!
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper))
	}
	*/
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
			rumor := pkg.Rumor
			printRumor(gossiper, fromPeer, rumor)

			poll := rumor.pollQuestion
			if poll.StartTime.Add(poll.Duration).Before(time.Now()) {
				handleOpenPoll(gossiper, *rumor, fromPeer)
			} else {
				handleClosedPoll(gossiper, *rumor, fromPeer)
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
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper))
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
