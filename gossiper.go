package pollparty

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"github.com/dedis/protobuf"

	crypto "crypto/rand" // alias needed as we import two libraries with name "rand"
	"math/big"
	"encoding/json"
	"crypto/sha256"
	"hash"
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
	Poll            *Poll
	Commitments     []Commitment
	PollCommitments *PollCommitments
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

func (s *PollSet) Store(pkg PollPacket) {
	s.Lock()
	defer s.Unlock()

	info := s.m[pkg.ID]

	if pkg.Poll != nil {
		poll := *pkg.Poll

		// TODO poll != *info.Poll -> bad rep

		info.Poll = &poll
	}

	if pkg.Commitment != nil {
		info.Commitments = append(info.Commitments, *pkg.Commitment)
	}

	if pkg.PollCommitments != nil {
		commits := *pkg.PollCommitments

		// TODO commits != *info.PollCommitments -> bad rep

		info.PollCommitments = &commits
	}

	if pkg.Vote != nil {
		info.Votes = append(info.Votes, *pkg.Vote)
	}

	s.m[pkg.ID] = info
}

// TODO maybe split in two to have voter/server separation
type RunningPollReader struct {
	Poll            <-chan Poll
	LocalVote       <-chan string
	Commitments     <-chan Commitment
	PollCommitments <-chan PollCommitments
	Votes           <-chan Vote
}

type RunningPollWriter struct {
	Poll            chan<- Poll
	LocalVote       chan<- string
	Commitments     chan<- Commitment
	PollCommitments chan<- PollCommitments
	Votes           chan<- Vote
}

var curve elliptic.Curve

func (s RunningPollWriter) Send(pkg PollPacket) {

	if pkg.Poll != nil {
		poll := *pkg.Poll
		if poll.IsTooLate() {
			log.Println("poll came in too late")
		} else {
			println(">>", pkg.ID.String())
			s.Poll <- poll
		}
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
	sync.RWMutex // TODO use everywhere (for id change also)
	Name         string
	LastID       uint64
	KeyPair      *ecdsa.PrivateKey
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

func NewPollKey(g *Gossiper) PollKey {
	g.Lock()
	defer g.Unlock()

	return PollKey{
		ID:     atomic.AddUint64(&g.LastID, 1),
		Origin: g.Name,
	}
}

type Dispatcher func(net.UDPAddr, GossipPacket)

func RunServer(gossiper *Gossiper, server *Server, dispatcher Dispatcher) {
	buf := make([]byte, 1024)

	for {
		bufSize, peerAddr, err := server.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("dropped connection:", err)
			continue
		}

		var msg GossipPacketWire
		err = protobuf.Decode(buf[:bufSize], &msg)
		if err != nil {
			log.Println("unable to decode msg:", err)
			continue
		}

		err = msg.Check()
		if err != nil {
			log.Println("invalid GossipPacketWire received:", err)
			return
		}

		pkg, err := msg.ToBase()
		if err != nil {
			log.Println("badly formatted GossipPacketWire received:", err)
			return
		}

		go dispatcher(*peerAddr, pkg)
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

func writeMsgToUDP(server *Server, peer *net.UDPAddr, poll *PollPacket, status *StatusPacket, signature *Signature) {
	msg := GossipPacket{
		Poll:      poll,
		Signature: signature,
		Status:    status,
	}.ToWire()
	toSend, err := protobuf.Encode(&msg)

	if err != nil {
		log.Printf("unable to encode answer: %s", err)
		return
	}

	server.Conn.WriteToUDP(toSend, peer)
}

func (g *Gossiper) SendPoll(id PollKey, msg Poll, gossiper Gossiper) {
	pkg := PollPacket{
		ID:   id,
		Poll: &msg,
	}
	sig, err := ecSignPollPacket(pkg, gossiper)
	if err != nil {
		return
	}
	g.SendPollPacket(&pkg, &sig, nil)
}

func (g *Gossiper) SendCommitment(id PollKey, msg Commitment, participants [][]*big.Int, tmpKey *ecdsa.PrivateKey,	pos int) {
	pkg := PollPacket{
		ID:         id,
		Commitment: &msg,
	}
	input, err := json.Marshal(pkg)
	if err != nil {
		log.Printf("unable to encode as json")
		return
	}

	lrs := linkableRingSignature(input, participants, tmpKey, pos)
	g.SendPollPacket(&pkg, &Signature{&lrs, nil},nil)
}

func (g *Gossiper) SendPollCommitments(id PollKey, msg PollCommitments, gossiper Gossiper) {
	pkg := PollPacket{
		ID:              id,
		PollCommitments: &msg,
	}

	ecsig, err := ecSignPollPacket(pkg, gossiper)
	if err != nil {
		return
	}

	g.SendPollPacket(&pkg, &ecsig,nil)
}

func ecSignPollPacket(pkg PollPacket, gossiper Gossiper) (Signature, error) {
	input, err := json.Marshal(pkg)
	if err != nil {
		log.Printf("unable to encode as json")
		return Signature{}, err
	}

	hash := sha256.New()
	_, err = hash.Write(input)
	r, s, err := ecdsa.Sign(crypto.Reader, gossiper.KeyPair, hash.Sum(nil)) // TODO use masterkey for signing
	if err != nil {
		log.Printf("error generating elliptic curve signature")
		return Signature{}, err
	}
	return Signature{nil, &EllipticCurveSignature{r,s}}, nil
}

func (g *Gossiper) SendVote(id PollKey, vote Vote, participants [][]*big.Int, tmpKey *ecdsa.PrivateKey,	pos int) {
	pkg := PollPacket{
		ID:   id,
		Vote: &vote,
	}

	input, err := json.Marshal(pkg)
	if err != nil {
		log.Printf("unable to encode as json")
		return
	}

	lrs := linkableRingSignature(input, participants, tmpKey, pos)
	g.SendPollPacket(&pkg, &Signature{&lrs,nil}, nil)
}

func (g *Gossiper) SendPollPacket(msg *PollPacket, sig *Signature, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&g.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(g.Server, peer, msg, nil, sig)

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
func syncStatus(gossiper *Gossiper, peer net.UDPAddr, s StatusPacket) {
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

func DispatcherPeersterMessage(g *Gossiper) Dispatcher {
	return func(fromPeer net.UDPAddr, pkg GossipPacket) {
		g.addPeer(fromPeer)

		if pkg.Poll != nil {
			poll := *pkg.Poll
			poll.Print(fromPeer)

			g.Polls.Store(poll)

			if !g.RunningPolls.Has(poll.ID) {
				g.RunningPolls.Add(poll.ID, VoterHandler(g))
			}

			assert(g.RunningPolls.Has(poll.ID))
			g.RunningPolls.Send(poll.ID, poll)
		}

		if pkg.Status != nil {
			syncStatus(g, fromPeer, *pkg.Status)
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

func AntiEntropyGossip(gossiper *Gossiper) {
	ticker := time.NewTicker(time.Second)

	for {
		_ = <-ticker.C

		peer := getRandomPeer(&gossiper.Peers, nil)
		if peer == nil {
			continue
		}

		printFlippedCoin(peer, "status")
		status := getStatus(gossiper)
		writeMsgToUDP(gossiper.Server, peer, nil, &status, nil)
	}
}
