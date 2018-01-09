package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func checkResp(r *http.Response) {
	if r.StatusCode != 200 {
		log.Fatalf("HTTP status error, got %d", r.StatusCode)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func poll_new(s Settings, args []string) {
	url := s.getUrl("poll")
	question := args[0]
	options := args[1:]

	var msg string
	msg += question
	msg += "\n"
	msg += strings.Join(options, "\n")

	toSend := bytes.NewBufferString(msg)

	resp, err := http.Post(url, "text/plain", toSend)
	check(err)
	defer resp.Body.Close()

	checkResp(resp)

	content, err := ioutil.ReadAll(resp.Body)
	check(err)

	fmt.Println(string(content))
}

func poll_list(s Settings, args []string) {
	url := s.getUrl("poll")

	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()

	checkResp(resp)

	content, err := ioutil.ReadAll(resp.Body)
	check(err)

	var list []string
	err = json.Unmarshal(content, &list)
	check(err)

	for _, id := range list {
		fmt.Println(id)
	}
}

func poll(s Settings, args []string) {
	action := args[0]
	tail := args[1:]

	switch action {
	case "new":
		poll_new(s, tail)
	case "list":
		poll_list(s, tail)
	default:
		panic("unkown poll action: " + action)
	}
}
