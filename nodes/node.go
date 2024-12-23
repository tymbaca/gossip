package nodes

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/tymbaca/gossip/logger"
	"golang.org/x/exp/slices"
)

var (
	ErrRemoved = fmt.Errorf("node is removed")
	ErrDown    = fmt.Errorf("node is temporarily down")
)

const (
	_removedTTL = 10 * time.Second
)

type Node struct {
	me     string
	ctx    context.Context
	cancel context.CancelFunc

	dead bool

	mu        sync.RWMutex
	sheeps    Sheeps
	peers     PeersList
	transport transport
}

func New(ctx context.Context, addr string, transport transport) *Node {
	ctx, cancel := context.WithCancel(ctx)
	return &Node{
		me:        addr,
		ctx:       ctx,
		cancel:    cancel,
		transport: transport,
		peers: map[string]Gossip[Peer]{
			addr: {
				Val:        Peer{Addr: addr},
				UpdateTime: time.Now(),
			},
		},
	}
}

type transport interface {
	SetSheeps(sender string, peer string, sheeps Sheeps) error
	InterchangePeers(sender string, peer string, peers PeersList) (PeersList, error)
}

func (n *Node) Launch(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	go n.clean(n.ctx) // FIX: see notes.md

	go func() {
		// ensure that we have our addr in peers list
		for range time.Tick(1 * time.Second) {
			n.mu.Lock()
			n.peers[n.me] = Gossip[Peer]{Val: Peer{Addr: n.me}, UpdateTime: time.Now()}
			n.mu.Unlock()
		}
	}()

	for {
		for addr, peer := range n.GetPeers() {
			n.mu.RLock()
			dead := n.dead
			me := n.me
			n.mu.RUnlock()

			if dead {
				<-t.C
				continue
			}
			if addr == me {
				continue
			}

			if peer.Val.Removed {
				continue
			}

			if n.ctx.Err() != nil {
				return
			}

			if err := n.transport.SetSheeps(n.me, addr, n.GetSheeps()); err != nil {
				logger.Errorf("can't interchange sheeps: %s", err)
				n.MarkRemoved(addr) // maybe we need few retries before this happens
				continue
			}

			hisPeers, err := n.transport.InterchangePeers(n.me, addr, n.GetPeers())
			if err != nil {
				logger.Errorf("can't interchange peers: %s", err)
				n.MarkRemoved(addr)
				continue
			}

			n.mu.Lock()
			{
				n.updatePeers(hisPeers)
				// we know that this peer is alive - so we update it's gossip time
				peer.UpdateTime = time.Now()
				n.peers[addr] = peer
			}
			n.mu.Unlock()

			<-t.C
		}
	}
}

func (n *Node) Stop() {
	n.cancel()
}

func (n *Node) Kill(dur time.Duration) {
	go func() {
		n.mu.Lock()
		n.dead = true
		n.mu.Unlock()

		<-time.After(dur)

		n.mu.Lock()
		n.dead = false
		n.mu.Unlock()
	}()
}

func (n *Node) ToggleDead() {
	n.mu.Lock()
	n.dead = !n.dead
	n.mu.Unlock()
}

func (n *Node) IsAlive() bool {
	return !n.dead
}

func (n *Node) Addr() string {
	return n.me
}

func (n *Node) HandleSetSheeps(newSheeps Sheeps) error {
	if n.dead {
		return ErrDown
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.sheeps.UpdateTime.After(newSheeps.UpdateTime) {
		return nil
	}

	n.sheeps = newSheeps
	return nil
}

func (n *Node) HandleInterchangePeers(sender string, newPeers PeersList) (PeersList, error) {
	if n.dead {
		return PeersList{}, ErrDown
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// bad
	// if sender != "" {
	// 	defer func() {
	// 		p.peers.Val[sender] = struct{}{}
	// 		p.peers.Time = time.Now()
	// 	}()
	// }

	// if p.peers.Time.After(newPeers.Time) {
	// 	return p.peers, nil // TODO: copy the internal map
	// }
	n.updatePeers(newPeers)

	return n.getPeersMap(), nil
}

func (n *Node) updatePeers(newPeers PeersList) {
	for addr, peer := range newPeers {
		// if i get peer with Removed == true and this node doesn't have
		// have that peer - we don't add it to our peer list
		_, ok := n.peers[addr]
		if peer.Val.Removed && !ok {
			continue
		}

		if n.peers[addr].UpdateTime.Before(peer.UpdateTime) {
			n.peers[addr] = peer
		}
	}
}

func (n *Node) AddPeer(sender, peer string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.peers[peer] = Gossip[Peer]{
		Val:        Peer{Addr: peer},
		UpdateTime: time.Now(),
	}
}

func (n *Node) MarkRemoved(addr string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	peer, ok := n.peers[addr]
	if !ok {
		return
	}

	// // we don't need to update it if it's already removed
	// // otherwise we will have dangling removed that will not reach their TTL
	// // and will not be deleted from peer list
	// if peer.Val.Removed {
	// 	return
	// }

	peer.Val.Removed = true
	peer.UpdateTime = time.Now()
	n.peers[addr] = peer
}

func (n *Node) GetSheeps() Gossip[int] {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.sheeps
}

func (n *Node) GetPeers() map[string]Gossip[Peer] {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.getPeersMap()
}

func (n *Node) getPeersMap() map[string]Gossip[Peer] {
	m := make(map[string]Gossip[Peer], len(n.peers))
	for addr, peer := range n.peers {
		m[addr] = peer
	}

	return m
}

func (n *Node) GetPeersList() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.getPeersList()
}

func (n *Node) getPeersList() []string {
	list := make([]string, 0, len(n.peers))
	for addr := range n.peers {
		// if !peer.Val.Removed {
		list = append(list, addr)
		// }
	}

	slices.Sort(list)

	return list
}

func (n *Node) GetSheepsTime() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.sheeps.UpdateTime
}

type (
	Sheeps    = Gossip[int]
	PeersList = map[string]Gossip[Peer]
)

type Peer struct {
	Addr string
	// Down    bool
	Removed bool
}

type Gossip[T any] struct {
	Val        T
	UpdateTime time.Time
}

func (n *Node) getRandomPeer() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, addr := range lo.Shuffle(n.getPeersList()) {
		if addr != n.me {
			return addr
		}
	}

	return n.me
}

func (n *Node) clean(ctx context.Context) {
	// TODO: ctx context.Context
	for range time.Tick(10 * time.Second) {
		if ctx.Err() != nil {
			return
		}
		n.mu.Lock()
		n.cleanOldRemovedPeers()
		n.mu.Unlock()
	}
}

func (n *Node) cleanOldRemovedPeers() {
	for addr, peer := range n.peers {
		if peer.Val.Removed && time.Since(peer.UpdateTime) > _removedTTL {
			delete(n.peers, addr)
		}
	}
}

func toBytes[T any](val T) []byte {
	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(val)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}
