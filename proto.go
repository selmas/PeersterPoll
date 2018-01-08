package pollparty

import (
	"crypto/sha256"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"math/big"
	"crypto/ecdsa"
)

type PollKey struct {
	Origin *ecdsa.PublicKey
	ID     uint64
}

func (msg PollKey) Check() error {
	if msg.Origin.X == nil || msg.Origin.Y == nil {
		log.Println("empty origin") // TODO bad rep
	}

	return nil
}

const PollKeySep = "/"

func (msg PollKey) String() string {
	return msg.Origin.X.String() + PollKeySep + msg.Origin.Y.String() + PollKeySep + strconv.FormatUint(msg.ID, 10)
}

func PollKeyFromString(packed string) (PollKey, error) {
	var ret PollKey

	splitted := strings.SplitN(packed, PollKeySep, 2)

	id, err := strconv.ParseUint(splitted[1], 10, 64)
	if err != nil {
		return ret, err
	}
	x, _ := new(big.Int).SetString(splitted[0],10)
	y, _ :=	new(big.Int).SetString(splitted[1],10)
	ret = PollKey{
		Origin: &ecdsa.PublicKey{Curve: curve, X: x, Y: y},
		ID:     id,
	}

	return ret, nil
}

type Poll struct {
	Question  string
	Options   []string
	StartTime time.Time
	Duration  time.Duration // After duration has passed, can no longer participate in votes
}

func (p Poll) IsTooLate() bool {
	return p.StartTime.Add(p.Duration).Before(time.Now())
}

func (msg Poll) Check() error {
	if msg.Question == "" {
		log.Println("no question")
	}

	if len(msg.Options) == 0 {
		log.Println("no choices")
	}

	return nil // TODO
}

const SaltSize = 20

// TODO not here, only used with voting
type Commitment struct {
	Hash [sha256.Size]byte
}

func (msg Commitment) Check() error {
	return nil
}

func (msg Commitment) ToWire() CommitmentWire {
	return CommitmentWire{
		Hash: msg.Hash[:],
	}
}

// used only on the network because protobuf lib fail to encode fixed size array
type CommitmentWire struct {
	Hash []byte
}

func (msg CommitmentWire) Check() error {
	return nil // TODO check hash and salt size
}

func (msg CommitmentWire) ToBase() (Commitment, error) {
	var c Commitment

	if len(msg.Hash) != sha256.Size {
		return c, errors.New("invalid hash size")
	}

	copy(c.Hash[:], msg.Hash)

	return c, nil
}

func NewCommitment(answer string) Commitment {
	var salt [SaltSize]byte
	rand.Read(salt[:])

	// TODO move to dedicated hash func
	toHash := make([]byte, 0)
	toHash = append(toHash, []byte(answer)[:]...)
	toHash = append(toHash, salt[:]...)

	hash := sha256.Sum256(toHash)

	return Commitment{
		Hash: hash,
	}
}

type PollCommitments struct {
	Commitments []Commitment
}

func (msg PollCommitments) Check() error {
	for _, v := range msg.Commitments {
		err := v.Check()
		if err != nil {
			return err
		}
	}

	return nil
}

func (msg PollCommitments) Has(c Commitment) bool {
	for _, v := range msg.Commitments {
		if v == c {
			return true
		}
	}

	return false
}

type Vote struct {
	Salt   [SaltSize]byte
	Option string
}

func (msg Vote) ToWire() VoteWire {
	return VoteWire{
		Salt:   msg.Salt[:],
		Option: msg.Option,
	}
}

type VoteWire struct {
	Salt   []byte
	Option string
}

func (msg VoteWire) Check() error {
	return nil
}

func (msg VoteWire) ToBase() (Vote, error) {
	var v Vote

	if len(msg.Salt) != SaltSize {
		return v, errors.New("invalid salt size")
	}

	copy(v.Salt[:], msg.Salt)

	return v, nil
}

type PollPacket struct {
	ID              PollKey
	Poll            *Poll
	Commitment      *Commitment
	PollCommitments *PollCommitments
	Vote            *Vote
}

func (msg PollPacket) ToWire() PollPacketWire {
	c := msg.Commitment.ToWire()
	v := msg.Vote.ToWire()
	return PollPacketWire{
		ID:              msg.ID,
		Poll:            msg.Poll,
		Commitment:      &c,
		PollCommitments: msg.PollCommitments,
		Vote:            &v,
	}
}

type PollPacketWire struct {
	ID              PollKey
	Poll            *Poll
	Commitment      *CommitmentWire
	PollCommitments *PollCommitments
	Vote            *VoteWire
}

func (msg PollPacketWire) ToBase() (PollPacket, error) {
	const head = "GossipPacketWire: "

	ret := PollPacket{
		ID:              msg.ID,
		Poll:            msg.Poll,
		PollCommitments: msg.PollCommitments,
	}

	c, err := msg.Commitment.ToBase()
	if err != nil {
		return ret, errors.New(head + err.Error())
	}

	v, err := msg.Vote.ToBase()
	if err != nil {
		return ret, errors.New(head + err.Error())
	}

	ret.Commitment = &c
	ret.Vote = &v

	return ret, nil
}

func (pkg PollPacketWire) Check() error {
	var nilCount uint = 0
	var err error = nil

	if pkg.Poll != nil {
		nilCount++
		err = pkg.Poll.Check()
	}

	if pkg.Commitment != nil {
		nilCount++
		err = pkg.Commitment.Check()
	}

	if pkg.PollCommitments != nil {
		nilCount++
		err = pkg.PollCommitments.Check()
	}

	if pkg.Vote != nil {
		nilCount++
		err = pkg.PollCommitments.Check()
	}

	if err != nil {
		return errors.New("PollPacketWire: " + err.Error())
	}
	if nilCount > 1 {
		return errors.New("too much fields defined")
	} else if nilCount == 0 {
		return errors.New("no field defined")
	}

	return nil
}

// TODO warn if asked for same key in separated round -> bad rep
type StatusPacket struct {
	Polls            []PollKey                // may be found via forwarding
	Commitments      map[PollKey][]Commitment // so we can fetch the missing commitments
	PollsCommitments []PollKey                // there is only one PollCommitments
	Votes            map[PollKey][]Vote       // again, so we can fetch the missing votes
}

func (pkg StatusPacket) Check() error {
	for _, poll := range pkg.Polls {
		err := poll.Check()
		if err != nil {
			return errors.New("StatusPacket: " + err.Error())
		}
	}

	return nil
}

type EllipticCurveSignature struct {
	r *big.Int
	s *big.Int
}

type Signature struct {
	linkableRingSig  *LinkableRingSignature
	ellipticCurveSig *EllipticCurveSignature
}

type GossipPacket struct {
	Poll      *PollPacket
	Signature *Signature
	Status    *StatusPacket
}

func (msg GossipPacket) ToWire() GossipPacketWire {
	wire := msg.Poll.ToWire()
	return GossipPacketWire{
		Poll:   &wire,
		Status: msg.Status,
	}
}

type GossipPacketWire struct {
	Poll   *PollPacketWire
	Status *StatusPacket
}

func (msg GossipPacketWire) ToBase() (GossipPacket, error) {
	ret := GossipPacket{
		Status: msg.Status,
	}

	wire, err := msg.Poll.ToBase()
	if err != nil {
		return ret, errors.New("GossipPacketWire: " + err.Error())
	}

	ret.Poll = &wire

	return ret, nil
}

func (pkg GossipPacketWire) Check() error {
	var nilCount uint = 0
	var err error = nil
	const head string = "GossipPacket: "

	if pkg.Poll != nil {
		nilCount++
		err = pkg.Poll.Check()
	}

	if pkg.Status != nil {
		nilCount++
		err = pkg.Status.Check()
	}

	if nilCount > 1 {
		return errors.New(head + "too much fields defined")
	} else if nilCount == 0 {
		return errors.New(head + "no field defined")
	}

	if err != nil {
		return errors.New(head + err.Error())
	}

	return nil
}
