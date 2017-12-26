package main

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/Peerster/proto"
)

func updateRouting(gossiper *Gossiper, peerAddr *net.UDPAddr, rumor *proto.RumorMessage, newerRoute bool) {
	if rumor.Origin == gossiper.Name {
		return
	}

	gossiper.Routes.Lock()
	defer gossiper.Routes.Unlock()

	if !newerRoute {
		oldRoute, found := gossiper.Routes.Table[rumor.Origin]
		if found && oldRoute.IsDirect {
			return
		}
	}

	gossiper.Routes.Table[rumor.Origin] = Route{
		IsDirect: rumor.IsDirect(),
		Addr:     *peerAddr,
	}
}

func getNextHop(gossiper *Gossiper, origin string) *net.UDPAddr {
	gossiper.Routes.RLock()
	defer gossiper.Routes.RUnlock()

	nextHop, found := gossiper.Routes.Table[origin]

	if found {
		return &nextHop.Addr
	}

	return nil
}

func antiEntropyRouting(gossiper *Gossiper, routingTimeout uint) {
	ticker := time.NewTicker(time.Duration(routingTimeout) * time.Second)

	for {
		peer := getRandomPeer(&gossiper.Peers, nil)
		if peer == nil {
			_ = <-ticker.C
			continue
		}

		newUid := atomic.AddUint32(&gossiper.LastUid, 1)

		msg := proto.RumorMessage{
			PeerMessage: proto.PeerMessage{
				Origin: gossiper.Name,
				ID:     newUid,
				Text:   "",
			},
		}

		storeRumor(gossiper, &msg)
		sendRumor(gossiper, &msg, nil)

		_ = <-ticker.C
	}
}
