package main

import (
	"context"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/samber/lo"
	"github.com/tymbaca/study/gossip/peer"
	"golang.org/x/exp/rand"
)

const (
	_winWidth    = 1200
	_winHeight   = 1000
	_nodeRadius  = 20
	_textSize    = 20
	_addrSize    = 8
	_infoSize    = 6
	_oldestColor = 8 * time.Second
)

func launchWindow(ctx context.Context) {
	rl.InitWindow(_winWidth, _winHeight, "gossip")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawFPS(10, 10)
		rl.DrawText("LMB - Pass random data to node\nRMB - Kill/revive the node\n1 - Toggle names\n2 - Toggle peer lists", 10, 35, _infoSize, rl.Gray)

		mu.Lock()
		positions := CircleLayout(len(peers), float64(rl.Lerp(100, _winHeight, float32(len(peers))/100)), _winWidth/2, _winHeight/2)
		addrs := lo.Keys(peers)
		slices.Sort(addrs)

		// Draw links
		drawLinks(peers, addrs, positions)

		drawNodes(peers, addrs, positions)

		if clicked := getClickedPeer(peers, addrs, positions); clicked != nil {
			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				clicked.ToggleDead()
			}

			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				clicked.HandleSetSheeps(peer.Gossip[int]{Val: rand.Intn(100), Time: time.Now()})
			}
		}

		if rl.IsKeyPressed(rl.KeyEqual) {
			spawnPeer(ctx)
		}

		if rl.IsKeyPressed(rl.KeyMinus) {
			removePeer()
		}

		if rl.IsKeyPressed(rl.KeyOne) {
			_drawAddrs = !_drawAddrs
		}

		if rl.IsKeyPressed(rl.KeyTwo) {
			_drawInfo = !_drawInfo
		}

		mu.Unlock()

		rl.EndDrawing()
	}
}

func getClickedPeer(allPeers map[string]*peer.Peer, addrs []string, positions []Vector2) *peer.Peer {
	if !rl.IsMouseButtonPressed(rl.MouseButtonLeft) && !rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		return nil
	}

	for i, nodePos := range positions {
		if rl.Vector2Distance(rl.GetMousePosition(), rl.Vector2(nodePos)) <= _nodeRadius {
			return allPeers[addrs[i]]
		}
	}

	return nil
}

var (
	_drawAddrs = true
	_drawInfo  = true
)

func drawNodes(allPeers map[string]*peer.Peer, addrs []string, positions []Vector2) {
	for i, addr := range addrs {
		peer := allPeers[addr]
		pos := positions[i]

		rl.DrawCircleV(rl.Vector2(pos), _nodeRadius, getColor(peer))
		rl.DrawText(strconv.Itoa(peer.GetSheeps()), int32(pos.X)-10, int32(pos.Y-10), _textSize, rl.Black)
		hisPeers := peer.GetPeersList()

		if _drawAddrs {
			rl.DrawText(addr, int32(pos.X)+10, int32(pos.Y+20), _addrSize, rl.DarkGreen)
		}

		if _drawInfo {
			rl.DrawText(strings.Join(hisPeers, "\n"), int32(pos.X)+10, int32(pos.Y+25+_addrSize), _infoSize, rl.DarkBrown)
		}
	}
}

func drawLinks(allPeers map[string]*peer.Peer, addrs []string, positions []Vector2) {
	for i, addr := range addrs {
		from := rl.Vector2(positions[i])
		this := allPeers[addr]
		peers := this.GetPeersList()

		for _, peer := range peers {
			if addr == peer {
				continue
			}
			toIdx := slices.Index(addrs, peer)
			if toIdx < 0 {
				continue
			}

			to := rl.Vector2(positions[toIdx])
			rl.DrawLineV(from, to, rl.DarkGray)

			pointFac := (rl.Vector2Distance(from, to) - _nodeRadius - 5) / rl.Vector2Distance(from, to)
			pointPos := rl.Vector2Lerp(from, to, pointFac)
			rl.DrawCircleV(pointPos, 3, rl.DarkGray)
		}
	}
}

func getColor(peer *peer.Peer) rl.Color {
	if !peer.IsAlive() {
		return rl.NewColor(70, 10, 15, 255)
	}
	t := peer.GetSheepsTime()
	oldness := time.Since(t)

	factor := rl.Clamp(float32(oldness)/float32(_oldestColor), 0, 1)
	val := rl.Lerp(20, 255, factor) // from green to red
	return color.RGBA{R: uint8(val), G: 122, B: 122, A: 255}
}

// Vector2 struct represents a 2D vector with x and y coordinates
type Vector2 struct {
	X, Y float32
}

// CircleLayout arranges `n` points evenly spaced in a circle
// and returns their positions as a slice of Vector2
func CircleLayout(n int, radius float64, offsetX, offsetY float32) []Vector2 {
	if n <= 0 {
		return nil // No points to place
	}

	// Slice to hold the calculated positions
	positions := make([]Vector2, n)

	for i := 0; i < n; i++ {
		// Calculate angle for each point
		angle := 2 * math.Pi * float64(i) / float64(n)
		// Calculate x and y coordinates
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		positions[i] = Vector2{X: float32(x) + offsetX, Y: float32(y) + offsetY}
	}

	return positions
}
