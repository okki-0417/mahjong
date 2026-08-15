package hand_test

import (
	"testing"

	"github.com/okki-0417/mahjong/internal/goldentest"
	"github.com/okki-0417/mahjong/tile"
)

type goldenHand struct {
	Closed    []string          `json:"closed"`
	Melds     []goldentest.Meld `json:"melds"`
	Shanten   int               `json:"shanten"`
	Improving []string          `json:"improving"`
	Waits     []string          `json:"waits"`
}

// TestGoldenShanten replays hands recorded from the Ruby implementation and
// checks that shanten, improving tiles, and waits agree.
func TestGoldenShanten(t *testing.T) {
	goldentest.Each(t, "hand_shanten.jsonl", func(line int, g goldenHand) {
		h := goldentest.Hand(t, g.Closed, g.Melds)
		if got := h.Shanten(); int(got) != g.Shanten {
			t.Errorf("line %d %v: shanten %d, want %d", line, h, got, g.Shanten)
		}
		if got := tile.Labels(h.ImprovingTiles()); !goldentest.SameLabels(got, g.Improving) {
			t.Errorf("line %d %v: improving %v, want %v", line, h, got, g.Improving)
		}
		if got := tile.Labels(h.Waits()); !goldentest.SameLabels(got, g.Waits) {
			t.Errorf("line %d %v: waits %v, want %v", line, h, got, g.Waits)
		}
	})
}
