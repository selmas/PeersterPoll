package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dedis/protobuf"
	"github.com/gorilla/mux"

	"./proto"
)

type Server struct {
	Addr *net.UDPAddr
	Conn *net.UDPConn
}

type PeerSet struct {
	sync.RWMutex
	Set map[string]bool
}

type MessageSet struct {
	sync.RWMutex
	Set map[string][]proto.PeerMessage
}

type Gossiper struct {
	Name     string
	LastUid  uint32
	Peers    PeerSet
	Messages MessageSet
	Server   *Server
}

func NewServer(address string) *Server {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		Addr: addr,
		Conn: conn,
	}
}

func NewGossiper(name string, server *Server) *Gossiper {
	return &Gossiper{
		Name:   name,
		Server: server,
		Peers: PeerSet{
			Set: make(map[string]bool),
		},
		Messages: MessageSet{
			Set: make(map[string][]proto.PeerMessage),
		},
	}
}

type Dispatcher func(*Gossiper, *net.UDPAddr, proto.GossipPacket)

func runServer(gossiper *Gossiper, server *Server, dispatcher Dispatcher) {
	buf := make([]byte, 1024)

	for {
		bufSize, peerAddr, err := server.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("dropped connection: ", err)
			continue
		}

		var msg proto.GossipPacket
		err = protobuf.Decode(buf[:bufSize], &msg)
		if err != nil {
			log.Printf("unable to decode msg: ", err)
			continue
		}

		go dispatcher(gossiper, peerAddr, msg)
	}
}

func getRandomPeer(peers *PeerSet, butNotThisPeer *net.UDPAddr) *net.UDPAddr {
	peers.RLock()
	defer peers.RUnlock()

	var addr *net.UDPAddr = nil
	for peer := range peers.Set {
		if peer == butNotThisPeer.String() {
			continue
		}

		addr = parseAddr(peer)
	}

	return addr
}

func writeMsgToUDP(server *Server, peer *net.UDPAddr, rumor *proto.RumorMessage, status *proto.StatusPacket) {
	toSend, err := protobuf.Encode(&proto.GossipPacket{
		Rumor:  rumor,
		Status: status,
	})

	if err != nil {
		log.Printf("unable to encode answer: ", err)
		return
	}

	if rumor != nil {
		printMongering(peer)
	}

	server.Conn.WriteToUDP(toSend, peer)
}

func sendRumor(gossiper *Gossiper, msg *proto.RumorMessage, fromPeer *net.UDPAddr) {
	for {
		peer := getRandomPeer(&gossiper.Peers, fromPeer)
		if peer == nil {
			break
		}

		writeMsgToUDP(gossiper.Server, peer, msg, nil)

		printFlippedCoin(peer, "rumor")
		if rand.Intn(2) == 0 {
			break
		}
	}
}

func handleClientMessage(gossiper *Gossiper, _ *net.UDPAddr, pkg proto.GossipPacket) {
	printClientRumor(gossiper, pkg.Rumor)

	newUid := atomic.AddUint32(&gossiper.LastUid, 1)

	msg := &proto.RumorMessage{
		Origin: gossiper.Name,
		PeerMessage: proto.PeerMessage{
			ID:   newUid,
			Text: pkg.Rumor.PeerMessage.Text,
		},
	}

	storeRumor(gossiper, msg)
	sendRumor(gossiper, msg, nil)
}

func peerIsAheadOfUs(gossiper *Gossiper, s *proto.StatusPacket) bool {
	for _, status := range s.Want {
		gossiper.Messages.RLock()
		msgs, found := gossiper.Messages.Set[status.Identifier]
		gossiper.Messages.RUnlock()

		if !found {
			return true
		}

		if uint32(len(msgs)) < (status.NextID - 1) {
			return true
		}
	}

	return false
}

func getWantedRumor(gossiper *Gossiper, s *proto.StatusPacket) *proto.RumorMessage {
	gossiper.Messages.RLock()
	defer gossiper.Messages.RUnlock()

	for origin, msgs := range gossiper.Messages.Set {
		var status *proto.PeerStatus = nil
		for _, current := range s.Want {
			if current.Identifier == origin {
				status = &current
			}
		}

		if status == nil || uint32(len(msgs)) >= status.NextID {
			var pos uint
			if status == nil {
				pos = 0
			} else {
				pos = uint(len(msgs)) - 1
			}

			msg := msgs[pos]
			return &proto.RumorMessage{
				Origin:      origin,
				PeerMessage: msg,
			}
		}
	}

	return nil
}

func getStatus(gossiper *Gossiper) *proto.StatusPacket {
	gossiper.Messages.Lock()
	defer gossiper.Messages.Unlock()

	wanted := make([]proto.PeerStatus, len(gossiper.Messages.Set))
	i := 0
	for origin, msgs := range gossiper.Messages.Set {
		wanted[i] = proto.PeerStatus{
			Identifier: origin,
			NextID:     uint32(len(msgs)) + 1,
		}

		i++
	}

	return &proto.StatusPacket{
		Want: wanted,
	}
}

func syncStatus(gossiper *Gossiper, peer *net.UDPAddr, msg *proto.StatusPacket) {
	rumor := getWantedRumor(gossiper, msg)

	if rumor != nil {
		writeMsgToUDP(gossiper.Server, peer, rumor, nil)
	} else if peerIsAheadOfUs(gossiper, msg) {
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper))
	} else {
		printInSyncWith(peer)
	}
}

func storeRumor(gossiper *Gossiper, rumor *proto.RumorMessage) bool {
	added := false

	gossiper.Messages.Lock()
	defer gossiper.Messages.Unlock()

	msgs := gossiper.Messages.Set[rumor.Origin]

	if rumor.PeerMessage.ID == uint32(len(msgs))+1 {
		msgs = append(msgs, rumor.PeerMessage)
		added = true
		gossiper.Messages.Set[rumor.Origin] = msgs
	}

	return added
}

func handlePeersterMessage(gossiper *Gossiper, peerAddr *net.UDPAddr, pkg proto.GossipPacket) {
	gossiper.Peers.Lock()
	gossiper.Peers.Set[peerAddr.String()] = true
	gossiper.Peers.Unlock()

	if pkg.Rumor != nil {
		printRumor(gossiper, peerAddr, pkg.Rumor)
		added := storeRumor(gossiper, pkg.Rumor)
		if added {
			sendRumor(gossiper, pkg.Rumor, peerAddr)
		}
	} else if pkg.Status != nil {
		printStatus(gossiper, peerAddr, pkg.Status)
		syncStatus(gossiper, peerAddr, pkg.Status)
	} else {
		log.Printf("message without Rumor or Status, drop")
	}
}

func parseAddr(str string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", str)
	if err != nil {
		log.Fatal("invalid peer \"", str, "\": ", err)
	}

	return addr
}

func antiEntropy(gossiper *Gossiper) {
	ticker := time.NewTicker(time.Second)

	for {
		_ = <-ticker.C

		peer := getRandomPeer(&gossiper.Peers, nil)
		if peer == nil {
			continue
		}

		printFlippedCoin(peer, "status")
		writeMsgToUDP(gossiper.Server, peer, nil, getStatus(gossiper))
	}
}

func main() {
	uiPort := flag.String("UIPort", "10000", "port for the client to connect")
	gossipPort := flag.String("gossipPort", "127.0.0.1:5000", "port to connect the gossiper server")
	name := flag.String("name", "nodeA", "server identifier")
	peersStr := flag.String("peers", "127.0.0.1:5001_10.1.1.7:5002", "underscore separated list of peers")
	flag.Parse()

	gossiper := NewGossiper(*name, NewServer(*gossipPort))
	defer gossiper.Server.Conn.Close()

	for _, peer := range strings.Split(*peersStr, "_") {
		if peer == "" {
			continue
		}

		_ = parseAddr(peer)
		gossiper.Peers.Set[peer] = true
	}

	r := mux.NewRouter()
	r.HandleFunc("/message", apiGetMessages(gossiper)).Methods("GET")
	r.HandleFunc("/message", apiPutMessage(gossiper)).Methods("POST")
	r.HandleFunc("/node", apiGetNodes(gossiper)).Methods("GET")
	r.HandleFunc("/node", apiPutNode(gossiper)).Methods("POST")
	r.HandleFunc("/id", apiGetId(gossiper)).Methods("GET")
	r.HandleFunc("/id", apiChangeId(gossiper)).Methods("POST")
	r.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/", r)

	// one should stay main thread'ed to avoid exiting
	go runServer(gossiper, gossiper.Server, handlePeersterMessage)
	go antiEntropy(gossiper)
	http.ListenAndServe(":"+*uiPort, nil)
}
