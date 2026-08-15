package hand

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/tile"
)

// MeldKind is how a meld was formed.
type MeldKind uint8

const (
	// Chi is a sequence called from the player to the left.
	Chi MeldKind = iota + 1
	// Pon is a triplet called from any other player.
	Pon
	// Minkan is an open quad: called from a discard, or a pon upgraded by
	// kakan. Kakan keeps Minkan because the pon underneath was a call.
	Minkan
	// Ankan is a concealed quad made only from the player's own tiles, so it
	// does not open the hand.
	Ankan
)

var meldKindNames = map[MeldKind]string{Chi: "chi", Pon: "pon", Minkan: "minkan", Ankan: "ankan"}

var meldKindsByName = func() map[string]MeldKind {
	m := make(map[string]MeldKind, len(meldKindNames))
	for k, n := range meldKindNames {
		m[n] = k
	}
	return m
}()

// String returns the lowercase name: chi, pon, minkan, ankan.
func (k MeldKind) String() string {
	if n, ok := meldKindNames[k]; ok {
		return n
	}
	return fmt.Sprintf("MeldKind(%d)", uint8(k))
}

// ParseMeldKind returns the kind for its lowercase name.
func ParseMeldKind(name string) (MeldKind, error) {
	k, ok := meldKindsByName[name]
	if !ok {
		return 0, fmt.Errorf("%w: unknown kind %q", ErrInvalidMeld, name)
	}
	return k, nil
}

func (k MeldKind) tileCount() int {
	if k == Minkan || k == Ankan {
		return 4
	}
	return 3
}

// IsCalled reports whether a meld of this kind was taken from another
// player's discard, opening the hand.
func (k MeldKind) IsCalled() bool {
	return k == Chi || k == Pon || k == Minkan
}

// IsKan reports whether a meld of this kind is a quad.
func (k MeldKind) IsKan() bool {
	return k == Minkan || k == Ankan
}

// ErrInvalidMeld is returned when tiles do not form a meld of the given kind.
var ErrInvalidMeld = errors.New("hand: invalid meld")

// Meld is an exposed set of tiles: a sequence, triplet, or quad. It is an
// immutable value and can be compared with ==.
type Meld struct {
	kind  MeldKind
	n     uint8
	tiles [4]tile.Tile
}

// NewMeld validates that tiles form a meld of the given kind. Tiles keep the
// order they were given in.
func NewMeld(kind MeldKind, tiles []tile.Tile) (Meld, error) {
	if _, ok := meldKindNames[kind]; !ok {
		return Meld{}, fmt.Errorf("%w: unknown kind %v", ErrInvalidMeld, kind)
	}
	if len(tiles) != kind.tileCount() {
		return Meld{}, fmt.Errorf("%w: %v needs %d tiles, got %d", ErrInvalidMeld, kind, kind.tileCount(), len(tiles))
	}
	for _, t := range tiles {
		if !t.IsValid() {
			return Meld{}, fmt.Errorf("%w: %v", ErrInvalidMeld, t)
		}
	}
	m := Meld{kind: kind, n: uint8(len(tiles))}
	copy(m.tiles[:], tiles)

	if kind == Chi {
		if err := validateChi(tiles); err != nil {
			return Meld{}, err
		}
		return m, nil
	}
	for _, t := range tiles[1:] {
		if !t.SameKind(tiles[0]) {
			return Meld{}, fmt.Errorf("%w: %v tiles must be the same kind", ErrInvalidMeld, kind)
		}
	}
	return m, nil
}

// MustMeld is like NewMeld but panics on an invalid meld.
func MustMeld(kind MeldKind, tiles []tile.Tile) Meld {
	m, err := NewMeld(kind, tiles)
	if err != nil {
		panic(err)
	}
	return m
}

func validateChi(tiles []tile.Tile) error {
	suit := tiles[0].Suit()
	if !suit.IsNumeric() {
		return fmt.Errorf("%w: chi cannot use honors", ErrInvalidMeld)
	}
	var numbers [3]int
	for i, t := range tiles {
		if t.Suit() != suit {
			return fmt.Errorf("%w: chi tiles must share a suit", ErrInvalidMeld)
		}
		numbers[i] = t.EffectiveNumber()
	}
	sortThree(&numbers)
	if numbers[1] != numbers[0]+1 || numbers[2] != numbers[0]+2 {
		return fmt.Errorf("%w: chi tiles must be consecutive", ErrInvalidMeld)
	}
	return nil
}

func sortThree(n *[3]int) {
	if n[0] > n[1] {
		n[0], n[1] = n[1], n[0]
	}
	if n[1] > n[2] {
		n[1], n[2] = n[2], n[1]
	}
	if n[0] > n[1] {
		n[0], n[1] = n[1], n[0]
	}
}

// Kind returns how the meld was formed.
func (m Meld) Kind() MeldKind {
	return m.kind
}

// Tiles returns the meld's tiles in the order given to NewMeld.
func (m Meld) Tiles() []tile.Tile {
	out := make([]tile.Tile, m.n)
	copy(out, m.tiles[:m.n])
	return out
}

// IsCalled reports whether the meld opened the hand.
func (m Meld) IsCalled() bool {
	return m.kind.IsCalled()
}

// IsKan reports whether the meld is a quad.
func (m Meld) IsKan() bool {
	return m.kind.IsKan()
}

// String renders the meld as kind and tile labels, e.g. "pon[1z 1z 1z]".
func (m Meld) String() string {
	return fmt.Sprintf("%v%v", m.kind, m.Tiles())
}

type meldJSON struct {
	Kind  string      `json:"kind"`
	Tiles []tile.Tile `json:"tiles"`
}

// MarshalJSON renders {"kind": "pon", "tiles": ["1z", "1z", "1z"]}.
func (m Meld) MarshalJSON() ([]byte, error) {
	return json.Marshal(meldJSON{Kind: m.kind.String(), Tiles: m.Tiles()})
}

// UnmarshalJSON reads the form MarshalJSON writes and validates the meld.
func (m *Meld) UnmarshalJSON(data []byte) error {
	var j meldJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	kind, err := ParseMeldKind(j.Kind)
	if err != nil {
		return err
	}
	parsed, err := NewMeld(kind, j.Tiles)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}
