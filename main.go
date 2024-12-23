package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/tymbaca/gossip/nodes"
)

var (
	_allNodes = map[string]*nodes.Node{}
	_mu       = new(sync.RWMutex)
)

const (
	_updateInterval = 300 * time.Millisecond
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	for range 1 {
		SpawnPeer(ctx)
	}

	// addrs := lo.Keys(peers)
	// addrsMap := lo.SliceToMap(addrs, func(addr string) (string, struct{}) { return addr, struct{}{} })
	// for key := range peers {
	// 	peers[key].HandleInterchangePeers("", nodes.Gossip[map[string]struct{}]{Val: addrsMap, Time: time.Now()})
	// 	go peers[key].Launch(_updateInterval)
	// } // WARN:

	entry := ChoosePeer()
	entry.HandleSetSheeps(nodes.Gossip[int]{Val: 10, UpdateTime: time.Now()})

	// go func() {
	// 	for range time.Tick(1500 * time.Millisecond) {
	// 		entry.SetSheeps(peer.Gossip[int]{Val: rand.Intn(100), Time: time.Now()})
	// 	}
	// }()

	// go func() {
	// 	for range time.Tick(5000 * time.Millisecond) {
	// 		SpawnPeer()
	// 	}
	// }()

	// go func() {
	// 	for range time.Tick(1700 * time.Millisecond) {
	// 		RemovePeer()
	// 	}
	// }()

	//--------------------------------------------------------------------------------------------------

	go countRPS()
	launchWindow(ctx)
}
