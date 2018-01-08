package pollparty

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

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

func ApiStart(gossiper *Gossiper, uiPort string) {
	r := mux.NewRouter()
	r.HandleFunc("/id", apiGetId(gossiper)).Methods("GET")
	r.HandleFunc("/id", apiChangeId(gossiper)).Methods("POST")
	r.HandleFunc("/poll", apiNewPoll(gossiper)).Methods("PUT")
	r.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/", r)

	http.ListenAndServe(":"+uiPort, nil)
}
