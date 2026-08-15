package winning

import (
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

// MentsuKind is the shape of a set in a winning form.
type MentsuKind uint8

const (
	// Shuntsu is a sequence.
	Shuntsu MentsuKind = iota + 1
	// Koutsu is a triplet.
	Koutsu
	// Kantsu is a quad.
	Kantsu
)

// String returns "shuntsu", "koutsu", or "kantsu".
func (k MentsuKind) String() string {
	switch k {
	case Shuntsu:
		return "shuntsu"
	case Koutsu:
		return "koutsu"
	case Kantsu:
		return "kantsu"
	}
	return fmt.Sprintf("MentsuKind(%d)", uint8(k))
}

// Mentsu is one set of a standard winning form. Open sets came from a call,
// or were completed by ron; a concealed quad is Kantsu and not open.
type Mentsu struct {
	kind  MentsuKind
	n     uint8
	tiles [4]tile.Tile
	open  bool
}

func newMentsu(kind MentsuKind, tiles []tile.Tile, open bool) Mentsu {
	m := Mentsu{kind: kind, n: uint8(len(tiles)), open: open}
	copy(m.tiles[:], tiles)
	return m
}

func mentsuFromMeld(m hand.Meld) Mentsu {
	switch m.Kind() {
	case hand.Chi:
		return newMentsu(Shuntsu, m.Tiles(), true)
	case hand.Pon:
		return newMentsu(Koutsu, m.Tiles(), true)
	case hand.Minkan:
		return newMentsu(Kantsu, m.Tiles(), true)
	default:
		return newMentsu(Kantsu, m.Tiles(), false)
	}
}

// Kind returns the shape of the set.
func (m Mentsu) Kind() MentsuKind {
	return m.kind
}

// Tiles returns the set's tiles.
func (m Mentsu) Tiles() []tile.Tile {
	out := make([]tile.Tile, m.n)
	copy(out, m.tiles[:m.n])
	return out
}

// IsOpen reports whether the set is exposed: called, or completed by ron.
func (m Mentsu) IsOpen() bool {
	return m.open
}

// IsTriplet reports whether the set is a triplet or a quad.
func (m Mentsu) IsTriplet() bool {
	return m.kind == Koutsu || m.kind == Kantsu
}

// representative is the lowest tile: the start of a sequence, or the kind
// of a triplet (a plain five when the set holds a red five).
func (m Mentsu) representative() tile.Tile {
	best := m.tiles[0].Kind()
	for _, t := range m.tiles[1:m.n] {
		if k := t.Kind(); k < best {
			best = k
		}
	}
	return best
}

func (m Mentsu) suit() tile.Suit {
	return m.representative().Suit()
}

func (m Mentsu) number() int {
	return m.representative().EffectiveNumber()
}

func (m Mentsu) fu() int {
	if m.kind == Shuntsu {
		return 0
	}
	base := 4
	if m.open {
		base = 2
	}
	if m.representative().IsTerminalOrHonor() {
		base *= 2
	}
	if m.kind == Kantsu {
		base *= 4
	}
	return base
}

func (m Mentsu) containsTerminalOrHonor() bool {
	for _, t := range m.tiles[:m.n] {
		if t.IsTerminalOrHonor() {
			return true
		}
	}
	return false
}

func (m Mentsu) containsTerminal() bool {
	for _, t := range m.tiles[:m.n] {
		if t.IsTerminal() {
			return true
		}
	}
	return false
}

// String renders the kind, openness, and tiles.
func (m Mentsu) String() string {
	if m.open {
		return fmt.Sprintf("%v(open)%v", m.kind, m.Tiles())
	}
	return fmt.Sprintf("%v%v", m.kind, m.Tiles())
}
