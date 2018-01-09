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

const PackBigIntBase = 36 // len(0-9) + len(a-z)

type PollKey struct {
	Origin ecdsa.PublicKey
	ID     uint64
}

type PublicKeyMap struct {
	X string
	Y string
}

func PublicKeyMapFromEcdsa(k ecdsa.PublicKey) PublicKeyMap {
	return PublicKeyMap{
		X: k.X.Text(PackBigIntBase),
		Y: k.Y.Text(PackBigIntBase),
	}
}

func (pk PublicKeyMap) toEcdsa() ecdsa.PublicKey {
	// TODO we don't handle errors, as it should be safe world
	x, _ := new(big.Int).SetString(pk.X, PackBigIntBase)
	y, _ := new(big.Int).SetString(pk.Y, PackBigIntBase)

	if x == nil || y == nil {
		panic("fail to unpack")
	}

	return ecdsa.PublicKey{
		Curve: Curve(),
		X:     x,
		Y:     y,
	}
}

type PollKeyMap struct {
	Origin PublicKeyMap
	ID     uint64
}

func (k PollKey) Pack() PollKeyMap {
	return PollKeyMap{
		Origin: PublicKeyMapFromEcdsa(k.Origin),
		ID:     k.ID,
	}
}

func (k PollKeyMap) Unpack() PollKey {

	return PollKey{
		Origin: k.Origin.toEcdsa(),
		ID:     k.ID,
	}
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

func NewCommitment(answer string) (Commitment, [SaltSize]byte) {
	var salt [SaltSize]byte
	rand.Read(salt[:])

	// TODO move to dedicated hash func
	toHash := make([]byte, 0)
	toHash = append(toHash, []byte(answer)[:]...)
	toHash = append(toHash, salt[:]...)

	hash := sha256.Sum256(toHash)

	return Commitment{
		Hash: hash,
	}, salt
}

type VoteKey struct {
	publicKey ecdsa.PublicKey
	tmpKey    ecdsa.PublicKey
}

func (vk VoteKey) Pack() VoteKeyMap {
	return VoteKeyMap{
		publicKey: PublicKeyMapFromEcdsa(vk.publicKey),
		tmpKey:    PublicKeyMapFromEcdsa(vk.tmpKey),
	}
}

// TODO do not use *Wire
type VoteKeyMap struct {
	publicKey PublicKeyMap
	tmpKey    PublicKeyMap
}

func (vk VoteKeyMap) Unpack() VoteKey {
	return VoteKey{
		publicKey: vk.publicKey.toEcdsa(),
		tmpKey:    vk.tmpKey.toEcdsa(),
	}
}

type VoteKeys struct {
	Keys []VoteKey
}

func (msg VoteKeys) ToParticipants() [][2]*big.Int {
	ret := make([][2]*big.Int, len(msg.Keys))

	for i, k := range msg.Keys {
		ret[i][0] = k.publicKey.X
		ret[i][0] = k.publicKey.Y
	}

	return ret
}

type Vote struct {
	Salt   [SaltSize]byte
	Option string
}

type PollPacket struct {
	ID         PollKey
	Poll       *Poll
	VoteKey    *VoteKey
	VoteKeys   *VoteKeys
	Commitment *Commitment
	Vote       *Vote
}

type StatusPacket struct {
	PollPkts       map[Signature]bool
	ReputationPkts map[Signature]bool
}

type GossipPacket struct {
	Poll       *PollPacket
	Signature  *Signature
	Status     *StatusPacket
	Reputation *ReputationPacket
}

type EllipticCurveSignature struct {
	R big.Int
	S big.Int
}

type Signature struct {
	Linkable *LinkableRingSignature
	Elliptic *EllipticCurveSignature
}
