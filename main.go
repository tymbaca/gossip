package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/tymbaca/study/gossip/peer"
)

var (
	peers = map[string]*peer.Peer{}
	mu    = new(sync.RWMutex)
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

	addrs := lo.Keys(peers)
	addrsMap := lo.SliceToMap(addrs, func(addr string) (string, struct{}) { return addr, struct{}{} })
	for key := range peers {
		peers[key].HandleSetPeers("", peer.Gossip[map[string]struct{}]{Val: addrsMap, Time: time.Now()})
		go peers[key].Launch(_updateInterval)
	}

	entry := ChoosePeer()
	entry.HandleSetSheeps(peer.Gossip[int]{Val: 10, Time: time.Now()})

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

	launchWindow(ctx)
}
