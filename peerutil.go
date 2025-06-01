package main

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/tymbaca/gossip/logger"
	"github.com/tymbaca/gossip/nodes"
)

func ChoosePeer() *nodes.Node {
	_mu.RLock()
	defer _mu.RUnlock()

	return choosePeer()
}

func choosePeer() *nodes.Node {
	for _, peer := range _allNodes {
		return peer
	}

	return nil
}

func SpawnPeer(ctx context.Context) {
	_mu.Lock()
	defer _mu.Unlock()

	spawnPeer(ctx)
}

func spawnPeer(ctx context.Context) {
	newAddr := gofakeit.IPv4Address()
	// retrying until get unique value
	// (if we just replace with non-unique value - old node with that name
	// will run infinitly in background)
	for {
		_, ok := _allNodes[newAddr]
		if !ok {
			break
		}

		logger.Warnf("%s addr already exists, trying againg", newAddr)
		newAddr = gofakeit.Noun()
	}

	randomPeer := choosePeer()
	if randomPeer != nil {
		randomPeer.AddPeer("", newAddr)
	}

	newPeer := nodes.New(ctx, newAddr, &mapTransport{})

	_allNodes[newAddr] = newPeer
	go newPeer.Launch(_updateInterval)
}

// func RemovePeer(allNodes map[string]*nodes.Node, addrs []string) {
// 	_mu.Lock()
// 	defer _mu.Unlock()
//
// 	removePeer(allNodes, addrs)
// }

func removePeer(allNodes map[string]*nodes.Node, addrs []string) {
	for _, addr := range addrs {
		peer := allNodes[addr]
		delete(_allNodes, addr)
		peer.Stop()
		break
	}
}
