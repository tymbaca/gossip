package main

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/tymbaca/study/gossip/nodes"
)

func ChoosePeer() *nodes.Node {
	mu.RLock()
	defer mu.RUnlock()

	return choosePeer()
}

func choosePeer() *nodes.Node {
	for _, peer := range peers {
		return peer
	}

	return nil
}

func SpawnPeer(ctx context.Context) {
	mu.Lock()
	defer mu.Unlock()

	spawnPeer(ctx)
}

func spawnPeer(ctx context.Context) {
	newAddr := gofakeit.Noun()

	randomPeer := choosePeer()
	if randomPeer != nil {
		randomPeer.AddPeer("", newAddr)
	}

	newPeer := nodes.New(ctx, newAddr, &mapTransport{})
	peers[newAddr] = newPeer
	go newPeer.Launch(_updateInterval)
}

func RemovePeer() {
	mu.Lock()
	defer mu.Unlock()

	removePeer()
}

func removePeer() {
	for addr, peer := range peers {
		delete(peers, addr)
		peer.Stop()
		break
	}
}
