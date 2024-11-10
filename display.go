package main

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"slices"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/samber/lo"
	"github.com/tymbaca/study/gossip/nodes"
	"golang.org/x/exp/rand"
)

const (
	_winWidth    = 800
	_winHeight   = 600
	_nodeRadius  = 20
	_textSize    = 20
	_addrSize    = 8
	_infoSize    = 6
	_oldestColor = 8 * time.Second
)

func launchWindow(ctx context.Context) {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(_winWidth, _winHeight, "gossip")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		// rl.DrawFPS(10, 10)
		rl.DrawText("LMB - Pass random data to node\nRMB - Kill/revive the node\n'1' - Toggle names\n'2' - Toggle peer lists\n'=' - Add new node\n'-' - Remove random node", 10, 35, _infoSize, rl.Gray)
		rl.DrawText(fmt.Sprintf("Request count: %d (RPS: %0.2f)", _reqCount.Load(), _rps), 10, 10, _infoSize, rl.Gray)

		// TODO window resize - get window size
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
				clicked.HandleSetSheeps(nodes.Gossip[int]{Val: rand.Intn(100), UpdateTime: time.Now()})
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

func getClickedPeer(allPeers map[string]*nodes.Node, addrs []string, positions []Vector2) *nodes.Node {
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

func drawNodes(allPeers map[string]*nodes.Node, addrs []string, positions []Vector2) {
	for i, addr := range addrs {
		peer := allPeers[addr]
		pos := positions[i]

		rl.DrawCircleV(rl.Vector2(pos), _nodeRadius, getColor(peer))
		rl.DrawText(strconv.Itoa(peer.GetSheeps()), int32(pos.X)-10, int32(pos.Y-10), _textSize, rl.Black)
		hisPeers := peer.GetPeers()
		hisPeersList := peer.GetPeersList()

		if _drawAddrs {
			rl.DrawText(addr, int32(pos.X)+10, int32(pos.Y+20), _addrSize, rl.Green)
		}
		if _drawInfo {
			y := int32(pos.Y + 25 + _addrSize)
			for _, addr := range hisPeersList {
				peer, ok := hisPeers[addr]
				if !ok {
					continue
				}

				if !peer.Val.Removed {
					rl.DrawText(addr, int32(pos.X)+10, y, _infoSize, rl.DarkGreen)
				} else {
					rl.DrawText(addr, int32(pos.X)+10, y, _infoSize, rl.DarkBrown)
				}
				y += _infoSize + 3
			}
		}
	}
}

func drawLinks(allPeers map[string]*nodes.Node, addrs []string, positions []Vector2) {
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

func getColor(peer *nodes.Node) rl.Color {
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
