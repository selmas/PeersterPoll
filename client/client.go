package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"strconv"
)

type Settings struct {
	Port uint64
}

func (s Settings) getUrl(post string) string {
	return "http://localhost:" + strconv.FormatUint(s.Port, 10) + "/" + post
}

func propose(s Settings, args []string) {
	url := s.getUrl("propose")
	msg := bytes.NewBufferString("TODO")

	resp, err := http.Post(url, "text/plain", msg)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func main() {
	port := flag.Uint64("UIPort", 10000, "port to connect to")
	action := flag.String("action", "propose", "what to do")
	flag.Parse()

	s := Settings{
		Port: *port,
	}

	switch *action {
	case "propose":
		propose(s, flag.Args())
	}
}
