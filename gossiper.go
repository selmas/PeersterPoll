package pollparty

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	secrand "crypto/rand" // alias needed as we import two libraries with name "rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/dedis/protobuf"
	"log"
	"math/big"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ShareablePollInfo struct {
	Poll         *Poll
	Participants [][]*big.Int
	Commitments  []Commitment
	Votes        []Vote
}

type PollInfo struct {
	ShareablePollInfo
	Tags     map[[2]*big.Int]Commitment // mapping from tag to commitment to detect double voting
	Registry *crypto.PublicKey
}

type Server struct {
	Addr *net.UDPAddr
	Conn *net.UDPConn
}

type PeerSet struct {
	sync.RWMutex
	Set map[string]bool
}

type PollSet struct {
	sync.RWMutex
	m map[PollKeyMap]PollInfo
}

func (s *PollSet) Has(k PollKey) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.m[k.Pack()]
	return ok
}

func (s *PollSet) Get(k PollKey) PollInfo {
	s.RLock()
	defer s.RUnlock()

	return s.m[k.Pack()]
}

// initialize tags mapping
func (s *PollSet) Store(pkg PollPacket) {
	s.Lock()
	defer s.Unlock()

	info := s.m[pkg.ID.Pack()]

	if pkg.Poll != nil {
		poll := *pkg.Poll

		// TODO poll != *info.Poll -> bad rep

		info.Poll = &poll
	}

	if pkg.Commitment != nil {
		info.Commitments = append(info.Commitments, *pkg.Commitment)
	}

	if pkg.Vote != nil {
		info.Votes = append(info.Votes, *pkg.Vote)
	}

	s.m[pkg.ID.Pack()] = info
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

func (s RunningPollWriter) Send(pkg PollPacket) {

	if pkg.Poll != nil {
		poll := *pkg.Poll
		if poll.IsTooLate() {
			log.Println("poll came in too late")
		} else {
			s.Poll <- poll
		}
	}

	if pkg.Commitment != nil {
		s.Commitments <- *pkg.Commitment
	}

	if pkg.PollCommitments != nil {
		s.PollCommitments <- *pkg.PollCommitments
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
	m map[PollKeyMap]RunningPollWriter
}

func (s *RunningPollSet) Has(k PollKey) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.m[k.Pack()]
	return ok
}

func (s *RunningPollSet) Get(k PollKey) RunningPollWriter {
	assert(s.Has(k))

	s.RLock()
	defer s.RUnlock()

	return s.m[k.Pack()]
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
	s.m[k.Pack()] = w
	s.Unlock()

	go handler(k, r)
}

func (s *RunningPollSet) Send(pkg PollPacket) {
	s.RLock()
	defer s.RUnlock()

	r := s.m[pkg.ID.Pack()]
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

func Curve() elliptic.Curve {
	return elliptic.P256()
}

type Gossiper struct {
	sync.RWMutex // TODO use everywhere (for id change also)
	Name         string
	LastID       uint64
	KeyPair      ecdsa.PrivateKey
	Peers        PeerSet
	RunningPolls RunningPollSet
	Polls        PollSet
	Server       Server
	ValidKeys    []ecdsa.PublicKey
	Reputations  ReputationInfo
}

func (g *Gossiper) addPeer(addr net.UDPAddr) {
	g.Peers.Lock()
	defer g.Peers.Unlock()

	g.Peers.Set[addr.String()] = true
}

func NewServer(address string) Server {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}

	return Server{
		Addr: addr,
		Conn: conn,
	}
}

func NewGossiper(name string, server Server) (*Gossiper, error) {
	keyPair, err := PrivateKeyLoad(PrivateKeyFileName(name))
	if err != nil {
		return nil, errors.New("Elliptic Curve Generation: " + err.Error())
	}

	validKeys, err := KeyFileLoad()
	if err != nil {
		return nil, errors.New("NewGossiper: " + err.Error())
	}

	return &Gossiper{
		Name:    name,
		KeyPair: keyPair,
		Server:  server,
		Peers: PeerSet{
			Set: make(map[string]bool),
		},
		RunningPolls: RunningPollSet{
			m: make(map[PollKeyMap]RunningPollWriter),
		},
		Polls: PollSet{
			m: make(map[PollKeyMap]PollInfo),
		},
		ValidKeys: validKeys,
		Reputations: NewReputationInfo(),
	}, nil
}

func NewPollKey(g *Gossiper) PollKey {
	g.Lock()
	defer g.Unlock()

	return PollKey{
		ID:     atomic.AddUint64(&g.LastID, 1),
		Origin: g.KeyPair.PublicKey,
	}
}

type Dispatcher func(net.UDPAddr, GossipPacket)

func RunServer(gossiper *Gossiper, server Server, dispatcher Dispatcher) {
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

		pkg := msg.ToBase()
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

func writeMsgToUDP(server Server, peer *net.UDPAddr, poll *PollPacket, status *StatusPacket, signature *Signature,
	reputation *ReputationPacket) {
	msg := GossipPacket{
		Poll:       poll,
		Signature:  signature,
		Status:     status,
		Reputation: reputation,
	}.ToWire()
	err := msg.Check()
	if err != nil {
		panic(err)
	}

	toSend, err := protobuf.Encode(&msg)
	if err != nil {
		panic(err)
	}

	server.Conn.WriteToUDP(toSend, peer)
}

func (g *Gossiper) SendPoll(id PollKey, msg Poll) {
	pkg := PollPacket{
		ID:   id,
		Poll: &msg,
	}
	sig, err := ecSignature(g, pkg)
	if err != nil {
		panic(err)
	}
	g.SendPollPacket(&pkg, &sig, nil)
}

func (g *Gossiper) SendCommitment(id PollKey, msg Commitment, participants [][]*big.Int, tmpKey *ecdsa.PrivateKey, pos int) {
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
	g.SendPollPacket(&pkg, &Signature{&lrs, nil}, nil)
}

func (g *Gossiper) SendPollCommitments(id PollKey, msg PollCommitments) {
	pkg := PollPacket{
		ID:              id,
		PollCommitments: &msg,
	}

	sig, err := ecSignature(g, pkg)
	if err != nil {
		return
	}

	g.SendPollPacket(&pkg, &sig, nil)
}

func ecSignature(g *Gossiper, poll PollPacket) (Signature, error) {
	input, err := json.Marshal(poll)
	if err != nil {
		log.Printf("unable to encode as json")
		return Signature{}, err
	}

	hash := sha256.New()
	_, err = hash.Write(input)
	r, s, err := ecdsa.Sign(secrand.Reader, &g.KeyPair, hash.Sum(nil))
	if err != nil {
		log.Printf("error generating elliptic curve signature")
		return Signature{}, err
	}
	return Signature{nil, &EllipticCurveSignature{*r, *s}}, nil
}

func (g *Gossiper) SendVote(id PollKey, vote Vote, participants [][]*big.Int, tmpKey *ecdsa.PrivateKey, pos int) {
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
	g.SendPollPacket(&pkg, &Signature{&lrs, nil}, nil)
}

func (g *Gossiper) SendPollPacket(msg *PollPacket, sig *Signature, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&g.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(g.Server, peer, msg, nil, sig, nil)

		printFlippedCoin(peer, "poll")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func getStatus(g *Gossiper) StatusPacket {
	g.Polls.Lock()
	defer g.Polls.Unlock()

	infos := make(map[PollKeyMap]ShareablePollInfo)

	for id, info := range g.Polls.m {
		infos[id] = info.ShareablePollInfo
	}

	return StatusPacket{
		Infos: infos,
	}
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

			if !g.SignatureValid(pkg) {
				log.Println("invalid signature found but not handled")
				// TODO suspect peer
				return
			}

			if pkg.Signature.Linkable != nil {
				if doubleVoted(g, pkg) {
					log.Println("double vote but not handled")
					// TODO suspect peer
					return
				}
				g.Polls.m[pkg.Poll.ID.Pack()].Tags[pkg.Signature.Linkable.Tag] = *pkg.Poll.Commitment
			}

			poll.Print(fromPeer)

			g.Polls.Store(poll)

			if !g.RunningPolls.Has(poll.ID) {
				g.RunningPolls.Add(poll.ID, VoterHandler(g))
			}

			assert(g.RunningPolls.Has(poll.ID))
			g.RunningPolls.Send(poll)
		}

		if pkg.Status != nil {
			status := *pkg.Status
			syncStatus(g, fromPeer, status)
		}

		if pkg.Reputation != nil {

			// TODO check if packet is new

			if !repSignatureValid(g, pkg) {
				g.Reputations.Suspect(fromPeer.String())
			}

			pollID := pkg.Reputation.PollID

			// store Reputation in receivedOpinions[poll]
			g.Reputations.AddPeerOpinion(pkg.Reputation, pollID)

			/* TODO if recvRep.len == #peers || timeout{
				g.Reputations.AddReputations(pollID)
				g.Reputations.AddTablesWait[pollID] <- true
			}*/

			//TODO gossip packet
		}
	}
}

func doubleVoted(g *Gossiper, pkg GossipPacket) bool {
	tag := pkg.Signature.Linkable.Tag
	commit, stored := g.Polls.m[pkg.Poll.ID.Pack()].Tags[tag]

	if stored {
		return !(string(commit.Hash[:]) == string(pkg.Poll.Commitment.Hash[:]))
	}

	return false
}

func (g *Gossiper) SignatureValid(pkg GossipPacket) bool {
	poll := pkg.Poll

	if (poll.Commitment != nil || poll.Vote != nil) && !(poll.Commitment != nil && poll.Vote != nil) {
		return pkg.Signature.Linkable != nil && verifySig(*pkg.Signature.Linkable, g.Polls.m[pkg.Poll.ID.Pack()].Participants)
	}

	if poll.PollCommitments != nil || poll.Poll != nil {
		input, err := json.Marshal(poll)
		if err != nil {
			log.Printf("unable to encode as json")
		}

		hash := sha256.New()
		_, err = hash.Write(input)
		if err != nil {
			log.Printf("error generating elliptic curve signature")
		}

		return pkg.Signature.Elliptic != nil && ecdsa.Verify(&pkg.Poll.ID.Origin, hash.Sum(nil),
			&pkg.Signature.Elliptic.R, &pkg.Signature.Elliptic.S)
	}

	return false
}

func parseAddr(str string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", str)
	if err != nil {
		panic(err)
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

		//printFlippedCoin(peer, "status")
		status := getStatus(gossiper)
		writeMsgToUDP(gossiper.Server, peer, nil, &status, nil, nil)
	}
}
