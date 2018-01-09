package main

import (
	"flag"
	"strconv"
)

type Settings struct {
	Port uint64
}

func (s Settings) getUrl(elems ...string) string {
	url := "http://localhost:" + strconv.FormatUint(s.Port, 10)

	for _, e := range elems {
		url += "/" + e
	}

	return url
}

func main() {
	port := flag.Uint64("UIPort", 10000, "port to connect to")
	flag.Parse()

	s := Settings{
		Port: *port,
	}

	args := flag.Args()
	action := args[0]
	tail := args[1:]

	switch action {
	case "poll":
		poll(s, tail)
	case "key":
		key(s, tail)
	case "vote":
		vote(s, tail)
	default:
		panic("unkown action: " + action)
	}
}
