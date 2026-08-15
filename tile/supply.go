package tile

import (
	"errors"
	"fmt"
)

// ErrOversupplied is returned when more than four of a kind are seen.
var ErrOversupplied = errors.New("tile: more than four of a kind seen")

// Supply answers how many of each kind are still unseen, given every tile
// already seen: a hand, its melds, the discards, dora indicators. Where a
// tile was seen makes no difference to the supply, so it is not recorded.
//
// The zero value has seen nothing: four of every kind remain.
type Supply struct {
	seen map[Tile]int
}

// NewSupply records the seen tiles. It fails when more than four of a kind
// are seen.
func NewSupply(seen []Tile) (Supply, error) {
	counts := make(map[Tile]int, len(seen))
	for _, t := range seen {
		if !t.IsValid() {
			return Supply{}, fmt.Errorf("%w: %v", ErrInvalidLabel, t)
		}
		counts[t.Kind()]++
	}
	for k, c := range counts {
		if c > CopiesPerKind {
			return Supply{}, fmt.Errorf("%w: %v", ErrOversupplied, k)
		}
	}
	return Supply{seen: counts}, nil
}

// MustSupply is like NewSupply but panics on error.
func MustSupply(seen []Tile) Supply {
	s, err := NewSupply(seen)
	if err != nil {
		panic(err)
	}
	return s
}

// Remaining returns how many of t's kind are still unseen.
func (s Supply) Remaining(t Tile) int {
	return CopiesPerKind - s.seen[t.Kind()]
}
