package main

// TODO: add option for origin node needs to sign / commit Poll to guarantee integrity of it
type Poll struct {
	Question      string
	AnswerOptions []string
	OriginNode    string
	ID            uint32
	Participants  []string
}