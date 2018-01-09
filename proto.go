package pollparty

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
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
		Origin: ecdsa.PublicKey{Curve: Curve(), X: x, Y: y},
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

const SaltSize = 20

// TODO not here, only used with voting
type Commitment struct {
	Hash [sha256.Size]byte
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

type StatusPacket struct {
	Infos map[PollKey]ShareablePollInfo
}

type GossipPacket struct {
	Poll      *PollPacket
	Signature *Signature
	Status    *StatusPacket
}

type EllipticCurveSignature struct {
	R big.Int
	S big.Int
}

type Signature struct {
	Linkable *LinkableRingSignature
	Elliptic *EllipticCurveSignature
}
