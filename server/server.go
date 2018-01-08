package main

import (
	"flag"
	"log"
	"strings"

	pkg "github.com/ValerianRousset/Peerster"
)

func main() {
	uiPort := flag.String("UIPort", "10000", "port for the client to connect")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port to connect the gossiper server")
	name := flag.String("name", "nodeA", "server identifier")
	peersStr := flag.String("peers", "127.0.0.1:5001_10.1.1.7:5002", "underscore separated list of peers")
	flag.Parse()

	gossiper, err := pkg.NewGossiper(*name, pkg.NewServer(*gossipAddr))
	if err != nil {
		log.Fatal(err)
	}
	defer gossiper.Server.Conn.Close()

	for _, peer := range strings.Split(*peersStr, "_") {
		if peer == "" {
			continue
		}

		gossiper.Peers.Set[peer] = true
	}

	// one should stay main thread'ed to avoid exiting
	go pkg.RunServer(gossiper, gossiper.Server, pkg.DispatcherPeersterMessage(gossiper))
	go pkg.AntiEntropyGossip(gossiper)
	pkg.ApiStart(gossiper, *uiPort)
}
