package kyoku

import "github.com/okki-0417/mahjong/tile"

// liveWall is the tiles still to be drawn, in draw order. The last tile
// (haitei) is what moves to the dead wall after a kan.
type liveWall struct {
	tiles []tile.Tile
}

func (w liveWall) remaining() int { return len(w.tiles) }
func (w liveWall) empty() bool    { return len(w.tiles) == 0 }

func (w liveWall) nextDraw() (tile.Tile, bool) {
	if w.empty() {
		return 0, false
	}
	return w.tiles[0], true
}

func (w liveWall) draw() liveWall {
	return liveWall{tiles: w.tiles[1:]}
}

func (w liveWall) haitei() tile.Tile {
	return w.tiles[len(w.tiles)-1]
}

func (w liveWall) dropHaitei() liveWall {
	return liveWall{tiles: w.tiles[:len(w.tiles)-1]}
}

const (
	rinshanSize       = 4
	uradoraBase       = 4
	doraIndicatorBase = 9
	maxDoraIndicators = 5
)

// deadWall is the 14 tiles that supply rinshan draws and dora indicators:
// positions 0-3 are rinshan tiles, 4-8 uradora indicators, 9-13 dora
// indicators. A rinshan tile taken is replaced in place by the haitei, so
// the indicator positions never move.
type deadWall struct {
	tiles          []tile.Tile
	doraIndicators int
	rinshanDrawn   int
}

func newDeadWall(tiles []tile.Tile) deadWall {
	return deadWall{tiles: tiles, doraIndicators: 1}
}

func (d deadWall) nextRinshan() (tile.Tile, bool) {
	if d.rinshanExhausted() {
		return 0, false
	}
	return d.tiles[d.rinshanDrawn], true
}

func (d deadWall) rinshanExhausted() bool {
	return d.rinshanDrawn >= rinshanSize
}

func (d deadWall) drawRinshan(replenishment tile.Tile) deadWall {
	tiles := append([]tile.Tile(nil), d.tiles...)
	tiles[d.rinshanDrawn] = replenishment
	return deadWall{tiles: tiles, doraIndicators: d.doraIndicators, rinshanDrawn: d.rinshanDrawn + 1}
}

func (d deadWall) withKanDoraRevealed() deadWall {
	return deadWall{tiles: d.tiles, doraIndicators: min(d.doraIndicators+1, maxDoraIndicators), rinshanDrawn: d.rinshanDrawn}
}

func (d deadWall) doraIndicatorTiles() []tile.Tile {
	return append([]tile.Tile(nil), d.tiles[doraIndicatorBase:doraIndicatorBase+d.doraIndicators]...)
}

func (d deadWall) uradoraIndicatorTiles() []tile.Tile {
	return append([]tile.Tile(nil), d.tiles[uradoraBase:uradoraBase+d.doraIndicators]...)
}
