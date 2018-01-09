package pollparty

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"math/big"
)

type PollKeyWire struct {
	X  []byte
	Y  []byte
	ID uint64
}

func (msg PollKey) toWire() PollKeyWire {
	return PollKeyWire{
		X:  msg.Origin.X.Bytes(),
		Y:  msg.Origin.Y.Bytes(),
		ID: msg.ID,
	}
}

func (msg PollKeyWire) ToBase() PollKey {
	return PollKey{
		Origin: ecdsa.PublicKey{
			Curve: Curve(),
			X:     new(big.Int).SetBytes(msg.X),
			Y:     new(big.Int).SetBytes(msg.Y),
		},
		ID: msg.ID,
	}
}

// used only on the network because protobuf lib fail to encode fixed size array
type CommitmentWire struct {
	Hash []byte
}

func (msg CommitmentWire) check() error {
	if len(msg.Hash) != sha256.Size {
		return errors.New("invalid hash size")
	}

	return nil
}

func (msg Commitment) toWire() CommitmentWire {
	return CommitmentWire{
		Hash: msg.Hash[:],
	}
}

func (msg CommitmentWire) ToBase() Commitment {
	var c Commitment
	copy(c.Hash[:], msg.Hash)
	return c
}

type VoteWire struct {
	Salt   []byte
	Option string
}

func (msg VoteWire) check() error {
	if len(msg.Salt) != SaltSize {
		return errors.New("invalid salt size")
	}

	return nil
}

func (msg Vote) toWire() VoteWire {
	return VoteWire{
		Salt:   msg.Salt[:],
		Option: msg.Option,
	}
}

func (msg VoteWire) ToBase() Vote {
	var v Vote
	copy(v.Salt[:], msg.Salt)
	return v
}

type PollPacketWire struct {
	ID              PollKeyWire
	Poll            *Poll
	Commitment      *CommitmentWire
	PollCommitments *PollCommitments
	Vote            *VoteWire
}

func (pkg PollPacketWire) check() error {
	var nilCount uint = 0
	var err error = nil
	retErr := func(err string) error {
		return errors.New("PollPacketWire: " + err)
	}

	if pkg.Poll != nil {
		nilCount++
	}

	if pkg.Commitment != nil {
		nilCount++
		err = pkg.Commitment.check()
	}

	if pkg.PollCommitments != nil {
		nilCount++
	}

	if pkg.Vote != nil {
		nilCount++
	}

	if err != nil {
		return retErr(err.Error())
	}

	if nilCount > 1 {
		return retErr("too much fields defined")
	} else if nilCount == 0 {
		return retErr("no field defined")
	}

	return nil
}

func (msg PollPacket) toWire() PollPacketWire {
	var c *CommitmentWire = nil
	if msg.Commitment != nil {
		wired := msg.Commitment.toWire()
		c = &wired
	}

	var v *VoteWire = nil
	if msg.Vote != nil {
		wired := msg.Vote.toWire()
		v = &wired
	}

	return PollPacketWire{
		ID:              msg.ID.toWire(),
		Poll:            msg.Poll,
		Commitment:      c,
		PollCommitments: msg.PollCommitments,
		Vote:            v,
	}
}

func (msg PollPacketWire) ToBase() PollPacket {
	const head = "GossipPacketWire: "

	ret := PollPacket{
		ID:              msg.ID.ToBase(),
		Poll:            msg.Poll,
		PollCommitments: msg.PollCommitments,
	}

	var c *Commitment
	if msg.Commitment != nil {
		wired := msg.Commitment.ToBase()
		c = &wired
	}

	var v *Vote
	if msg.Vote != nil {
		wired := msg.Vote.ToBase()
		v = &wired
	}

	ret.Commitment = c
	ret.Vote = v

	return ret
}

// nice protobuf, do not support map with any type
type StatusPacketWire struct {
	Infos map[string]ShareablePollInfo
}

func (pkg StatusPacketWire) check() error {
	errRet := func(err error) error {
		return errors.New("StatusPacketWire: " + err.Error())
	}

	for k, _ := range pkg.Infos {
		_, err := PollKeyFromString(k)
		if err != nil {
			return errRet(err)
		}
	}

	return nil
}

func (pkg StatusPacket) toWire() StatusPacketWire {
	infos := make(map[string]ShareablePollInfo)

	for id, info := range pkg.Infos {
		infos[id.String()] = info
	}

	return StatusPacketWire{
		Infos: infos,
	}
}

func (pkg StatusPacketWire) ToBase() StatusPacket {
	ret := StatusPacket{
		Infos: make(map[PollKey]ShareablePollInfo),
	}

	for k, info := range pkg.Infos {
		id, _ := PollKeyFromString(k) // check()'ed before
		ret.Infos[id] = info
	}

	return ret
}

type GossipPacketWire struct {
	Poll      *PollPacketWire
	Signature *SignatureWire
	Status    *StatusPacketWire
}

func (pkg GossipPacketWire) Check() error {
	var nilCount uint = 0
	var err error = nil
	errRet := func(err string) error {
		return errors.New("GossipPacketWire: " + err)
	}

	if pkg.Poll != nil {
		nilCount++
		err = pkg.Poll.check()

		if pkg.Signature == nil {
			return errRet("poll without signature")
		}
	}

	if pkg.Status != nil {
		nilCount++
		err = pkg.Status.check()
	}

	if err != nil {
		return errRet(err.Error())
	}

	if nilCount > 1 {
		return errRet("too much fields defined")
	} else if nilCount == 0 {
		return errRet("no field defined")
	}

	return nil
}

func (msg GossipPacket) ToWire() GossipPacketWire {
	// why did we invented good languages when we can write this nice
	// boilerplate code in go instead
	var p *PollPacketWire = nil
	if msg.Poll != nil {
		wired := msg.Poll.toWire()
		p = &wired
	}

	var s *StatusPacketWire = nil
	if msg.Status != nil {
		wired := msg.Status.toWire()
		s = &wired
	}

	var sig *SignatureWire = nil
	if msg.Signature != nil {
		wired := msg.Signature.toWire()
		sig = &wired
	}

	return GossipPacketWire{
		Poll:      p,
		Signature: sig,
		Status:    s,
	}
}

func (msg GossipPacketWire) ToBase() GossipPacket {
	var ret GossipPacket

	if msg.Poll != nil {
		wire := msg.Poll.ToBase()
		ret.Poll = &wire
	}

	if msg.Status != nil {
		wire := msg.Status.ToBase()
		ret.Status = &wire
	}

	if msg.Signature != nil {
		wire := msg.Signature.ToBase()
		ret.Signature = &wire
	}

	return ret
}

type EllipticCurveSignatureWire struct {
	R []byte
	S []byte
}

func (msg EllipticCurveSignature) toWire() EllipticCurveSignatureWire {
	return EllipticCurveSignatureWire{
		R: msg.R.Bytes(),
		S: msg.S.Bytes(),
	}
}

func (msg EllipticCurveSignatureWire) ToBase() EllipticCurveSignature {
	r := new(big.Int).SetBytes(msg.R)
	s := new(big.Int).SetBytes(msg.S)

	return EllipticCurveSignature{
		R: *r,
		S: *s,
	}
}

type SignatureWire struct {
	Linkable *LinkableRingSignatureWire
	Elliptic *EllipticCurveSignatureWire
}

func (msg SignatureWire) check() error {
	var err error = nil

	if (msg.Linkable == nil) != (msg.Elliptic == nil) { // bool xor
		return errors.New("SignatureWire: no/all field definied")
	}

	if msg.Linkable != nil {
		err = msg.Linkable.check()
	}

	return err
}

func (msg Signature) toWire() SignatureWire {
	var l *LinkableRingSignatureWire = nil
	if msg.Linkable != nil {
		wired := msg.Linkable.toWire()
		l = &wired
	}

	var e *EllipticCurveSignatureWire = nil
	if msg.Elliptic != nil {
		wired := msg.Elliptic.toWire()
		e = &wired
	}

	return SignatureWire{
		Linkable: l,
		Elliptic: e,
	}
}

func (msg SignatureWire) ToBase() Signature {
	var ret Signature

	if msg.Linkable != nil {
		l := msg.Linkable.ToBase()
		ret.Linkable = &l
	}

	if msg.Elliptic != nil {
		e := msg.Elliptic.ToBase()
		ret.Elliptic = &e
	}

	return ret
}

type LinkableRingSignatureWire struct {
	Message []byte
	C0      []byte
	S       [][]byte
	Tag     [][]byte
}

func (msg LinkableRingSignatureWire) check() error {
	if len(msg.Tag) != 2 {
		return errors.New("LinkableRingSignatureWire: tag size isn't 2")
	}

	return nil
}

func (msg LinkableRingSignature) toWire() LinkableRingSignatureWire {
	ss := make([][]byte, len(msg.S))
	for i, s := range msg.S {
		ss[i] = s.Bytes()
	}

	tag := make([][]byte, 2)
	for i, t := range msg.Tag {
		tag[i] = t.Bytes()
	}

	return LinkableRingSignatureWire{
		Message: msg.Message,
		C0:      msg.C0,
		S:       ss,
		Tag:     tag,
	}
}

func (msg LinkableRingSignatureWire) ToBase() LinkableRingSignature {
	ret := LinkableRingSignature{
		Message: msg.Message,
		C0:      msg.C0,
		S:       make([]*big.Int, len(msg.S)),
	}

	for i, s := range msg.S {
		ret.S[i] = new(big.Int).SetBytes(s)
	}

	for i, t := range msg.Tag {
		ret.Tag[i] = new(big.Int).SetBytes(t)
	}

	return ret
}
