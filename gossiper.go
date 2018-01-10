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
	"strings"
)

type ShareablePollInfo struct {
	Poll         Poll
	Participants [][2]big.Int
	Commitments  []Commitment
	Votes        []Vote
	Tags         map[[2]*big.Int][]Commitment // mapping from tag to []commitment to detect double voting
}

func (info ShareablePollInfo) Results() map[string]int {
	ret := make(map[string]int)

	for _, v := range info.Votes {
		ret[v.Option]++
	}

	return ret
}

type PollInfo struct {
	ShareablePollInfo
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

func (s *PollSet) Store(pkg PollPacket) bool {
	s.Lock()
	defer s.Unlock()

	added := false

	info := s.m[pkg.ID.Pack()]

	if pkg.Poll != nil {
		poll := *pkg.Poll

		// TODO poll != *info.Poll -> bad rep
		exist := false
		if info.Poll.Question == poll.Question &&
			info.Poll.StartTime.Equal(poll.StartTime) &&
			info.Poll.Duration.Minutes() == poll.Duration.Minutes() &&
			strings.Join(info.Poll.Options, ",") == strings.Join(poll.Options,",") {
				exist = true
		}

		if !exist{
			added = true
			info.Poll = poll
			info.Tags = make(map[[2]*big.Int][]Commitment)
		}
	}

	if pkg.Commitment != nil {
		exist := false
		for _,com := range info.Commitments{
			if string(com.Hash[:]) == string(pkg.Commitment.Hash[:]) {
				exist = true
			}
		}
		if !exist{
			info.Commitments = append(info.Commitments, *pkg.Commitment)
			added = true
		}
	}

	if pkg.Vote != nil {
		exist := false
		for _, vote := range info.Votes{
			if string(vote.Salt[:]) == string(pkg.Vote.Salt[:]) && vote.Option == pkg.Vote.Option {
				exist = true
			}
		}
		if !exist{
			info.Votes = append(info.Votes, *pkg.Vote)
			added = true
		}
	}

	s.m[pkg.ID.Pack()] = info

	return added
}

func (s *PollSet) Set(id PollKey, p PollInfo) {
	s.Lock()
	defer s.Unlock()

	s.m[id.Pack()] = p
}

type VoteAndSender struct {
	Vote   Vote
	Sender *net.UDPAddr
}

type RunningPollReader struct {
	Poll       <-chan Poll
	LocalVote  <-chan string
	VoteKey    <-chan VoteKey
	VoteKeys   <-chan VoteKeys
	Commitment <-chan Commitment
	Vote       <-chan VoteAndSender
}

type RunningPollWriter struct {
	Poll       chan<- Poll
	LocalVote  chan<- string
	VoteKey    chan<- VoteKey
	VoteKeys   chan<- VoteKeys
	Commitment chan<- Commitment
	Vote       chan<- VoteAndSender
}

func (s RunningPollWriter) Send(pkg PollPacket, fromPeer *net.UDPAddr) {
	if pkg.Poll != nil {
		poll := *pkg.Poll
		if poll.IsTooLate() {
			log.Println("poll came in too late")
		} else {
			s.Poll <- poll
		}
	}

	if pkg.VoteKey != nil {
		s.VoteKey <- *pkg.VoteKey
	}

	if pkg.VoteKeys != nil {
		s.VoteKeys <- *pkg.VoteKeys
	}

	if pkg.Commitment != nil {
		s.Commitment <- *pkg.Commitment
	}

	if pkg.Vote != nil {
		s.Vote <- VoteAndSender{
			Vote:   *pkg.Vote,
			Sender: fromPeer,
		}
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
	localVote := make(chan string)
	commitment := make(chan Commitment)
	voteKey := make(chan VoteKey)
	voteKeys := make(chan VoteKeys)
	vote := make(chan VoteAndSender)

	r := RunningPollReader{
		Poll:       poll,
		LocalVote:  localVote,
		VoteKey:    voteKey,
		VoteKeys:   voteKeys,
		Commitment: commitment,
		Vote:       vote,
	}

	w := RunningPollWriter{
		Poll:       poll,
		LocalVote:  localVote,
		VoteKey:    voteKey,
		VoteKeys:   voteKeys,
		Commitment: commitment,
		Vote:       vote,
	}

	s.Lock()
	s.m[k.Pack()] = w
	s.Unlock()

	key, err := ecdsa.GenerateKey(Curve(), secrand.Reader) // generates vote key
	if err != nil {
		panic(err)
	}

	go handler(k, *key, r)
}

func (s *RunningPollSet) Send(pkg PollPacket, fromPeer *net.UDPAddr) {
	s.RLock()
	defer s.RUnlock()

	r := s.m[pkg.ID.Pack()]
	r.Send(pkg, fromPeer)
}

type Route struct {
	IsDirect bool
	Addr     net.UDPAddr
}

func Curve() elliptic.Curve {
	return elliptic.P256()
}

type Gossiper struct {
	sync.RWMutex
	Name         string
	LastID       uint64
	KeyPair      ecdsa.PrivateKey
	Peers        PeerSet
	RunningPolls RunningPollSet
	Polls        PollSet
	Server       Server
	ValidKeys    [][2]big.Int
	Reputations  ReputationInfo
	Status       Status
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

type Status struct {
	sync.RWMutex
	PktStatus        map[SignatureMap]*PollPacket
	ReputationStatus map[SignatureMap]*ReputationPacket
}

func (s *Status) GetRep(k SignatureMap) *ReputationPacket {
	assert(s.HasRep(k))

	s.RLock()
	defer s.RUnlock()

	return s.ReputationStatus[k]
}

func (s *Status) HasRep(k SignatureMap) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.ReputationStatus[k]

	return ok
}

func (s *Status) SetRep(k SignatureMap, r *ReputationPacket) {
	s.Lock()
	defer s.Unlock()

	s.ReputationStatus[k] = r
}

func (s *Status) GetPkt(k SignatureMap) *PollPacket {
	s.RLock()
	defer s.RUnlock()

	return s.PktStatus[k]
}

func (s *Status) HasPkt(k SignatureMap) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.PktStatus[k]

	return ok
}

func (s *Status) SetPkt(k SignatureMap, p *PollPacket) {
	s.Lock()
	defer s.Unlock()

	s.PktStatus[k] = p
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
		ValidKeys:   validKeys,
		Reputations: NewReputationInfo(),
		Status: Status{
			PktStatus:        make(map[SignatureMap]*PollPacket),
			ReputationStatus: make(map[SignatureMap]*ReputationPacket),
		},
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
	buf := make([]byte, 16*1024)

	for {
		bufSize, peerAddr, err := server.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("dropped connection:", err)
			continue
		}

		if gossiper.Reputations.IsBlacklisted(peerAddr.String()) {
			log.Println("Received message from blacklisted peer. Ignoring...")
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

func (g *Gossiper) SendCommitment(id PollKey, msg Commitment, participants [][2]big.Int, tmpKey ecdsa.PrivateKey, pos int) {
	pkg := PollPacket{
		ID:         id,
		Commitment: &msg,
	}
	input, err := json.Marshal(pkg)
	if err != nil {
		log.Printf("unable to encode as json")
		return
	}

	lrs := linkableRingSignature(input, participants, &tmpKey, pos)
	g.SendPollPacket(&pkg, &Signature{&lrs, nil}, nil)
}

func (g *Gossiper) SendVoteKey(id PollKey, msg VoteKey) {
	pkg := PollPacket{
		ID:      id,
		VoteKey: &msg,
	}

	sig, err := ecSignature(g, pkg)
	if err != nil {
		return
	}

	g.SendPollPacket(&pkg, &sig, nil)
}

func (g *Gossiper) SendVoteKeys(id PollKey, msg VoteKeys) {
	pkg := PollPacket{
		ID:       id,
		VoteKeys: &msg,
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

func (g *Gossiper) SendVote(id PollKey, vote Vote, participants [][2]big.Int, tmpKey ecdsa.PrivateKey, pos int) {
	pkg := PollPacket{
		ID:   id,
		Vote: &vote,
	}

	input, err := json.Marshal(pkg)
	if err != nil {
		log.Printf("unable to encode as json")
		return
	}

	lrs := linkableRingSignature(input, participants, &tmpKey, pos)
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

func getStatus(g *Gossiper) StatusPacketMap {
	g.Status.RLock()
	defer g.Status.RUnlock()

	signatures := make(map[SignatureMap]bool)
	for sig := range g.Status.PktStatus {
		signatures[sig] = true
	}

	reputations := make(map[SignatureMap]bool)
	for rep := range g.Status.ReputationStatus {
		reputations[rep] = true
	}

	return StatusPacketMap{
		signatures,
		reputations,
	}
}

func syncStatus(g *Gossiper, peer net.UDPAddr, status StatusPacket) {
	rcvStatus := status.toMap()
	// check if peer is missing something and send it to him
	myStatus := getStatus(g)
	for wireSig := range myStatus.ReputationPkts {
		_, exist := rcvStatus.ReputationPkts[wireSig]
		if !exist {
			sig := wireSig.toBase()
			writeMsgToUDP(g.Server, &peer, nil, nil, &sig, g.Status.GetRep(wireSig))
		}
	}

	for wireSig := range myStatus.PollPkts {
		_, exist := rcvStatus.PollPkts[wireSig]
		if !exist {
			sig := wireSig.toBase()
			writeMsgToUDP(g.Server, &peer, g.Status.GetPkt(wireSig), nil, &sig, nil)
		}
	}

	// check if I am missing something and request it
	myStatusBase := myStatus.toBase()
	for rep := range rcvStatus.ReputationPkts {
		_, exist := myStatus.ReputationPkts[rep]
		if !exist {
			writeMsgToUDP(g.Server, &peer, nil, &myStatusBase, nil, nil)
			return
		}
	}

	for sig := range rcvStatus.PollPkts {
		_, exist := myStatus.PollPkts[sig]
		if !exist {
			writeMsgToUDP(g.Server, &peer, nil, &myStatusBase, nil, nil)
			return
		}
	}
}

func DispatcherPeersterMessage(g *Gossiper) Dispatcher {
	return func(fromPeer net.UDPAddr, pkg GossipPacket) {
		g.addPeer(fromPeer)

		if pkg.Poll != nil {
			poll := *pkg.Poll

			if !g.SignatureValid(pkg) {
				log.Println("invalid signature found, suspect sender " + fromPeer.String())
				g.Reputations.Suspect(fromPeer.String())
				return
			}

			if pkg.Signature.Linkable != nil {
				if doubleVoted(g, pkg) {
					log.Println("double vote, suspect sender " + fromPeer.String())
					g.Reputations.Suspect(fromPeer.String())
					return
				}
				if pkg.Poll.Vote != nil && invalidVote(g, pkg) {
					log.Println("invalid open message , suspect sender " + fromPeer.String())
					g.Reputations.Suspect(fromPeer.String())
					return
				}
				g.storeTag(pkg)
			}

			added := g.Polls.Store(poll)
			if !added {
				return
			}

			poll.Print(fromPeer)

			g.Status.SetPkt(pkg.Signature.toMap(), pkg.Poll)

			if !g.RunningPolls.Has(poll.ID) {
				g.RunningPolls.Add(poll.ID, VoterHandler(g))
			}

			assert(g.RunningPolls.Has(poll.ID))
			g.RunningPolls.Send(poll, &fromPeer)
		}

		if pkg.Status != nil {
			status := *pkg.Status
			syncStatus(g, fromPeer, status)
		}

		if pkg.Reputation != nil {

			if g.Status.HasRep(pkg.Signature.toMap()) {
				return
			}

			if !repSignatureValid(g, pkg) {
				g.Reputations.Suspect(fromPeer.String())
			}

			g.Status.SetRep(pkg.Signature.toMap(), pkg.Reputation)

			pollID := pkg.Reputation.PollID

			// store Reputation in receivedOpinions[poll]
			g.Reputations.AddPeerOpinion(pkg.Reputation, pollID)

			if len(g.Reputations.PeersOpinions[pollID]) == len(g.Polls.Get(pollID).Participants) {
				// TODO or if timedout
				g.Reputations.AddReputations(pollID)
				g.Reputations.AddTablesWait[pollID] <- true
			}

			g.SendReputation(pollID, &fromPeer)
		}
	}
}

func invalidVote(g *Gossiper, pkg GossipPacket) bool {
	var hash [sha256.Size]byte
	vote := *pkg.Poll.Vote

	toHash := make([]byte, 0)
	toHash = append(toHash, []byte(vote.Option)[:]...)
	toHash = append(toHash, vote.Salt[:]...)

	hash = sha256.Sum256(toHash)

	if g.Status.HasPkt(pkg.Signature.toMap()) {
		storedPkg := g.Status.GetPkt(pkg.Signature.toMap())
		return !(string(storedPkg.Commitment.Hash[:]) == string(hash[:]))
	}

	return true
}

func doubleVoted(g *Gossiper, pkg GossipPacket) bool {
	tag := pkg.Signature.Linkable.Tag
	commit, stored := g.Polls.Get(pkg.Poll.ID).Tags[tag]

	if stored && len(commit) == 1 {
		return !(string(commit[0].Hash[:]) == string(pkg.Poll.Commitment.Hash[:]))
	}

	return false
}

func (g *Gossiper) SignatureValid(pkg GossipPacket) bool {
	poll := pkg.Poll

	if (poll.Commitment != nil || poll.Vote != nil) && !(poll.Commitment != nil && poll.Vote != nil) {
		return pkg.Signature.Linkable != nil && verifySig(*pkg.Signature.Linkable, g.Polls.Get(pkg.Poll.ID).Participants)
	}

	if poll.VoteKeys != nil || poll.Poll != nil || poll.VoteKey != nil {
		input, err := json.Marshal(poll)
		if err != nil {
			log.Printf("unable to encode as json")
		}

		hash := sha256.Sum256(input)

		if poll.VoteKey != nil {
			for _,pubkey := range g.ValidKeys{
				ecKey := ecdsa.PublicKey{Curve(), &pubkey[0], &pubkey[1]}
				if  pkg.Signature.Elliptic != nil && ecdsa.Verify(&ecKey, hash[:],
					&pkg.Signature.Elliptic.R, &pkg.Signature.Elliptic.S){
					return true
				}
			}
			return false
		} else {
			return pkg.Signature.Elliptic != nil && ecdsa.Verify(&pkg.Poll.ID.Origin, hash[:],
				&pkg.Signature.Elliptic.R, &pkg.Signature.Elliptic.S)
		}
	}
	return false
}
func (g *Gossiper) storeTag(pkg GossipPacket) {
	id := pkg.Poll.ID

	var commit Commitment
	if pkg.Poll.Commitment != nil {
		commit = *pkg.Poll.Commitment
	} else if pkg.Poll.Vote != nil {
		vote := *pkg.Poll.Vote

		toHash := make([]byte, 0)
		toHash = append(toHash, []byte(vote.Option)[:]...)
		toHash = append(toHash, vote.Salt[:]...)

		hash := sha256.Sum256(toHash)

		commit = Commitment{
			Hash: hash,
		}
	}

	g.Polls.Lock()
	defer g.Polls.Unlock()

	commitments := g.Polls.m[id.Pack()].Tags[pkg.Signature.Linkable.Tag]

	addCommitment := true
	for _, com := range commitments {
		if string(com.Hash[:]) == string(commit.Hash[:]) {
			addCommitment = false
		}
	}

	if addCommitment {
		commitments = append(commitments, commit)
	}
	g.Polls.m[id.Pack()].Tags[pkg.Signature.Linkable.Tag] = commitments
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
		status := getStatus(gossiper).toBase()
		writeMsgToUDP(gossiper.Server, peer, nil, &status, nil, nil)
	}
}
