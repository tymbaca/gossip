package main

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/tymbaca/study/gossip/peer"
)

func ChoosePeer() *peer.Peer {
	mu.RLock()
	defer mu.RUnlock()

	return choosePeer()
}

func choosePeer() *peer.Peer {
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

	newPeer := peer.New(ctx, newAddr, &mapTransport{})
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

type mapTransport struct{}

func (t *mapTransport) SetSheeps(sender string, addr string, sheeps peer.Sheeps) error {
	mu.RLock()
	defer mu.RUnlock()

	toPeer, ok := peers[addr]
	if !ok {
		return peer.ErrRemoved
	}

	toPeer.HandleSetSheeps(sheeps)
	return nil
}

func (t *mapTransport) SetPeers(sender string, addr string, addrs peer.PeersList) error {
	mu.RLock()
	defer mu.RUnlock()

	toPeer, ok := peers[addr]
	if !ok {
		return peer.ErrRemoved
	}

	toPeer.HandleSetPeers(sender, addrs)
	return nil
}
