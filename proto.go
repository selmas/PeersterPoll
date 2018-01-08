package pollparty

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type PollKey struct {
	Origin ecdsa.PublicKey
	ID     uint64
}

func (msg PollKey) Check() error {
	if msg.Origin.X == nil || msg.Origin.Y == nil {
		log.Println("empty origin") // TODO bad rep
	}

	return nil
}

const PollKeySep = "|"

func (msg PollKey) String() string {
	return msg.Origin.X.String() + PollKeySep + msg.Origin.Y.String() + PollKeySep + strconv.FormatUint(msg.ID, 10)
}

func PollKeyFromString(packed string) (PollKey, error) {
	var ret PollKey
	errRet := func(v string) (PollKey, error) {
		return ret, errors.New("unable to parse \"" + v + "\" as int")
	}

	splitted := strings.SplitN(packed, PollKeySep, 3)

	x, ok := new(big.Int).SetString(splitted[0], 10)
	if !ok {
		return errRet(splitted[0])
	}

	y, ok := new(big.Int).SetString(splitted[1], 10)
	if !ok {
		return errRet(splitted[1])
	}

	id, err := strconv.ParseUint(splitted[2], 10, 64)
	if err != nil {
		return ret, err
	}

	ret = PollKey{
		Origin: ecdsa.PublicKey{Curve: curve, X: x, Y: y},
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

type PollPacket struct {
	ID              PollKey
	Poll            *Poll
	Commitment      *Commitment
	PollCommitments *PollCommitments
	Vote            *Vote
}

// TODO warn if asked for same key in separated round -> bad rep
type StatusPacket struct {
	Infos map[PollKey]ShareablePollInfo
}

type GossipPacket struct {
	Poll      *PollPacket
	Signature *Signature
	Status    *StatusPacket
}

type EllipticCurveSignature struct {
	r *big.Int
	s *big.Int
}

type Signature struct {
	linkableRingSig  *LinkableRingSignature
	ellipticCurveSig *EllipticCurveSignature
}
