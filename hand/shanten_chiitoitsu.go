package hand

import "github.com/okki-0417/mahjong/tile"

const (
	chiitoitsuPairs = 7
	chiitoitsuBase  = chiitoitsuPairs - 1
)

// chiitoitsuShanten is 6 - pairs + max(0, 7 - kinds). Three or four of a
// kind still count as one pair, and fewer than seven kinds is a penalty
// because a seven-pairs hand needs seven distinct kinds.
func chiitoitsuShanten(closed []tile.Tile) Shanten {
	counts := make(map[tile.Tile]int, len(closed))
	for _, t := range closed {
		counts[t.Kind()]++
	}
	pairs := 0
	for _, c := range counts {
		if c >= 2 {
			pairs++
		}
	}
	return Shanten(chiitoitsuBase - pairs + max(0, chiitoitsuPairs-len(counts)))
}
