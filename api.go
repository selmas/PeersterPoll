package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync/atomic"

	"github.com/gorilla/mux"

	"github.com/ValerianRousset/Peerster/proto"
)

type ApiPrivateMessage struct {
	Dest string
	Text string
}

func apiGetMessages(ms *MessageSet) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(formatMessageSet(ms))

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
			PeerMessage: proto.PeerMessage{
				Origin: gossiper.Name,
				ID:     newUid,
				Text:   text,
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

func apiGetRoutes(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gossiper.Routes.RLock()
		defer gossiper.Routes.RUnlock()

		peers := make([]string, len(gossiper.Routes.Table))
		i := 0
		for origin, _ := range gossiper.Routes.Table {
			peers[i] = origin
			i++
		}
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

func apiPutPrivateMessage(gossiper *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		var pm ApiPrivateMessage

		size, _ := r.Body.Read(buf)
		err := json.Unmarshal(buf[:size], &pm)
		if err != nil {
			log.Printf("unable to decode as ApiPrivateMessage")
			return
		}

		msg := &proto.PrivateMessage{
			PeerMessage: proto.PeerMessage{
				Origin: gossiper.Name,
				ID:     0,
				Text:   pm.Text,
			},
			Dest:     pm.Dest,
			HopLimit: 10,
		}

		sendPrivateMessage(gossiper, nil, msg)
	}
}

func formatMessageSet(ms *MessageSet) map[string][]string {
	ms.RLock()
	defer ms.RUnlock()

	messages := make(map[string][]string)
	for origin, peerMessages := range ms.Set {

		msgs := make([]string, 0, len(peerMessages))
		for _, msg := range peerMessages {
			if msg.Text != "" {
				msgs = append(msgs, msg.Text)
			}
		}

		if len(msgs) > 0 {
			messages[origin] = msgs
		}
	}

	return messages
}

func apiStart(gossiper *Gossiper, uiPort string) {
	r := mux.NewRouter()
	r.HandleFunc("/message", apiGetMessages(&gossiper.Messages)).Methods("GET")
	r.HandleFunc("/message", apiPutMessage(gossiper)).Methods("POST")
	r.HandleFunc("/node", apiGetNodes(gossiper)).Methods("GET")
	r.HandleFunc("/node", apiPutNode(gossiper)).Methods("POST")
	r.HandleFunc("/id", apiGetId(gossiper)).Methods("GET")
	r.HandleFunc("/id", apiChangeId(gossiper)).Methods("POST")
	r.HandleFunc("/routes", apiGetRoutes(gossiper)).Methods("GET")
	r.HandleFunc("/private_messages", apiGetMessages(&gossiper.PrivateMessages)).Methods("GET")
	r.HandleFunc("/private_message", apiPutPrivateMessage(gossiper)).Methods("POST")
	r.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/", r)

	http.ListenAndServe(":"+uiPort, nil)
}
