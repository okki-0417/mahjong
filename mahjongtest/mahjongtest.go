// Package mahjongtest builds tiles, melds, and hands from the compact label
// notation used throughout the tests ("1m 2m 3m", "0p" for a red five, "7z"
// for chun). The domain packages never accept strings; every conversion
// lives here.
package mahjongtest

import (
	"strings"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

// T returns the tile for one label. It panics on an invalid label.
func T(label string) tile.Tile {
	return tile.MustParse(label)
}

// Tiles parses whitespace-separated labels. Several strings may be given and
// are concatenated. It panics on an invalid label.
func Tiles(labels ...string) []tile.Tile {
	var out []tile.Tile
	for _, group := range labels {
		for _, l := range strings.Fields(group) {
			out = append(out, tile.MustParse(l))
		}
	}
	return out
}

// Chi builds a chi meld from labels.
func Chi(labels string) hand.Meld {
	return hand.MustMeld(hand.Chi, Tiles(labels))
}

// Pon builds a pon meld from labels.
func Pon(labels string) hand.Meld {
	return hand.MustMeld(hand.Pon, Tiles(labels))
}

// Minkan builds a minkan meld from labels.
func Minkan(labels string) hand.Meld {
	return hand.MustMeld(hand.Minkan, Tiles(labels))
}

// Ankan builds an ankan meld from labels.
func Ankan(labels string) hand.Meld {
	return hand.MustMeld(hand.Ankan, Tiles(labels))
}

// Hand builds a hand from concealed tile labels and melds. It panics when the
// hand is invalid.
func Hand(closed string, melds ...hand.Meld) hand.Hand {
	return hand.Must(Tiles(closed), melds)
}
