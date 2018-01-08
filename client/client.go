package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
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

func poll_new(s Settings, args []string) {
	url := s.getUrl("poll")
	// TODO hardcoded for now
	msg := pkg.Poll{
		Question: "What's the time?",
		Options: []string{
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

	var client http.Client
	req, err := http.NewRequest("PUT", url, bytes.NewReader(json))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}

func key_new(s Settings, args []string) {
	origin := args[0]

	keys, err := pkg.KeyFileLoad()
	if err != nil {
		log.Fatal(err)
	}

	k, err := ecdsa.GenerateKey(pkg.Curve(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	keys = append(keys, k.PublicKey)

	err = pkg.PrivateKeySave(pkg.PrivateKeyFileName(origin), *k)
	if err != nil {
		log.Fatal(err)
	}

	err = pkg.KeyFileSave(keys)
	if err != nil {
		log.Fatal(err)
	}
}

func poll(s Settings, args []string) {
	action := args[0]

	switch action {
	case "new":
		poll_new(s, args[1:])
	default:
		panic("unkown poll action: " + action)
	}
}

func key(s Settings, args []string) {
	action := args[0]

	switch action {
	case "new":
		key_new(s, args[1:])
	default:
		panic("unkown poll action: " + action)
	}
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
	default:
		panic("unkown action: " + action)
	}
}
