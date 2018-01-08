package main

import (
	"bytes"
	"encoding/json"
	"flag"
	pkg "github.com/ValerianRousset/Peerster"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Settings struct {
	Port uint64
}

func (s Settings) getUrl(post string) string {
	return "http://localhost:" + strconv.FormatUint(s.Port, 10) + "/" + post
}

func propose(s Settings, args []string) {
	url := s.getUrl("poll")
	// TODO hardcoded for now
	msg := pkg.Poll{
		Question: "What's the time?",
		VoteOptions: []string{
			"10:25",
			"10:30",
			"Time is a weird and purely local definition",
		},
		StartTime: time.Now(),
		Duration:  time.Duration(1 * time.Minute),
	}

	json, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.NewRequest("PUT", url, bytes.NewReader(json))
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}

func main() {
	port := flag.Uint64("UIPort", 10000, "port to connect to")
	flag.Parse()

	s := Settings{
		Port: *port,
	}

	args := flag.Args()
	action := args[0]

	switch action {
	case "propose":
		propose(s, args[1:])
	}
}
