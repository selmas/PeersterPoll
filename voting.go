package main

// TODO: add option for origin node to sign / commit Poll to guarantee integrity of it
type Poll struct {
	Question      string
	AnswerOptions []string
	OriginNode    string
	ID            uint32
	Participants  []string
}

type Vote struct {


}

type PollMessage struct {
	Poll *Poll
	Vote *Vote
}