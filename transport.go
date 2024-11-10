package main

import (
	"sync/atomic"
	"time"

	"github.com/tymbaca/study/gossip/nodes"
)

var (
	_reqCount atomic.Int64
	_rps      float64 // i'm gonna throw up
)

var (
	_prevReqCount     int64
	_prevReqCountTime time.Time
)

func countRPS() {
	prev := _reqCount.Load()
	prevTime := time.Now()

	for range time.Tick(1 * time.Second) {
		current := _reqCount.Load()
		delta := float64(current - prev)
		deltaTime := time.Since(prevTime).Seconds()

		_rps = delta / deltaTime

		prev = current
		prevTime = time.Now()
	}
}

type mapTransport struct{}

func (t *mapTransport) SetSheeps(sender string, addr string, sheeps nodes.Sheeps) error {
	_mu.RLock()
	defer _mu.RUnlock()
	_reqCount.Add(1)

	toPeer, ok := _allNodes[addr]
	if !ok {
		return nodes.ErrRemoved
	}

	toPeer.HandleSetSheeps(sheeps)
	return nil
}

func (t *mapTransport) InterchangePeers(sender string, addr string, addrs nodes.PeersList) (nodes.PeersList, error) {
	_mu.RLock()
	defer _mu.RUnlock()
	_reqCount.Add(1)

	toPeer, ok := _allNodes[addr]
	if !ok {
		return nodes.PeersList{}, nodes.ErrRemoved
	}

	return toPeer.HandleInterchangePeers(sender, addrs)
}
