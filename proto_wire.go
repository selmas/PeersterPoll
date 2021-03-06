package pollparty

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"math/big"
)

type PublicKeyWire struct {
	X []byte
	Y []byte
}

func PublicKeyWireFromEcdsa(pk ecdsa.PublicKey) PublicKeyWire {
	return PublicKeyWire{
		X: pk.X.Bytes(),
		Y: pk.Y.Bytes(),
	}
}

func (pk PublicKeyWire) toEcdsa() ecdsa.PublicKey {
	return ecdsa.PublicKey{
		Curve: Curve(),
		X:     new(big.Int).SetBytes(pk.X),
		Y:     new(big.Int).SetBytes(pk.Y),
	}
}

type PollKeyWire struct {
	Origin PublicKeyWire
	ID     uint64
}

func (msg PollKey) toWire() PollKeyWire {
	return PollKeyWire{
		Origin: PublicKeyWireFromEcdsa(msg.Origin),
		ID:     msg.ID,
	}
}

func (msg PollKeyWire) toBase() PollKey {
	return PollKey{
		Origin: msg.Origin.toEcdsa(),
		ID:     msg.ID,
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

func (msg CommitmentWire) toBase() Commitment {
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

func (msg VoteWire) toBase() Vote {
	v := Vote{
		Option: msg.Option,
	}
	copy(v.Salt[:], msg.Salt)
	return v
}

type PollPacketWire struct {
	ID         PollKeyWire
	Poll       *Poll
	VoteKey    *VoteKeyWire
	VoteKeys   *VoteKeysWire
	Commitment *CommitmentWire
	Vote       *VoteWire
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

	if pkg.VoteKey != nil {
		nilCount++
	}

	if pkg.VoteKeys != nil {
		nilCount++
	}

	if pkg.Commitment != nil {
		nilCount++
		err = pkg.Commitment.check()
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
	var vk *VoteKeyWire = nil
	if msg.VoteKey != nil {
		wired := msg.VoteKey.toWire()
		vk = &wired
	}

	var vks *VoteKeysWire = nil
	if msg.VoteKeys != nil {
		wired := msg.VoteKeys.toWire()
		vks = &wired
	}

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
		ID:         msg.ID.toWire(),
		Poll:       msg.Poll,
		VoteKey:    vk,
		VoteKeys:   vks,
		Commitment: c,
		Vote:       v,
	}
}

func (msg PollPacketWire) toBase() PollPacket {
	const head = "PollPacketWire: "

	ret := PollPacket{
		ID:   msg.ID.toBase(),
		Poll: msg.Poll,
	}

	if msg.VoteKey != nil {
		wired := msg.VoteKey.toBase()
		ret.VoteKey = &wired
	}

	if msg.VoteKeys != nil {
		wired := msg.VoteKeys.toBase()
		ret.VoteKeys = &wired
	}

	if msg.Commitment != nil {
		wired := msg.Commitment.toBase()
		ret.Commitment = &wired
	}

	if msg.Vote != nil {
		wired := msg.Vote.toBase()
		ret.Vote = &wired
	}

	return ret
}

type StatusPacketWire struct {
	PollPkts       map[SignatureWire]bool
	ReputationPkts map[SignatureWire]bool
}

func (pkg StatusPacketWire) check() error {
	return nil
}

func (pkg StatusPacket) toWire() StatusPacketWire {
	polls := make(map[SignatureWire]bool)
	reps := make(map[SignatureWire]bool)

	for p, _ := range pkg.PollPkts {
		polls[p.toWire()] = true
	}

	for r, _ := range pkg.ReputationPkts {
		reps[r.toWire()] = true
	}

	return StatusPacketWire{
		PollPkts:       polls,
		ReputationPkts: reps,
	}
}

func (pkg StatusPacketWire) toBase() StatusPacket {
	ret := StatusPacket{
		PollPkts:       make(map[Signature]bool),
		ReputationPkts: make(map[Signature]bool),
	}

	for p, _ := range pkg.PollPkts {
		ret.PollPkts[p.toBase()] = true
	}

	for r, _ := range pkg.ReputationPkts {
		ret.ReputationPkts[r.toBase()] = true
	}

	return ret
}

type GossipPacketWire struct {
	Poll       *PollPacketWire
	Signature  *SignatureWire
	Status     *StatusPacketWire
	Reputation *ReputationPacketWire
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

	if pkg.Reputation != nil {
		nilCount++
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

	var r *ReputationPacketWire = nil
	if msg.Reputation != nil {
		wired := msg.Reputation.ToWire()
		r = &wired
	}

	return GossipPacketWire{
		Poll:       p,
		Signature:  sig,
		Status:     s,
		Reputation: r,
	}
}

func (msg GossipPacketWire) ToBase() GossipPacket {
	var ret GossipPacket

	if msg.Poll != nil {
		wire := msg.Poll.toBase()
		ret.Poll = &wire
	}

	if msg.Status != nil {
		wire := msg.Status.toBase()
		ret.Status = &wire
	}

	if msg.Signature != nil {
		wire := msg.Signature.toBase()
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

func (msg EllipticCurveSignatureWire) toBase() EllipticCurveSignature {
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

func (msg SignatureWire) toBase() Signature {
	var ret Signature

	if msg.Linkable != nil {
		l := msg.Linkable.toBase()
		ret.Linkable = &l
	}

	if msg.Elliptic != nil {
		e := msg.Elliptic.toBase()
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

func (msg LinkableRingSignatureWire) toBase() LinkableRingSignature {
	ret := LinkableRingSignature{
		Message: msg.Message,
		C0:      msg.C0,
		S:       make([]*big.Int, 0),
	}

	for _, s := range msg.S {
		ret.S = append(ret.S, new(big.Int).SetBytes(s))
	}

	for i, t := range msg.Tag {
		ret.Tag[i] = new(big.Int).SetBytes(t)
	}

	return ret
}

type VoteKeyWire struct {
	PublicKey PublicKeyWire
	VoteKey   PublicKeyWire
}

func (msg VoteKey) toWire() VoteKeyWire {
	return VoteKeyWire{
		VoteKey:   PublicKeyWireFromEcdsa(msg.tmpKey),
	}
}

func (msg VoteKeyWire) toBase() VoteKey {
	return VoteKey{
		tmpKey:    msg.VoteKey.toEcdsa(),
	}
}

type VoteKeysWire struct {
	Keys []VoteKeyWire
}

func (msg VoteKeys) toWire() VoteKeysWire {
	keys := make([]VoteKeyWire, len(msg.Keys))

	for i, k := range msg.Keys {
		keys[i] = k.toWire()
	}

	return VoteKeysWire{
		Keys: keys,
	}
}

func (msg VoteKeysWire) toBase() VoteKeys {
	keys := make([]VoteKey, len(msg.Keys))

	for i, k := range msg.Keys {
		keys[i] = k.toBase()
	}

	return VoteKeys{
		Keys: keys,
	}
}
