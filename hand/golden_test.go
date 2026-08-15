package hand_test

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

type goldenMeld struct {
	Kind  string   `json:"kind"`
	Tiles []string `json:"tiles"`
}

type goldenHand struct {
	Closed    []string     `json:"closed"`
	Melds     []goldenMeld `json:"melds"`
	Shanten   int          `json:"shanten"`
	Improving []string     `json:"improving"`
	Waits     []string     `json:"waits"`
}

func (g goldenHand) hand(t *testing.T) hand.Hand {
	t.Helper()
	closed, err := tile.ParseAll(g.Closed)
	if err != nil {
		t.Fatal(err)
	}
	melds := make([]hand.Meld, 0, len(g.Melds))
	for _, m := range g.Melds {
		kind, err := hand.ParseMeldKind(m.Kind)
		if err != nil {
			t.Fatal(err)
		}
		tiles, err := tile.ParseAll(m.Tiles)
		if err != nil {
			t.Fatal(err)
		}
		melds = append(melds, hand.MustMeld(kind, tiles))
	}
	return hand.Must(closed, melds)
}

// TestGoldenShanten replays hands recorded from the Ruby implementation
// (testdata/hand_shanten.jsonl, written by mahjong-yaritai's
// api/gems/mahjong/bin/export_golden) and checks that shanten, improving
// tiles, and waits agree.
func TestGoldenShanten(t *testing.T) {
	f, err := os.Open("../testdata/hand_shanten.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		var g goldenHand
		if err := json.Unmarshal(scanner.Bytes(), &g); err != nil {
			t.Fatalf("line %d: %v", line, err)
		}
		h := g.hand(t)
		if got := h.Shanten(); int(got) != g.Shanten {
			t.Errorf("line %d %v: shanten %d, want %d", line, h, got, g.Shanten)
		}
		if got := tile.Labels(h.ImprovingTiles()); !sameLabels(got, g.Improving) {
			t.Errorf("line %d %v: improving %v, want %v", line, h, got, g.Improving)
		}
		if got := tile.Labels(h.Waits()); !sameLabels(got, g.Waits) {
			t.Errorf("line %d %v: waits %v, want %v", line, h, got, g.Waits)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if line == 0 {
		t.Fatal("no golden records")
	}
}

func sameLabels(got, want []string) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}
