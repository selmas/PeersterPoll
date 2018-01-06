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

type PollInfo struct {
	Poll            Poll
	Commitments     []Commitment
	PollCommitments []Commitment
	Votes           []Vote
}

type PollSet struct {
	sync.RWMutex
	m map[PollKey]PollInfo
}

func (s *PollSet) Has(k PollKey) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.m[k]
	return ok
}

func (s *PollSet) Get(k PollKey) PollInfo {
	s.RLock()
	defer s.RUnlock()

	return s.m[k]
}

type RunningPollReader struct {
	// TODO maybe split in two to have voter/server separation
	Poll            <-chan Poll
	Commitments     <-chan Commitment
	PollCommitments <-chan PollCommitments
	Votes           <-chan Vote
}

type RunningPollWriter struct {
	Poll            chan<- Poll
	Commitments     chan<- Commitment
	PollCommitments chan<- PollCommitments
	Votes           chan<- Vote
}

func (s RunningPollWriter) Send(pkg PollPacket) {

	if pkg.Poll != nil {
		s.Poll <- *pkg.Poll
		close(s.Poll)
	}

	if pkg.Commitment != nil {
		s.Commitments <- *pkg.Commitment
	}

	if pkg.PollCommitments != nil {
		s.PollCommitments <- *pkg.PollCommitments
		close(s.PollCommitments)
	}

	if pkg.PollCommitments != nil {
		s.PollCommitments <- *pkg.PollCommitments
	}

	if pkg.Vote != nil {
		s.Votes <- *pkg.Vote
	}
}

type RunningPollSet struct {
	sync.RWMutex
	m map[PollKey]RunningPollWriter
}

func (s *RunningPollSet) Has(k PollKey) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.m[k]
	return ok
}

func (s *RunningPollSet) Get(k PollKey) RunningPollWriter {
	assert(s.Has(k))

	s.RLock()
	defer s.RUnlock()

	return s.m[k]
}

func (s *RunningPollSet) Add(k PollKey, handler PoolPacketHandler) {
	assert(!s.Has(k))

	poll := make(chan Poll)
	commitments := make(chan Commitment)
	pollCommitments := make(chan PollCommitments)
	votes := make(chan Vote)

	r := RunningPollReader{
		Poll:            poll,
		Commitments:     commitments,
		PollCommitments: pollCommitments,
		Votes:           votes,
	}

	w := RunningPollWriter{
		Poll:            poll,
		Commitments:     commitments,
		PollCommitments: pollCommitments,
		Votes:           votes,
	}

	s.Lock()
	s.m[k] = w
	s.Unlock()

	go handler(k, r)
}

func (s *RunningPollSet) Send(k PollKey, pkg PollPacket) {
	s.RLock()
	defer s.RUnlock()

	r := s.m[k]
	r.Send(pkg)
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
	Name         string
	KeyPair      *ecdsa.PrivateKey
	lastPollID   uint32
	Peers        PeerSet
	RunningPolls RunningPollSet
	Polls        PollSet
	Server       *Server
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

func NewGossiper(name string, server *Server) (*Gossiper, error) {
	curve = elliptic.P256()
	// Reader is a global, shared instance of a cryptographically strong pseudo-random generator.
	keyPair, err := ecdsa.GenerateKey(curve, crypto.Reader) // generates key pair

	if err != nil {
		return nil, errors.New("Elliptic Curve Generation: " + err.Error())
	}

	return &Gossiper{
		Name:    name,
		KeyPair: keyPair,
		Server:  server,
		Peers: PeerSet{
			Set: make(map[string]bool),
		},
		RunningPolls: RunningPollSet{
			m: make(map[PollKey]RunningPollWriter),
		},
		Polls: PollSet{
			m: make(map[PollKey]PollInfo),
		},
	}, nil
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

func writeMsgToUDP(server *Server, peer *net.UDPAddr, poll *PollPacket, status *StatusPacket) {
	toSend, err := protobuf.Encode(&GossipPacket{
		Poll:   poll,
		Status: status,
	})

	if err != nil {
		log.Printf("unable to encode answer: %s", err)
		return
	}

	server.Conn.WriteToUDP(toSend, peer)
}

func (g *Gossiper) SendCommitment(id PollKey, msg Commitment) {
	pkg := PollPacket{
		ID:         id,
		Commitment: &msg,
	}

	g.SendPollPacket(&pkg, nil)
}

func (g *Gossiper) SendVote(id PollKey, option string) {
	vote := Vote{option}
	pkg := PollPacket{
		ID:   id,
		Vote: &vote,
	}

	g.SendPollPacket(&pkg, nil)
}

func (g *Gossiper) SendPollPacket(msg *PollPacket, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&g.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(g.Server, peer, msg, nil)

		printFlippedCoin(peer, "rumor")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func getStatus(gossiper *Gossiper) StatusPacket {
	gossiper.Polls.Lock()
	defer gossiper.Polls.Unlock()

	println("empty status") // TODO

	return StatusPacket{}
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

func dispatcherPeersterMessage(g *Gossiper) Dispatcher {
	return func(fromPeer *net.UDPAddr, pkg *GossipPacket) {
		err := pkg.Check()
		if err != nil {
			log.Println("invalid GossipPacket received:", err)
			return
		}

		g.addPeer(*fromPeer)

		if pkg.Poll != nil {
			poll := *pkg.Poll
			printPollPacket(g, fromPeer, poll)

			if g.Polls.Has(poll.ID) {
				log.Println("dropping msg related to finished poll:", poll.ID)
				return
			}

			if !g.RunningPolls.Has(poll.ID) {
				g.RunningPolls.Add(poll.ID, VoterHandler(g))
			}

			assert(g.RunningPolls.Has(poll.ID))
			g.RunningPolls.Send(poll.ID, poll)
		}

		if pkg.Status != nil {
			printStatus(g, fromPeer, pkg.Status)
			syncStatus(g, fromPeer, pkg.Status)
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
		status := getStatus(gossiper)
		writeMsgToUDP(gossiper.Server, peer, nil, &status)
	}
}

func main() {
	uiPort := flag.String("UIPort", "10000", "port for the client to connect")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port to connect the gossiper server")
	name := flag.String("name", "nodeA", "server identifier")
	peersStr := flag.String("peers", "127.0.0.1:5001_10.1.1.7:5002", "underscore separated list of peers")
	flag.Parse()

	gossiper, err := NewGossiper(*name, NewServer(*gossipAddr))
	if err != nil {
		log.Fatal(err)
	}
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
