package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

func vote_show(s Settings, args []string) {
	id := args[0]
	url := s.getUrl("vote", id)

	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()

	checkResp(resp)

	content, err := ioutil.ReadAll(resp.Body)
	check(err)

	fmt.Println(string(content))
}

func vote(s Settings, args []string) {
	action := args[0]
	tail := args[1:]

	switch action {
	case "put":
		vote_put(s, tail)
	case "show":
		vote_show(s, tail)
	default:
		panic("unkown vote action: " + action)
	}
}
