package ukeire_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
	"github.com/okki-0417/mahjong/winning"
)

func labelsOf(u ukeire.Ukeire) string {
	var out []string
	for _, e := range u.Entries() {
		out = append(out, e.Tile.String())
	}
	return strings.Join(out, " ")
}

func remainingOf(u ukeire.Ukeire) map[string]int {
	out := map[string]int{}
	for _, e := range u.Entries() {
		out[e.Tile.String()] = e.Remaining
	}
	return out
}

func with(closed, seen string, melds ...hand.Meld) ukeire.Ukeire {
	h := mt.Hand(closed, melds...)
	return ukeire.Of(h, tile.MustSupply(append(h.AllTiles(), mt.Tiles(seen)...)))
}

func TestOfHand(t *testing.T) {
	t.Run("a ryanmen wait has two entries with four remaining each", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z"))
		if labelsOf(u) != "3s 6s" || remainingOf(u)["3s"] != 4 || remainingOf(u)["6s"] != 4 {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("a kanchan wait has one entry", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 3s 5z 5z"))
		if labelsOf(u) != "2s" || remainingOf(u)["2s"] != 4 {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("a shanpon wait has two entries with two remaining each", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 2m 3m 4m 5m 6m 7p 8p 9p 5z 5z 9s 9s"))
		if labelsOf(u) != "9s 5z" || remainingOf(u)["5z"] != 2 || remainingOf(u)["9s"] != 2 {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("the thirteen-sided kokushi wait lists all thirteen kinds", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"))
		if labelsOf(u) != "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z" {
			t.Fatalf("got %q", labelsOf(u))
		}
		for _, e := range u.Entries() {
			if e.Remaining != 3 {
				t.Errorf("%v: %d", e.Tile, e.Remaining)
			}
		}
	})
	t.Run("a chiitoitsu tenpai has one entry", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z"))
		if labelsOf(u) != "7z" || remainingOf(u)["7z"] != 3 {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("every entry of a 1-shanten hand lowers the shanten", func(t *testing.T) {
		h := mt.Hand("1m 2m 3m 4m 5m 6m 7p 8p 1s 2s 5z 5z 9s")
		u := ukeire.OfHand(h)
		if h.Shanten() != 1 || len(u.Entries()) == 0 {
			t.Fatalf("shanten %d entries %d", h.Shanten(), len(u.Entries()))
		}
		for _, e := range u.Entries() {
			if got := hand.ShantenOf(append(h.ClosedTiles(), e.Tile), nil); got != 0 {
				t.Errorf("%v: shanten %d", e.Tile, got)
			}
		}
	})
	t.Run("counts meld tiles as seen", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("2m 3m 4m 5m 6m 7m 3p 4p 5p 5s", mt.Pon("1z 1z 1z")))
		for _, e := range u.Entries() {
			if e.Tile == tile.East && e.Remaining != 1 {
				t.Fatalf("east remaining %d", e.Remaining)
			}
		}
	})
	t.Run("sorts entries in tile order", func(t *testing.T) {
		u := ukeire.OfHand(mt.Hand("1m 3m 5m 7m 9m 1p 3p 5p 7p 9p 1s 3s 5s"))
		labels := strings.Fields(labelsOf(u))
		tiles, _ := tile.ParseAll(labels)
		if !reflect.DeepEqual(tiles, tile.Sorted(tiles)) {
			t.Fatalf("got %v", labels)
		}
	})
}

func TestOf(t *testing.T) {
	t.Run("subtracts tiles seen elsewhere", func(t *testing.T) {
		u := with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "3s 3s")
		if remainingOf(u)["3s"] != 2 || remainingOf(u)["6s"] != 4 {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("keeps a wait whose four copies are all seen, at zero", func(t *testing.T) {
		u := with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "3s 3s 3s 3s")
		if remainingOf(u)["3s"] != 0 || !strings.Contains(labelsOf(u), "3s") {
			t.Fatalf("got %+v", u.Entries())
		}
	})
	t.Run("does not change which tiles improve the hand", func(t *testing.T) {
		if labelsOf(with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "3s 3s")) != labelsOf(with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "")) {
			t.Fatal("tiles changed")
		}
	})
	t.Run("RemainingTotal sums the remaining copies", func(t *testing.T) {
		if with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "").RemainingTotal() != 8 ||
			with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "3s 3s").RemainingTotal() != 6 {
			t.Fatal("total")
		}
	})
	t.Run("Entries returns a copy", func(t *testing.T) {
		u := with("1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "")
		u.Entries()[0].Remaining = 99
		if u.Entries()[0].Remaining == 99 {
			t.Fatal("shared")
		}
	})
}

func TestWaitKinds(t *testing.T) {
	cases := []struct {
		name   string
		closed string
		tile   string
		want   []winning.WaitKind
	}{
		{"ryanmen", "1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "3s", []winning.WaitKind{winning.Ryanmen}},
		{"kanchan", "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 3s 5z 5z", "2s", []winning.WaitKind{winning.Kanchan}},
		{"penchan", "1m 2m 3m 4m 5m 6m 7p 8p 9p 8s 9s 5z 5z", "7s", []winning.WaitKind{winning.Penchan}},
		{"shanpon", "1m 2m 3m 4m 5m 6m 7p 8p 9p 5z 5z 9s 9s", "5z", []winning.WaitKind{winning.Shanpon}},
		{"tanki", "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 1z 9m", "9m", []winning.WaitKind{winning.Tanki}},
		{"a tile that does not complete the hand has none", "1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z", "9s", nil},
		{"a hand that is not tenpai has none", "1m 3m 5m 7m 9m 1p 3p 5p 7p 9p 1s 3s 5s", "2m", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ukeire.OfHand(mt.Hand(c.closed)).WaitKinds(mt.T(c.tile))
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
	t.Run("lists each reading of a multi-sided wait once", func(t *testing.T) {
		got := ukeire.OfHand(mt.Hand("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m")).WaitKinds(tile.M3)
		if len(got) != 2 {
			t.Fatalf("got %v", got)
		}
	})
}
