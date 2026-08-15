// Package tile defines the 34 kinds of riichi mahjong tiles plus the three
// red fives, and the vocabulary for talking about them (suit, number, honor,
// terminal, dora).
package tile

import (
	"errors"
	"fmt"
	"sort"
)

// CopiesPerKind is how many tiles of each kind a full set contains.
const CopiesPerKind = 4

// ErrInvalidLabel is returned by Parse for a string that is not a tile label.
var ErrInvalidLabel = errors.New("tile: invalid label")

// Suit is one of the three numeric suits or the honor suit.
type Suit uint8

// The four suits, in sort order.
const (
	Man Suit = iota + 1
	Pin
	Sou
	Honor
)

// String returns the one-letter suit code used in labels: m, p, s, z.
func (s Suit) String() string {
	switch s {
	case Man:
		return "m"
	case Pin:
		return "p"
	case Sou:
		return "s"
	case Honor:
		return "z"
	}
	return "?"
}

// IsNumeric reports whether the suit is man, pin, or sou.
func (s Suit) IsNumeric() bool {
	return s == Man || s == Pin || s == Sou
}

// Tile is a single tile. The zero value is not a valid tile.
//
// Tiles are compared by identity: a red five and a plain five are different
// tiles of the same kind (see SameKind and Kind).
type Tile uint8

// Every tile, in sort order. Red fives sort immediately after the plain five
// of the same suit.
const (
	M1 Tile = iota + 1
	M2
	M3
	M4
	M5
	M5R
	M6
	M7
	M8
	M9
	P1
	P2
	P3
	P4
	P5
	P5R
	P6
	P7
	P8
	P9
	S1
	S2
	S3
	S4
	S5
	S5R
	S6
	S7
	S8
	S9
	East
	South
	West
	North
	Haku
	Hatsu
	Chun
)

const (
	tileCount = int(Chun)
	kindCount = 34
	windCount = 4
	dragons   = 3
	honors    = windCount + dragons
	numbers   = 9
)

var labels = [tileCount + 1]string{
	"",
	"1m", "2m", "3m", "4m", "5m", "0m", "6m", "7m", "8m", "9m",
	"1p", "2p", "3p", "4p", "5p", "0p", "6p", "7p", "8p", "9p",
	"1s", "2s", "3s", "4s", "5s", "0s", "6s", "7s", "8s", "9s",
	"1z", "2z", "3z", "4z", "5z", "6z", "7z",
}

var suits = [tileCount + 1]Suit{
	0,
	Man, Man, Man, Man, Man, Man, Man, Man, Man, Man,
	Pin, Pin, Pin, Pin, Pin, Pin, Pin, Pin, Pin, Pin,
	Sou, Sou, Sou, Sou, Sou, Sou, Sou, Sou, Sou, Sou,
	Honor, Honor, Honor, Honor, Honor, Honor, Honor,
}

var effectiveNumbers = [tileCount + 1]int{
	0,
	1, 2, 3, 4, 5, 5, 6, 7, 8, 9,
	1, 2, 3, 4, 5, 5, 6, 7, 8, 9,
	1, 2, 3, 4, 5, 5, 6, 7, 8, 9,
	1, 2, 3, 4, 5, 6, 7,
}

var byLabel = func() map[string]Tile {
	m := make(map[string]Tile, tileCount)
	for i := 1; i <= tileCount; i++ {
		m[labels[i]] = Tile(i)
	}
	return m
}()

var kinds = func() []Tile {
	ks := make([]Tile, 0, kindCount)
	for i := 1; i <= tileCount; i++ {
		if t := Tile(i); !t.IsRed() {
			ks = append(ks, t)
		}
	}
	return ks
}()

// Parse returns the tile for a label such as "1m", "0p" (red five), or "7z".
func Parse(label string) (Tile, error) {
	t, ok := byLabel[label]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrInvalidLabel, label)
	}
	return t, nil
}

// MustParse is like Parse but panics on an invalid label.
func MustParse(label string) Tile {
	t, err := Parse(label)
	if err != nil {
		panic(err)
	}
	return t
}

// Of returns the plain tile of a suit and number: Of(Man, 5) is M5, Of(Honor,
// 1) is East. It panics when no such tile exists.
func Of(s Suit, number int) Tile {
	if s.IsNumeric() && number >= 1 && number <= numbers {
		return MustParse(fmt.Sprintf("%d%s", number, s))
	}
	if s == Honor && number >= 1 && number <= honors {
		return East + Tile(number-1)
	}
	panic(fmt.Sprintf("tile: no tile for suit %v number %d", s, number))
}

// Kinds returns the 34 kinds of tiles in sort order. Red fives are not kinds
// of their own and are excluded.
func Kinds() []Tile {
	out := make([]Tile, len(kinds))
	copy(out, kinds)
	return out
}

// FullSet returns the 136 tiles used in a game: four of each kind, in sort
// order. With red, one five of each numeric suit is replaced by its red five.
// The caller shuffles before building a wall.
func FullSet(red bool) []Tile {
	out := make([]Tile, 0, kindCount*CopiesPerKind)
	for _, k := range kinds {
		for i := 0; i < CopiesPerKind; i++ {
			if red && i == 0 && k.IsNumeric() && k.Number() == 5 {
				out = append(out, k+1)
				continue
			}
			out = append(out, k)
		}
	}
	return out
}

// IsValid reports whether t is one of the 37 tiles.
func (t Tile) IsValid() bool {
	return t >= 1 && int(t) <= tileCount
}

// String returns the label: number then suit code, with 0 for a red five.
func (t Tile) String() string {
	if !t.IsValid() {
		return fmt.Sprintf("Tile(%d)", uint8(t))
	}
	return labels[t]
}

// Suit returns the tile's suit.
func (t Tile) Suit() Suit {
	return suits[t]
}

// Number returns the digit of the label: 1–9 for numeric tiles, 0 for a red
// five, 1–7 for honors (east, south, west, north, haku, hatsu, chun).
func (t Tile) Number() int {
	if t.IsRed() {
		return 0
	}
	return effectiveNumbers[t]
}

// EffectiveNumber is Number with a red five counted as 5.
func (t Tile) EffectiveNumber() int {
	return effectiveNumbers[t]
}

// IsRed reports whether the tile is a red five.
func (t Tile) IsRed() bool {
	return t == M5R || t == P5R || t == S5R
}

// Kind returns the tile that stands for t's kind: t itself, or the plain
// five for a red five. Use it as a map key when counting copies.
func (t Tile) Kind() Tile {
	if t.IsRed() {
		return t - 1
	}
	return t
}

// SameKind reports whether two tiles are the same kind. A red five and a
// plain five are the same kind: waits, furiten, and copy limits all work in
// kinds.
func (t Tile) SameKind(other Tile) bool {
	return t.Kind() == other.Kind()
}

// IsNumeric reports whether the tile is man, pin, or sou.
func (t Tile) IsNumeric() bool {
	return t.Suit().IsNumeric()
}

// IsHonor reports whether the tile is a wind or a dragon.
func (t Tile) IsHonor() bool {
	return t.Suit() == Honor
}

// IsWind reports whether the tile is east, south, west, or north.
func (t Tile) IsWind() bool {
	return t.IsHonor() && t.Number() <= windCount
}

// IsDragon reports whether the tile is haku, hatsu, or chun.
func (t Tile) IsDragon() bool {
	return t.IsHonor() && t.Number() > windCount
}

// IsTerminal reports whether the tile is a 1 or 9 of a numeric suit.
func (t Tile) IsTerminal() bool {
	if t.IsHonor() {
		return false
	}
	n := t.EffectiveNumber()
	return n == 1 || n == numbers
}

// IsTerminalOrHonor reports whether the tile is a terminal or an honor
// (yaochuu).
func (t Tile) IsTerminalOrHonor() bool {
	return t.IsHonor() || t.IsTerminal()
}

// Dora returns the tile that is dora when t is the indicator. Numbers wrap
// 9→1, winds cycle east→south→west→north→east, dragons cycle
// haku→hatsu→chun→haku. A red five indicates as a plain five.
func (t Tile) Dora() Tile {
	n := t.EffectiveNumber()
	switch {
	case t.IsWind():
		return Of(Honor, n%windCount+1)
	case t.IsDragon():
		return Of(Honor, windCount+(n-windCount)%dragons+1)
	default:
		return Of(t.Suit(), n%numbers+1)
	}
}

// DoraIndicator returns the tile that, as an indicator, makes t dora. It is
// the inverse of Dora.
func (t Tile) DoraIndicator() Tile {
	n := t.EffectiveNumber()
	switch {
	case t.IsWind():
		return Of(Honor, (n+windCount-2)%windCount+1)
	case t.IsDragon():
		return Of(Honor, windCount+(n-windCount+dragons-2)%dragons+1)
	default:
		return Of(t.Suit(), (n+numbers-2)%numbers+1)
	}
}

// Less orders tiles by suit (m, p, s, z) then number, with a red five right
// after its plain five.
func (t Tile) Less(other Tile) bool {
	return t < other
}

// Sort sorts tiles in place in Less order.
func Sort(tiles []Tile) {
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].Less(tiles[j]) })
}

// Sorted returns a sorted copy of tiles.
func Sorted(tiles []Tile) []Tile {
	out := make([]Tile, len(tiles))
	copy(out, tiles)
	Sort(out)
	return out
}

// Labels returns the labels of tiles in order.
func Labels(tiles []Tile) []string {
	out := make([]string, len(tiles))
	for i, t := range tiles {
		out[i] = t.String()
	}
	return out
}

// ParseAll parses each label and returns the tiles in order.
func ParseAll(labels []string) ([]Tile, error) {
	out := make([]Tile, len(labels))
	for i, l := range labels {
		t, err := Parse(l)
		if err != nil {
			return nil, err
		}
		out[i] = t
	}
	return out, nil
}
