package pollparty

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// TODO delete this
type FakeGossiper struct {
	PollQuestion string
	PollOptions  []string
	PollResults  string
	MyPollVote   string
}

func apiStartPoll(g *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		pollInfo := string(buf[:size])
		questionAndOpts := strings.Split(pollInfo, "\n")

		question := questionAndOpts[0]
		options := questionAndOpts[1:]

		log.Println("Starting poll \""+question+"\" with options", options)

		id := NewPollKey(g)
		g.RunningPolls.Add(id, MasterHandler(g))
		g.RunningPolls.Send(id, PollPacket{
			ID: id,
			Poll: &Poll{
				Question:  question,
				Options:   options,
				StartTime: time.Now(),                     // TODO user customizable
				Duration:  time.Duration(1 * time.Minute), // TODO user customizable
			},
		})
	}
}

func apiGetPollOptions(g *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := PollKeyFromString(mux.Vars(r)["id"])
		if err != nil {
			log.Println(err)
			return
		}

		info := g.Polls.Get(id)

		pollData := make([]string, 0)
		pollData = append(pollData, info.Poll.Question)
		pollData = append(pollData, info.Poll.Options...)

		// this expects a slice with question in first pos and then the options
		bytes, err := json.Marshal(pollData)

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

func apiVoteForPoll(g *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		option := string(buf[:size])

		id, err := PollKeyFromString(mux.Vars(r)["id"])
		if err != nil {
			log.Println(err)
			return
		}

		g.RunningPolls.Get(id).LocalVote <- option

		log.Println("My vote:", option)
	}
}

func apiGetPollResults(g *Gossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := PollKeyFromString(mux.Vars(r)["id"])
		if err != nil {
			log.Println(err)
			return
		}

		info := g.Polls.Get(id)

		//TODO change to get the actual results
		results := createFakePollResults(info.Poll.Options)

		fmt.Println("Results to send to GUI:", results)

		bytes, err := json.Marshal(results)

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

func createFakePollResults(options []string) map[string]int {
	results := make(map[string]int)

	for _, opt := range options {
		results[opt] = len(opt)
	}

	return results
}

func ApiStart(g *Gossiper, uiPort string) {
	r := mux.NewRouter()

	r.HandleFunc("/poll", apiStartPoll(g)).Methods("POST")
	r.HandleFunc("/poll/{id}", apiGetPollOptions(g)).Methods("GET")

	r.HandleFunc("/vote/{id}", apiGetPollResults(g)).Methods("GET")
	r.HandleFunc("/vote/{id}", apiVoteForPoll(g)).Methods("POST")

	r.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/", r)

	http.ListenAndServe(":"+uiPort, nil)
}
