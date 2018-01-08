package pollparty

import (
	"crypto/sha256"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type PollKey struct {
	Origin string
	ID     uint
}

func (msg PollKey) Check() error {
	if msg.Origin == "" {
		log.Println("empty origin") // TODO bad rep
	}

	return nil
}

func (msg PollKey) String() string {
	return msg.Origin + strconv.FormatUint(uint64(msg.ID), 10)
}

// TODO: add option for origin node to sign / commit Question to guarantee integrity of it
type Poll struct {
	Question    string
	VoteOptions []string
	StartTime   time.Time
	Duration    time.Duration // After duration has passed, can no longer participate in votes
}

func (p Poll) IsTooLate() bool {
	return p.StartTime.Add(p.Duration).Before(time.Now())
}

func (msg Poll) Check() error {
	if msg.Question == "" {
		log.Println("no question")
	}

	if len(msg.VoteOptions) == 0 {
		log.Println("no choices")
	}

	return nil // TODO
}

const SaltSize = 20

type Commitment struct {
	Hash [sha256.Size]byte
	Salt [SaltSize]byte
}

func (msg Commitment) Check() error {
	return nil // TODO check hash and salt size
}

func NewCommitment(answer string) Commitment {
	var salt [SaltSize]byte
	rand.Read(salt[:])

	toHash := make([]byte, 0)
	toHash = append(toHash, []byte(answer)[:]...)
	toHash = append(toHash, salt[:]...)

	hash := sha256.Sum256(toHash)

	return Commitment{
		Hash: hash,
		Salt: salt,
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
	Option string
}

func (msg Vote) Check() error {
	// TODO check that option is in list

	return nil
}

type PollPacket struct {
	ID              PollKey
	Poll            *Poll
	Commitment      *Commitment
	PollCommitments *PollCommitments
	Vote            *Vote
}

func (pkg PollPacket) Check() error {
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
		return errors.New("PollPacket: " + err.Error())
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

type GossipPacket struct {
	Poll   *PollPacket
	Status *StatusPacket
}

func (pkg GossipPacket) Check() error {
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
