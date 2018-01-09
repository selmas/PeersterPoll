package main

import (
	"bytes"
	"net/http"
)

func vote_put(s Settings, args []string) {
	id := args[0]
	url := s.getUrl("vote", id)
	option := args[1]

	toSend := bytes.NewBufferString(option)

	resp, err := http.Post(url, "text/plain", toSend)
	check(err)
	defer resp.Body.Close()

	checkResp(resp)
}

func vote(s Settings, args []string) {
	action := args[0]
	tail := args[1:]

	switch action {
	case "put":
		vote_put(s, tail)
	default:
		panic("unkown vote action: " + action)
	}
}
