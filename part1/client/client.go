package main

import (
	"bytes"
	"log"
	"flag"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Uint64("UIPort", 10000, "port to connect to")
	msgStr := flag.String("msg", "Hello", "message to send")
	flag.Parse()

	url := "http://localhost:" + strconv.FormatUint(*port, 10) + "/message"
	msg := bytes.NewBufferString(*msgStr)

	resp, err := http.Post(url, "text/plain", msg)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
