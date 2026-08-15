package kyoku

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/okki-0417/mahjong/tile"
)

const (
	handSize    = 13
	chunkSize   = 4
	chunkRounds = 3
	deadSize    = 14
	wallSize    = 136
	liveSize    = wallSize - Seats*handSize - deadSize
)

// ErrInvalidWall is returned for a wall that is not the full set of tiles.
var ErrInvalidWall = errors.New("kyoku: invalid wall")

// Wall is the shuffled stack of all 136 tiles for one kyoku. The first
// tiles are dealt, the last 14 become the dead wall, and the 70 between
// are drawn. Keeping the order is enough to replay the kyoku.
type Wall struct {
	tiles []tile.Tile
}

// NewWall builds a wall from an ordering of the full set.
func NewWall(tiles []tile.Tile) (Wall, error) {
	if len(tiles) != wallSize {
		return Wall{}, fmt.Errorf("%w: %d tiles, want %d", ErrInvalidWall, len(tiles), wallSize)
	}
	counts := map[tile.Tile]int{}
	for _, t := range tiles {
		if !t.IsValid() {
			return Wall{}, fmt.Errorf("%w: %v", ErrInvalidWall, t)
		}
		counts[t.Kind()]++
	}
	for _, k := range tile.Kinds() {
		if counts[k] != tile.CopiesPerKind {
			return Wall{}, fmt.Errorf("%w: %d of %v", ErrInvalidWall, counts[k], k)
		}
	}
	return Wall{tiles: append([]tile.Tile(nil), tiles...)}, nil
}

// MustWall is like NewWall but panics on error.
func MustWall(tiles []tile.Tile) Wall {
	w, err := NewWall(tiles)
	if err != nil {
		panic(err)
	}
	return w
}

// ShuffledWall shuffles the full set (with red fives) using r, or the
// package-level random source when r is nil.
func ShuffledWall(r *rand.Rand) Wall {
	tiles := tile.FullSet(true)
	shuffle := rand.Shuffle
	if r != nil {
		shuffle = r.Shuffle
	}
	shuffle(len(tiles), func(i, j int) { tiles[i], tiles[j] = tiles[j], tiles[i] })
	return Wall{tiles: tiles}
}

// Tiles returns the wall's ordering.
func (w Wall) Tiles() []tile.Tile {
	return append([]tile.Tile(nil), w.tiles...)
}

// hands deals: four tiles at a time for three rounds from the dealer on,
// then one each. Row 0 is the dealer's hand.
func (w Wall) hands() [Seats][]tile.Tile {
	var dealt [Seats][]tile.Tile
	pos := 0
	for round := 0; round < chunkRounds; round++ {
		for seat := 0; seat < Seats; seat++ {
			dealt[seat] = append(dealt[seat], w.tiles[pos:pos+chunkSize]...)
			pos += chunkSize
		}
	}
	for seat := 0; seat < Seats; seat++ {
		dealt[seat] = append(dealt[seat], w.tiles[pos])
		pos++
	}
	return dealt
}

func (w Wall) drawTiles() []tile.Tile {
	return append([]tile.Tile(nil), w.tiles[Seats*handSize:wallSize-deadSize]...)
}

func (w Wall) deadTiles() []tile.Tile {
	return append([]tile.Tile(nil), w.tiles[wallSize-deadSize:]...)
}
