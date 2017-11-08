package proto


type PeerMessage struct {
	ID	uint32
	Text	string
}

type RumorMessage struct {
	Origin		string
	PeerMessage	PeerMessage
}


type PeerStatus struct {
	Identifier	string
	NextID		uint32
}

type StatusPacket struct {
	Want	[]PeerStatus
}


type GossipPacket struct {
	Rumor	*RumorMessage
	Status	*StatusPacket
}
