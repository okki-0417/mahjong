package hand

import "github.com/okki-0417/mahjong/tile"

const kokushiKinds = 13

// kokushiShanten is 13 - distinct terminal/honor kinds - (1 if any of them is
// paired). A red five counts as a plain five and so is never a terminal.
func kokushiShanten(closed []tile.Tile) Shanten {
	counts := make(map[tile.Tile]int, kokushiKinds)
	for _, t := range closed {
		if t.IsTerminalOrHonor() {
			counts[t.Kind()]++
		}
	}
	pair := 0
	for _, c := range counts {
		if c >= 2 {
			pair = 1
			break
		}
	}
	return Shanten(kokushiKinds - len(counts) - pair)
}
