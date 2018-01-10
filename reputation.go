package pollparty

import (
	"crypto/ecdsa"
	secrand "crypto/rand" // alias needed as we import two libraries with name "rand"
	"crypto/sha256"
	"encoding/json"
	"log"
	"math/rand"
	"net"
)

// Reputation Opinions ---------------------------------------------------------------------------

type RepOpinions map[string]int

func (opinions RepOpinions) Suspect(peer string) {
	// also add peer to blacklist
	opinions[peer] = -1
}

func (opinions RepOpinions) Trust(peer string) {
	// Opinion can only change from trusting to suspecting,
	// not the other way around
	if opinions[peer] != -1 {
		opinions[peer] = 1
	}
}

func (opinions RepOpinions) hasInvalidOpinion() bool {
	for _, rep := range opinions {
		if rep != 1 && rep != -1 {
			return true
		}
	}
	return false
}

func (opinions RepOpinions) completePeers(peers map[string]bool) {
	for peer := range peers {
		opinions.Trust(peer)
	}
}

// Blacklist -------------------------------------------------------------------------------------

type Blacklist map[string]bool

func (bList Blacklist) IsBlacklisted(peer string) bool {
	return bList[peer]
}

func (bList Blacklist) add(peer string) {
	bList[peer] = true
}

func (bList Blacklist) Update(repTable map[string]int) {
	for peer, rep := range repTable {
		if rep < 0 {
			bList.add(peer)
		}
	}
}

func (bList Blacklist) String() string {
	str := "Peer\t\tBlacklisted?\n"
	for peer, status := range bList {
		str += peer + "\t\t"
		if status {
			str += "yes\n"
		} else {
			str += "no\n"
		}
	}

	return str
}

// Reputation Info -------------------------------------------------------------------------------

type ReputationInfo struct {
	Opinions      RepOpinions
	Blacklist     Blacklist
	PeersOpinions map[PollKey]map[ecdsa.PublicKey]RepOpinions
	AddTablesWait map[PollKey]chan bool
}

func NewReputationInfo() ReputationInfo {
	return ReputationInfo{
		Opinions:      make(RepOpinions),
		Blacklist:     make(Blacklist),
		PeersOpinions: make(map[PollKey]map[ecdsa.PublicKey]RepOpinions),
		AddTablesWait: make(map[PollKey]chan bool),
	}
}

func (repInfo ReputationInfo) AddPeerOpinion(pkg *ReputationPacket, pollID PollKey) {
	//repInfo.PeersOpinions[pollID] = append(repInfo.PeersOpinions[pollID], opinion)
	if repInfo.PeersOpinions[pollID] == nil {
		repInfo.PeersOpinions[pollID] = make(map[ecdsa.PublicKey]RepOpinions)
	}

	opinion := pkg.Opinions
	signer := pkg.Signer

	repInfo.PeersOpinions[pollID][signer] = opinion
}

func (repInfo ReputationInfo) AddReputations(pollID PollKey) {

	repTable := make(map[string]int)
	peers := make(map[string]bool)

	for _, peerOpinions := range repInfo.PeersOpinions[pollID] {
		for peer := range peerOpinions {
			peers[peer] = true
		}
	}

	for _, peerOpinions := range repInfo.PeersOpinions[pollID] {

		// If a peer has an invalid opinion of another peer
		// none of its opinions will be taken into account
		if peerOpinions.hasInvalidOpinion() {
			continue
		}

		peerOpinions.completePeers(peers)

		for peer, rep := range peerOpinions {
			tempUpdateRep(peer, rep, repTable)
		}
	}

	repInfo.Blacklist.Update(repTable)
}

func tempUpdateRep(peer string, rep int, repTable map[string]int) {
	_, exists := repTable[peer]

	if !exists {
		repTable[peer] = 0
	}

	repTable[peer] += rep
}

func (repInfo ReputationInfo) IsBlacklisted(peer string) bool {
	return repInfo.Blacklist.IsBlacklisted(peer)
}

func (repInfo ReputationInfo) Suspect(peer string) {
	repInfo.Opinions.Suspect(peer)
	repInfo.Blacklist.add(peer)
}

// Reputation Packet -----------------------------------------------------------------------------

type ReputationPacket struct {
	Signer   ecdsa.PublicKey
	Opinions RepOpinions
	PollID   PollKey
}

func UpdateReputations(g *Gossiper, pollID PollKey) {

	g.SendReputation(pollID, nil)

	<-g.Reputations.AddTablesWait[pollID]
}

func (g *Gossiper) SendReputationPacket(msg *ReputationPacket, sig *Signature, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&g.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(g.Server, peer, nil, nil, sig, msg)

		printFlippedCoin(peer, "reputation opinions")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func (g *Gossiper) SendReputation(key PollKey, fromPeer *net.UDPAddr) {
	pkg := ReputationPacket{
		PollID:   key,
		Opinions: g.Reputations.Opinions,
		Signer:   g.KeyPair.PublicKey,
	}

	sig, err := repSignature(g, pkg)
	if err != nil {
		return
	}

	g.SendReputationPacket(&pkg, &sig, fromPeer)
}

func repSignature(g *Gossiper, rep ReputationPacket) (Signature, error) {
	input, err := json.Marshal(rep)
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

func repSignatureValid(g *Gossiper, pkg GossipPacket) bool {
	rep := pkg.Reputation

	if pkg.Signature != nil && pkg.Signature.Elliptic != nil {
		input, err := json.Marshal(rep)
		if err != nil {
			log.Printf("unable to encode as json")
		}

		hash := sha256.New()
		_, err = hash.Write(input)
		if err != nil {
			log.Printf("error generating elliptic curve signature")
		}

		return ecdsa.Verify(&rep.Signer, hash.Sum(nil), &pkg.Signature.Elliptic.R,
			&pkg.Signature.Elliptic.S)
	}

	return false
}

// Wire ------------------------------------------------------------------------------------------

type ReputationPacketWire struct {
	Opinions RepOpinions
	PollID   PollKeyWire
	Signer   PublicKeyWire
}

func (msg ReputationPacket) ToWire() ReputationPacketWire {
	return ReputationPacketWire{
		PollID:   msg.PollID.toWire(),
		Opinions: msg.Opinions,
		Signer:   PublicKeyWireFromEcdsa(msg.Signer),
	}
}

func (msg ReputationPacketWire) ToBase() ReputationPacket {
	return ReputationPacket{
		PollID:   msg.PollID.toBase(),
		Opinions: msg.Opinions,
		Signer:   msg.Signer.toEcdsa(),
	}
}
