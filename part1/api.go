package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync/atomic"

	"./proto"
)

func apiGetMessages(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(formatMessages(gossiper))

		if err != nil {
			log.Printf("unable to encode as json")
			return
		}

		_, err = w.Write(bytes)
		if err != nil {
			log.Printf("unable to send answer")
		}
	}
}

func apiPutMessage(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		newUid := atomic.AddUint32(&gossiper.LastUid, 1)

		size, _ := r.Body.Read(buf)
		text := string(buf[:size])

		msg := &proto.RumorMessage{
			Origin: gossiper.Name,
			PeerMessage: proto.PeerMessage{
				ID:   newUid,
				Text: text,
			},
		}

		storeRumor(gossiper, msg)
		sendRumor(gossiper, msg, nil)
	}
}

func apiGetNodes(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gossiper.Peers.RLock()
		peers := make([]string, len(gossiper.Peers.Set))
		i := 0
		for peer, _ := range gossiper.Peers.Set {
			peers[i] = peer
			i++
		}
		gossiper.Peers.RUnlock()
		sort.Strings(peers)

		bytes, err := json.Marshal(peers)

		if err != nil {
			log.Printf("unable to encode as json")
			return
		}

		_, err = w.Write(bytes)
		if err != nil {
			log.Printf("unable to send answer")
		}
	}
}

func apiPutNode(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		node := string(buf[:size])

		gossiper.Peers.Lock()
		defer gossiper.Peers.Unlock()

		gossiper.Peers.Set[node] = true
	}
}

func apiGetId(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(gossiper.Name))
		if err != nil {
			log.Printf("unable to send answer")
		}
	}
}

func apiChangeId(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		gossiper.Name = string(buf[:size])
	}
}

func formatMessages(gossiper *Gossiper) map[string][]string {
	gossiper.Messages.RLock()
	defer gossiper.Messages.RUnlock()

	messages := make(map[string][]string)
	for origin, peerMessages := range gossiper.Messages.Set {

		msgs := make([]string, len(peerMessages))
		for i, msg := range peerMessages {
			msgs[i] = msg.Text
		}

		messages[origin] = msgs
	}

	return messages
}
