// Package ukeire counts the tiles that improve a hand and how many of each
// remain. Which tiles improve the hand comes from the hand's shape; how many
// remain comes from the tile supply; the answer needs both.
package ukeire

import (
	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// Entry is one improving tile and how many of it remain unseen.
type Entry struct {
	Tile      tile.Tile
	Remaining int
}

// Ukeire is the improving tiles of a hand in tile order. For a tenpai hand
// the improving tiles are the winning tiles, so the ukeire are the waits.
type Ukeire struct {
	hand    hand.Hand
	entries []Entry
}

// Of counts the ukeire against a supply. Pass the supply built from
// everything the player can see; OfHand uses only the hand itself.
func Of(h hand.Hand, supply tile.Supply) Ukeire {
	improving := h.ImprovingTiles()
	entries := make([]Entry, 0, len(improving))
	for _, t := range improving {
		entries = append(entries, Entry{Tile: t, Remaining: supply.Remaining(t)})
	}
	return Ukeire{hand: h, entries: entries}
}

// OfHand counts the ukeire when only the hand's own tiles have been seen.
func OfHand(h hand.Hand) Ukeire {
	return Of(h, tile.MustSupply(h.AllTiles()))
}

// Entries lists the improving tiles in tile order.
func (u Ukeire) Entries() []Entry {
	return append([]Entry(nil), u.entries...)
}

// RemainingTotal is the number of improving tiles left unseen; discard
// choices are compared on it.
func (u Ukeire) RemainingTotal() int {
	total := 0
	for _, e := range u.entries {
		total += e.Remaining
	}
	return total
}

// WaitKinds lists the wait shapes when t completes the hand, deduplicated
// in the order the readings are found. A tile that does not complete the
// hand, or a hand that is not tenpai, has none.
func (u Ukeire) WaitKinds(t tile.Tile) []winning.WaitKind {
	seen := map[winning.WaitKind]bool{}
	var out []winning.WaitKind
	for _, f := range winning.Forms(u.hand, t, winning.Situation{}) {
		if k := f.WaitKind(); !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}
