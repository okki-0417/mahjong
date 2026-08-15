package ukeire

import (
	"errors"
	"fmt"
	"sort"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// ErrDrawnCount is returned when the tiles given as a drawn hand are not
// 14 - 3 per meld.
var ErrDrawnCount = errors.New("ukeire: a drawn hand is 14 tiles less 3 per meld")

// Candidate is one discard from a drawn hand and the hand it leaves.
type Candidate struct {
	Tile    tile.Tile
	Hand    hand.Hand
	Shanten hand.Shanten
	Ukeire  Ukeire
}

// DiscardComparison lines up every discard from a hand plus its drawn tile
// and compares the hands left behind by shanten, then by ukeire. Tiles of
// one kind are one candidate, represented by the plain tile.
type DiscardComparison struct {
	hand       hand.Hand
	tsumo      tile.Tile
	candidates []Candidate
}

// CompareDiscards builds the comparison for a hand and the tile it drew.
func CompareDiscards(h hand.Hand, tsumo tile.Tile) (DiscardComparison, error) {
	if !tsumo.IsValid() {
		return DiscardComparison{}, fmt.Errorf("%w: %v", hand.ErrInvalidTile, tsumo)
	}
	c := DiscardComparison{hand: h, tsumo: tsumo}
	for _, t := range discardableKinds(h, tsumo) {
		after, err := h.Discard(t, tsumo)
		if err != nil {
			return DiscardComparison{}, err
		}
		c.candidates = append(c.candidates, Candidate{Tile: t, Hand: after, Shanten: after.Shanten(), Ukeire: OfHand(after)})
	}
	sort.SliceStable(c.candidates, func(i, j int) bool {
		a, b := c.candidates[i], c.candidates[j]
		if a.Shanten != b.Shanten {
			return a.Shanten < b.Shanten
		}
		return a.Ukeire.RemainingTotal() > b.Ukeire.RemainingTotal()
	})
	return c, nil
}

// CompareDrawnTiles builds the comparison from a drawn hand laid out as one
// list, taking the last tile as the draw.
func CompareDrawnTiles(tiles []tile.Tile, melds []hand.Meld) (DiscardComparison, error) {
	want := 14 - 3*len(melds)
	if len(tiles) != want {
		return DiscardComparison{}, fmt.Errorf("%w: %d melds need %d tiles, got %d", ErrDrawnCount, len(melds), want, len(tiles))
	}
	h, err := hand.New(tiles[:len(tiles)-1], melds)
	if err != nil {
		return DiscardComparison{}, err
	}
	return CompareDiscards(h, tiles[len(tiles)-1])
}

// Hand returns the hand before the draw.
func (c DiscardComparison) Hand() hand.Hand { return c.hand }

// Tsumo returns the drawn tile.
func (c DiscardComparison) Tsumo() tile.Tile { return c.tsumo }

// Candidates lists the discards, best first: lowest shanten, then most
// ukeire. Ties keep the order the tiles are held in.
func (c DiscardComparison) Candidates() []Candidate {
	return append([]Candidate(nil), c.candidates...)
}

// Shanten is the shanten of the drawn hand: agari when some discard's hand
// is completed by the discarded tile, otherwise the best candidate's.
func (c DiscardComparison) Shanten() hand.Shanten {
	if c.IsWinning() {
		return hand.Agari
	}
	return c.candidates[0].Shanten
}

// IsWinning reports whether the drawn hand is complete.
func (c DiscardComparison) IsWinning() bool {
	for _, cand := range c.candidates {
		if winning.IsForm(cand.Hand, cand.Tile) {
			return true
		}
	}
	return false
}

func discardableKinds(h hand.Hand, tsumo tile.Tile) []tile.Tile {
	seen := map[tile.Tile]bool{}
	var out []tile.Tile
	for _, t := range append(h.ClosedTiles(), tsumo) {
		if !seen[t.Kind()] {
			seen[t.Kind()] = true
			out = append(out, t.Kind())
		}
	}
	return out
}
