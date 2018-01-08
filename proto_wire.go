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

func (msg PollKey) ToWire() PollKeyWire {
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

func (msg PollPacket) ToWire() PollPacketWire {
	var c *CommitmentWire = nil
	if msg.Commitment != nil {
		wired := msg.Commitment.ToWire()
		c = &wired
	}

	var v *VoteWire = nil
	if msg.Vote != nil {
		wired := msg.Vote.ToWire()
		v = &wired
	}

	return PollPacketWire{
		ID:              msg.ID.ToWire(),
		Poll:            msg.Poll,
		Commitment:      c,
		PollCommitments: msg.PollCommitments,
		Vote:            v,
	}
}

type PollPacketWire struct {
	ID              PollKeyWire
	Poll            *Poll
	Commitment      *CommitmentWire
	PollCommitments *PollCommitments
	Vote            *VoteWire
}

func (msg PollPacketWire) ToBase() (PollPacket, error) {
	const head = "GossipPacketWire: "

	ret := PollPacket{
		ID:              msg.ID.ToBase(),
		Poll:            msg.Poll,
		PollCommitments: msg.PollCommitments,
	}

	var c *Commitment
	if msg.Commitment != nil {
		wired, err := msg.Commitment.ToBase()
		if err != nil {
			return ret, errors.New(head + err.Error())
		}
		c = &wired
	}

	var v *Vote
	if msg.Vote != nil {
		wired, err := msg.Vote.ToBase()
		if err != nil {
			return ret, errors.New(head + err.Error())
		}
		v = &wired
	}

	ret.Commitment = c
	ret.Vote = v

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

// nice protobuf, do not support map with any type
type StatusPacketWire struct {
	Infos map[string]ShareablePollInfo
}

func (pkg StatusPacketWire) Check() error {
	/*errRet := func(err error) error {
		return errors.New("StatusPacketWire: " + err.Error())
	}

	for id, poll := range pkg.Infos {
		/* TODO err := poll.Check()
		if err != nil {
			return errRet(err)
		}
	}*/

	return nil
}

func (pkg StatusPacket) ToWire() StatusPacketWire {
	infos := make(map[string]ShareablePollInfo)

	for id, info := range pkg.Infos {
		infos[id.String()] = info
	}

	return StatusPacketWire{
		Infos: infos,
	}
}

func (pkg StatusPacketWire) ToBase() (StatusPacket, error) {
	// why did we invented good languages when we can write this nice
	// boilerplate code in go instead
	errRet := func(err error) error {
		return errors.New("GossipPacketWire: " + err.Error())
	}

	ret := StatusPacket{
		Infos: make(map[PollKey]ShareablePollInfo),
	}

	for k, info := range pkg.Infos {
		id, err := PollKeyFromString(k)
		if err != nil {
			return ret, errRet(err)
		}

		ret.Infos[id] = info
	}

	return ret, nil
}

func (msg GossipPacket) ToWire() GossipPacketWire {
	var p *PollPacketWire = nil
	if msg.Poll != nil {
		wired := msg.Poll.ToWire()
		p = &wired
	}

	var s *StatusPacketWire = nil
	if msg.Status != nil {
		wired := msg.Status.ToWire()
		s = &wired
	}

	return GossipPacketWire{
		Poll:      p,
		Signature: msg.Signature,
		Status:    s,
	}
}

type GossipPacketWire struct {
	Poll      *PollPacketWire
	Signature *Signature // TODO it can't be optional, can it?
	Status    *StatusPacketWire
}

func (msg GossipPacketWire) ToBase() (GossipPacket, error) {
	ret := GossipPacket{
		Signature: msg.Signature,
	}

	if msg.Poll != nil {
		wire, err := msg.Poll.ToBase()
		if err != nil {
			return ret, errors.New("GossipPacketWire: " + err.Error())
		}
		ret.Poll = &wire
	}

	if msg.Status != nil {
		wire, err := msg.Status.ToBase()
		if err != nil {
			return ret, errors.New("GossipPacketWire: " + err.Error())
		}
		ret.Status = &wire
	}

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
