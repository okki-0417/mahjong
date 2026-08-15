package hand

import (
	"github.com/okki-0417/mahjong/tile"
)

// Shanten is the number of tile exchanges needed to reach tenpai. Agari is
// -1 (a complete hand), Tenpai is 0.
type Shanten int

const (
	// Agari is the shanten of a complete hand.
	Agari Shanten = -1
	// Tenpai is the shanten of a hand one tile from winning.
	Tenpai Shanten = 0
)

// IsTenpai reports whether the shanten is 0.
func (s Shanten) IsTenpai() bool {
	return s == Tenpai
}

// ShantenOf returns the smallest shanten over the standard form, chiitoitsu,
// and kokushi. Unlike Hand it accepts a 14-tile hand and answers Agari.
// Chiitoitsu and kokushi need a concealed hand, so they only count when there
// are no melds.
func ShantenOf(closed []tile.Tile, melds []Meld) Shanten {
	best := standardFormShanten(closed, len(melds))
	if len(melds) == 0 {
		best = min(best, chiitoitsuShanten(closed))
		best = min(best, kokushiShanten(closed))
	}
	return best
}
