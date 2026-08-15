package ukeire_test

import (
	"testing"

	"github.com/okki-0417/mahjong/internal/goldentest"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
)

type goldenEntry struct {
	Tile      string   `json:"tile"`
	Remaining int      `json:"remaining"`
	WaitKinds []string `json:"wait_kinds"`
}

type goldenUkeire struct {
	Closed         []string          `json:"closed"`
	Melds          []goldentest.Meld `json:"melds"`
	Seen           []string          `json:"seen"`
	Entries        []goldenEntry     `json:"entries"`
	RemainingTotal int               `json:"remaining_total"`
}

// TestGolden replays ukeire recorded from the Ruby implementation.
func TestGolden(t *testing.T) {
	goldentest.Each(t, "ukeire.jsonl", func(line int, g goldenUkeire) {
		h := goldentest.Hand(t, g.Closed, g.Melds)
		seen, err := tile.ParseAll(g.Seen)
		if err != nil {
			t.Fatal(err)
		}
		u := ukeire.Of(h, tile.MustSupply(seen))
		entries := u.Entries()
		if len(entries) != len(g.Entries) {
			t.Errorf("line %d %v: %d entries, want %d", line, h, len(entries), len(g.Entries))
			return
		}
		for i, e := range entries {
			want := g.Entries[i]
			if e.Tile.String() != want.Tile || e.Remaining != want.Remaining {
				t.Errorf("line %d %v: entry %d = %v/%d, want %s/%d", line, h, i, e.Tile, e.Remaining, want.Tile, want.Remaining)
			}
			var kinds []string
			for _, k := range u.WaitKinds(e.Tile) {
				kinds = append(kinds, k.String())
			}
			if !goldentest.SameLabels(kinds, want.WaitKinds) {
				t.Errorf("line %d %v: wait kinds of %v = %v, want %v", line, h, e.Tile, kinds, want.WaitKinds)
			}
		}
		if u.RemainingTotal() != g.RemainingTotal {
			t.Errorf("line %d %v: remaining total %d, want %d", line, h, u.RemainingTotal(), g.RemainingTotal)
		}
	})
}
