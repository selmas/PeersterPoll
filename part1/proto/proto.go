package proto

const CLIENT_TAG string = "from client"

type Message struct {
	OPName	*string
	Relay	*string
	Text	string
}
