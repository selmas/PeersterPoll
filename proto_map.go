package pollparty

import (
	"crypto/ecdsa"
	"math/big"
)

const PackBigIntBase = 36 // len(0-9) + len(a-z)

type BigIntMap struct {
	Value string
}

func BigIntMapFrom(v *big.Int) BigIntMap {
	return BigIntMap{
		v.Text(PackBigIntBase),
	}
}

func (i BigIntMap) toBase() *big.Int {
	ret, ok := new(big.Int).SetString(i.Value, PackBigIntBase)
	if !ok {
		panic("fail to base")
	}
	return ret
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

type StatusPacketMap struct {
	PollPkts       map[SignatureMap]bool
	ReputationPkts map[SignatureMap]bool
}

func (s StatusPacket) toMap() StatusPacketMap {
	ret := StatusPacketMap{
		make(map[SignatureMap]bool),
		make(map[SignatureMap]bool),
	}

	for sig, _ := range s.PollPkts {
		ret.PollPkts[sig.toMap()] = true
	}

	for sig, _ := range s.ReputationPkts {
		ret.ReputationPkts[sig.toMap()] = true
	}

	return ret
}

func (s StatusPacketMap) toBase() StatusPacket {
	ret := StatusPacket{
		make(map[Signature]bool),
		make(map[Signature]bool),
	}

	for sig, _ := range s.PollPkts {
		ret.PollPkts[sig.toBase()] = true
	}

	for sig, _ := range s.ReputationPkts {
		ret.ReputationPkts[sig.toBase()] = true
	}

	return ret
}

type EllipticCurveSignatureMap struct {
	R BigIntMap
	S BigIntMap
}

func (s EllipticCurveSignature) toMap() EllipticCurveSignatureMap {
	return EllipticCurveSignatureMap{
		BigIntMapFrom(&s.R),
		BigIntMapFrom(&s.S),
	}
}

func (s EllipticCurveSignatureMap) toBase() EllipticCurveSignature {
	return EllipticCurveSignature{
		*s.R.toBase(),
		*s.S.toBase(),
	}
}

type LinkableRingSignatureMap struct {
	Message string
	C0      string
	S       [32]BigIntMap // TODO depends on max network size, thanks go
	SSize   int
	Tag     [2]BigIntMap
}

func (s LinkableRingSignature) toMap() LinkableRingSignatureMap {
	ret := LinkableRingSignatureMap{
		Message: string(s.Message),
		C0:      string(s.C0),
		SSize:   len(s.S),
	}

	for i, v := range s.S {
		ret.S[i] = BigIntMapFrom(v)
	}

	for i, v := range s.Tag {
		ret.Tag[i] = BigIntMapFrom(v)
	}

	return ret
}

func (s LinkableRingSignatureMap) toBase() LinkableRingSignature {
	ret := LinkableRingSignature{
		Message: []byte(s.Message),
		C0:      []byte(s.C0),
		S:       make([]*big.Int, 0),
	}

	for i := 0; i < s.SSize; i++ {
		v := s.S[i]
		ret.S = append(ret.S, v.toBase())
	}

	for i, v := range s.Tag {
		ret.Tag[i] = v.toBase()
	}

	return ret
}

type SignatureMap struct {
	Linkable      LinkableRingSignatureMap
	LinkableEmpty bool
	Elliptic      EllipticCurveSignatureMap
	EllipticEmpty bool
}

func (s Signature) toMap() SignatureMap {
	ret := SignatureMap{
		LinkableEmpty: s.Linkable == nil,
		EllipticEmpty: s.Elliptic == nil,
	}

	if s.Linkable != nil {
		ret.Linkable = s.Linkable.toMap()
	}

	if s.Elliptic != nil {
		ret.Elliptic = s.Elliptic.toMap()
	}

	return ret
}

func (s SignatureMap) toBase() Signature {
	var l *LinkableRingSignature
	if !s.LinkableEmpty {
		sig := s.Linkable.toBase()
		l = &sig
	}

	var e *EllipticCurveSignature
	if !s.EllipticEmpty {
		sig := s.Elliptic.toBase()
		e = &sig
	}

	return Signature{l, e}
}
