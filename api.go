package pollparty

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"encoding/json"
	"strings"
	"fmt"
)

// TODO delete this
type FakeGossiper struct {
	PollQuestion string
	PollOptions []string
	PollResults string
	MyPollVote	string
}

func apiStartPoll(gossiper *FakeGossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		pollInfo := string(buf[:size])
		questionAndOpts := strings.Split(pollInfo, "\n")

		question := questionAndOpts[0]
		options := questionAndOpts[1:]

		fmt.Println("Starting poll \""+question+"\" with options", options)

		// TODO start vote with this
		gossiper.PollQuestion = question
		gossiper.PollOptions = options
	}
}

func apiGetPollOptions(gossiper *FakeGossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pollData := make([]string, 0)

		pollData = append(pollData, gossiper.PollQuestion)
		pollData = append(pollData, gossiper.PollOptions...)

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

func apiVoteForPoll(gossiper *FakeGossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		size, _ := r.Body.Read(buf)
		gossiper.MyPollVote = string(buf[:size])
		fmt.Println("My vote:", gossiper.MyPollVote)
	}
}

func apiGetPollResults(gossiper *FakeGossiper) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		//TODO change to get the actual results
		results := createFakePollResults(gossiper.PollOptions)

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

func apiNewPoll(g *Gossiper) func(http.ResponseWriter, *http.Request) {
	buf := make([]byte, 1024)

	return func(w http.ResponseWriter, r *http.Request) {
		var poll Poll

		size, _ := r.Body.Read(buf)
		err := json.Unmarshal(buf[:size], &poll)
		if err != nil {
			log.Println("unable to decode as Poll")
			return
		}

		log.Println("got new poll")

		id := NewPollKey(g)
		g.RunningPolls.Add(id, MasterHandler(g))
		g.RunningPolls.Send(id, PollPacket{
			ID:   id,
			Poll: &poll,
		})
	}
}

func apiStart(gossiper *Gossiper, uiPort string) {

	// TODO delete this and replace others with gossiper
	fgossiper := &FakeGossiper{}

	r := mux.NewRouter()

	r.HandleFunc("/poll", apiGetPollOptions(fgossiper)).Methods("GET")
	r.HandleFunc("/poll", apiStartPoll(fgossiper)).Methods("POST")

	r.HandleFunc("/vote", apiGetPollResults(fgossiper)).Methods("GET")
	r.HandleFunc("/vote", apiVoteForPoll(fgossiper)).Methods("POST")

  r.HandleFunc("/poll", apiNewPoll(gossiper)).Methods("PUT")

  r.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/", r)

	http.ListenAndServe(":"+uiPort, nil)
}
