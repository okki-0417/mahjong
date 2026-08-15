// Package hand models a player's hand: the concealed tiles, the melds, and
// what the hand can become (shanten, waits, calls).
package hand

import (
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/tile"
)

const (
	handSize = 13
	maxMelds = 4
	// A kan holds four tiles but takes the same three slots as any meld.
	meldTileSlots = 3
)

var (
	// ErrTooManyMelds is returned when a hand would hold more than four melds.
	ErrTooManyMelds = errors.New("hand: too many melds")
	// ErrClosedTileCount is returned when the concealed tiles are not 13 - 3
	// per meld. A hand is always the tiles left after discarding.
	ErrClosedTileCount = errors.New("hand: closed tile count does not match melds")
	// ErrTooManyCopies is returned when a hand holds more than four of a kind.
	ErrTooManyCopies = errors.New("hand: more than four of a kind")
	// ErrTileNotInHand is returned when a discard or call names a tile the
	// hand does not hold.
	ErrTileNotInHand = errors.New("hand: tile not in hand")
	// ErrNoPonToKakan is returned by Kakan when the hand has no pon of the tile.
	ErrNoPonToKakan = errors.New("hand: no pon to upgrade")
	// ErrInvalidTile is returned when a tile argument is not a valid tile.
	ErrInvalidTile = errors.New("hand: invalid tile")
)

// Hand is a player's hand after discarding: 13 - 3N concealed tiles and N
// melds. It is an immutable value; every operation returns a new Hand.
type Hand struct {
	closed []tile.Tile
	melds  []Meld
}

// New validates and builds a hand from concealed tiles and melds.
func New(closed []tile.Tile, melds []Meld) (Hand, error) {
	h := Hand{
		closed: append([]tile.Tile(nil), closed...),
		melds:  append([]Meld(nil), melds...),
	}
	for _, t := range h.closed {
		if !t.IsValid() {
			return Hand{}, fmt.Errorf("%w: %v", ErrInvalidTile, t)
		}
	}
	if len(h.melds) > maxMelds {
		return Hand{}, fmt.Errorf("%w: %d melds, max %d", ErrTooManyMelds, len(h.melds), maxMelds)
	}
	if want := handSize - meldTileSlots*len(h.melds); len(h.closed) != want {
		return Hand{}, fmt.Errorf("%w: %d melds need %d closed tiles, got %d", ErrClosedTileCount, len(h.melds), want, len(h.closed))
	}
	for k, c := range h.copiesByKind() {
		if c > tile.CopiesPerKind {
			return Hand{}, fmt.Errorf("%w: %v", ErrTooManyCopies, k)
		}
	}
	return h, nil
}

// Must is like New but panics on an invalid hand.
func Must(closed []tile.Tile, melds []Meld) Hand {
	h, err := New(closed, melds)
	if err != nil {
		panic(err)
	}
	return h
}

// ClosedTiles returns the concealed tiles in the order they are held.
func (h Hand) ClosedTiles() []tile.Tile {
	return append([]tile.Tile(nil), h.closed...)
}

// Melds returns the melds in the order they were made.
func (h Hand) Melds() []Meld {
	return append([]Meld(nil), h.melds...)
}

// AllTiles returns the concealed tiles followed by every meld's tiles.
func (h Hand) AllTiles() []tile.Tile {
	out := make([]tile.Tile, 0, len(h.closed)+4*len(h.melds))
	out = append(out, h.closed...)
	for _, m := range h.melds {
		out = append(out, m.tiles[:m.n]...)
	}
	return out
}

// IsMenzen reports whether the hand is concealed: no meld was called.
func (h Hand) IsMenzen() bool {
	for _, m := range h.melds {
		if m.IsCalled() {
			return false
		}
	}
	return true
}

// Discard removes tile from the hand while added enters it. A hand is always
// 13 - 3N tiles, so a discard is only balanced together with the tile that
// came in; discarding the drawn tile itself is added == tile.
func (h Hand) Discard(t, added tile.Tile) (Hand, error) {
	return h.exchange(added, []tile.Tile{t}, h.melds)
}

// Chi calls a sequence: another player's discard plus two consecutive tiles
// from the hand, followed by a discard. Two tiles leave the hand for a
// three-slot meld, so the call is only complete once a tile is discarded.
func (h Hand) Chi(called tile.Tile, consumed []tile.Tile, discard tile.Tile) (Hand, error) {
	return h.call(Chi, called, consumed, discard, true)
}

// Pon calls a triplet: another player's discard plus two matching tiles from
// the hand, followed by a discard.
func (h Hand) Pon(called tile.Tile, consumed []tile.Tile, discard tile.Tile) (Hand, error) {
	return h.call(Pon, called, consumed, discard, true)
}

// Minkan calls an open quad: another player's discard plus three matching
// tiles from the hand. Three tiles leave for three slots, so no discard is
// needed to balance the count.
func (h Hand) Minkan(called tile.Tile, consumed []tile.Tile) (Hand, error) {
	return h.call(Minkan, called, consumed, 0, false)
}

// Ankan declares a concealed quad from four matching tiles while added enters
// the hand. The hand stays concealed.
func (h Hand) Ankan(tiles []tile.Tile, added tile.Tile) (Hand, error) {
	m, err := NewMeld(Ankan, tiles)
	if err != nil {
		return Hand{}, err
	}
	melds := append(append([]Meld(nil), h.melds...), m)
	return h.exchange(added, tiles, melds)
}

// Kakan adds t to an existing pon of the same kind, upgrading it to a
// minkan, while added enters the hand. The pon was a call, so the hand stays
// open.
func (h Hand) Kakan(t, added tile.Tile) (Hand, error) {
	idx := -1
	for i, m := range h.melds {
		if m.kind == Pon && m.tiles[0].SameKind(t) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Hand{}, fmt.Errorf("%w: %v", ErrNoPonToKakan, t)
	}
	upgraded, err := NewMeld(Minkan, append(h.melds[idx].Tiles(), t))
	if err != nil {
		return Hand{}, err
	}
	melds := append([]Meld(nil), h.melds...)
	melds[idx] = upgraded
	return h.exchange(added, []tile.Tile{t}, melds)
}

// Shanten returns how many tiles away from a winning hand the hand is.
func (h Hand) Shanten() Shanten {
	return ShantenOf(h.closed, h.melds)
}

// IsTenpai reports whether the hand is one tile from winning.
func (h Hand) IsTenpai() bool {
	return h.Shanten().IsTenpai()
}

// ImprovingTiles returns the kinds of tiles that would lower the shanten,
// in sort order. A kind the hand already holds four of cannot be drawn and
// is left out. How many of each remain is a question for the tile supply.
func (h Hand) ImprovingTiles() []tile.Tile {
	current := h.Shanten()
	var out []tile.Tile
	trial := make([]tile.Tile, len(h.closed)+1)
	copy(trial, h.closed)
	for _, k := range tile.Kinds() {
		if !h.CanHold(k) {
			continue
		}
		trial[len(h.closed)] = k
		if ShantenOf(trial, h.melds) < current {
			out = append(out, k)
		}
	}
	return out
}

// Waits returns the winning tiles of a tenpai hand: the improving tiles when
// one more tile completes it. A hand that is not tenpai has no waits.
func (h Hand) Waits() []tile.Tile {
	if !h.IsTenpai() {
		return nil
	}
	return h.ImprovingTiles()
}

// CanHold reports whether the hand could take one more of t's kind. Only
// four of a kind exist, counting melds.
func (h Hand) CanHold(t tile.Tile) bool {
	return h.copiesByKind()[t.Kind()] < tile.CopiesPerKind
}

// Equal reports whether two hands hold the same tiles in the same order with
// the same melds.
func (h Hand) Equal(other Hand) bool {
	if len(h.closed) != len(other.closed) || len(h.melds) != len(other.melds) {
		return false
	}
	for i := range h.closed {
		if h.closed[i] != other.closed[i] {
			return false
		}
	}
	for i := range h.melds {
		if h.melds[i] != other.melds[i] {
			return false
		}
	}
	return true
}

// String renders the concealed tiles and melds.
func (h Hand) String() string {
	return fmt.Sprintf("%v %v", h.closed, h.melds)
}

func (h Hand) call(kind MeldKind, called tile.Tile, consumed []tile.Tile, discard tile.Tile, withDiscard bool) (Hand, error) {
	m, err := NewMeld(kind, append(append([]tile.Tile(nil), consumed...), called))
	if err != nil {
		return Hand{}, err
	}
	remaining, err := removeTiles(h.closed, consumed)
	if err != nil {
		return Hand{}, err
	}
	if withDiscard {
		remaining, err = removeTiles(remaining, []tile.Tile{discard})
		if err != nil {
			return Hand{}, err
		}
	}
	return New(remaining, append(append([]Meld(nil), h.melds...), m))
}

func (h Hand) exchange(added tile.Tile, removed []tile.Tile, melds []Meld) (Hand, error) {
	if !added.IsValid() {
		return Hand{}, fmt.Errorf("%w: %v", ErrInvalidTile, added)
	}
	with := append(append([]tile.Tile(nil), h.closed...), added)
	remaining, err := removeTiles(with, removed)
	if err != nil {
		return Hand{}, err
	}
	return New(remaining, melds)
}

// removeTiles takes each tile out of from, preferring the identical tile and
// falling back to any tile of the same kind (a red five for a plain five).
func removeTiles(from []tile.Tile, removed []tile.Tile) ([]tile.Tile, error) {
	remaining := append([]tile.Tile(nil), from...)
	for _, r := range removed {
		idx := indexOf(remaining, func(t tile.Tile) bool { return t == r })
		if idx < 0 {
			idx = indexOf(remaining, func(t tile.Tile) bool { return t.SameKind(r) })
		}
		if idx < 0 {
			return nil, fmt.Errorf("%w: %v", ErrTileNotInHand, r)
		}
		remaining = append(remaining[:idx], remaining[idx+1:]...)
	}
	return remaining, nil
}

func indexOf(tiles []tile.Tile, match func(tile.Tile) bool) int {
	for i, t := range tiles {
		if match(t) {
			return i
		}
	}
	return -1
}

func (h Hand) copiesByKind() map[tile.Tile]int {
	counts := make(map[tile.Tile]int, handSize)
	for _, t := range h.AllTiles() {
		counts[t.Kind()]++
	}
	return counts
}
