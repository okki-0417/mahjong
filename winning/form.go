package winning

import (
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

// FormKind is which of the three winning shapes a form is.
type FormKind uint8

const (
	// Standard is four sets and a pair.
	Standard FormKind = iota + 1
	// Chiitoitsu is seven distinct pairs.
	Chiitoitsu
	// Kokushi is one of each terminal and honor plus one duplicate.
	Kokushi
)

// String returns "standard", "chiitoitsu", or "kokushi".
func (k FormKind) String() string {
	switch k {
	case Standard:
		return "standard"
	case Chiitoitsu:
		return "chiitoitsu"
	case Kokushi:
		return "kokushi"
	}
	return fmt.Sprintf("FormKind(%d)", uint8(k))
}

// WaitKind is the shape the winning tile completed.
type WaitKind uint8

const (
	// Ryanmen waits on either end of two consecutive tiles.
	Ryanmen WaitKind = iota + 1
	// Kanchan waits on the middle of a sequence.
	Kanchan
	// Penchan waits on the 3 of 12 or the 7 of 89.
	Penchan
	// Tanki waits on the pair. Chiitoitsu and kokushi are always tanki.
	Tanki
	// Shanpon waits on either of two pairs.
	Shanpon
)

var waitKindNames = map[WaitKind]string{Ryanmen: "ryanmen", Kanchan: "kanchan", Penchan: "penchan", Tanki: "tanki", Shanpon: "shanpon"}

// String returns the lowercase romaji name.
func (k WaitKind) String() string {
	if n, ok := waitKindNames[k]; ok {
		return n
	}
	return fmt.Sprintf("WaitKind(%d)", uint8(k))
}

// Form is one reading of a complete hand: the sets and pair of a standard
// hand, the seven pairs of chiitoitsu, or the kokushi pair. Fourteen tiles
// can be read more than one way, and yaku and fu are defined on a single
// reading, so Forms lists every reading and Winning keeps the best.
type Form struct {
	kind        FormKind
	winningTile tile.Tile
	mentsu      []Mentsu
	pairTile    tile.Tile
	pairTiles   []tile.Tile
	waitKind    WaitKind
	menzen      bool
}

// Forms lists every winning reading of the hand plus the winning tile. A
// hand that is not complete has no forms.
//
// The situation only decides whether the set completed by ron is exposed;
// pass the zero Situation when fu do not matter, such as listing waits.
func Forms(h hand.Hand, winningTile tile.Tile, s Situation) []Form {
	melds := h.Melds()
	tiles := append(h.ClosedTiles(), winningTile)

	forms := standardForms(tiles, melds, winningTile, s)
	forms = append(forms, chiitoitsuForms(tiles, melds, winningTile)...)
	forms = append(forms, kokushiForms(tiles, melds, winningTile)...)
	return forms
}

// IsForm reports whether the hand plus the winning tile is a complete hand,
// yaku or not. Waits and discard comparisons only need the shape.
func IsForm(h hand.Hand, winningTile tile.Tile) bool {
	return len(Forms(h, winningTile, Situation{})) > 0
}

// Kind returns which shape the form is.
func (f *Form) Kind() FormKind {
	return f.kind
}

// WinningTile returns the tile that completed the hand.
func (f *Form) WinningTile() tile.Tile {
	return f.winningTile
}

// Mentsu returns the four sets of a standard form, concealed sets first
// then melds. Other kinds have none.
func (f *Form) Mentsu() []Mentsu {
	return append([]Mentsu(nil), f.mentsu...)
}

// PairTile returns the pair of a standard form or the duplicated tile of
// kokushi. Chiitoitsu has no single pair; ok is false.
func (f *Form) PairTile() (t tile.Tile, ok bool) {
	if f.kind == Chiitoitsu {
		return 0, false
	}
	return f.pairTile, true
}

// PairTiles returns one tile from each of chiitoitsu's seven pairs. Other
// kinds have none.
func (f *Form) PairTiles() []tile.Tile {
	return append([]tile.Tile(nil), f.pairTiles...)
}

// WaitKind returns the shape the winning tile completed.
func (f *Form) WaitKind() WaitKind {
	return f.waitKind
}

// IsMenzen reports whether the hand was concealed: no called meld. A set
// completed by ron is open for fu but does not break menzen; chiitoitsu and
// kokushi are always menzen.
func (f *Form) IsMenzen() bool {
	if f.kind != Standard {
		return true
	}
	return f.menzen
}

func (f *Form) allTiles() []tile.Tile {
	switch f.kind {
	case Standard:
		out := make([]tile.Tile, 0, 16)
		for _, m := range f.mentsu {
			out = append(out, m.tiles[:m.n]...)
		}
		return append(out, f.pairTile, f.pairTile)
	case Chiitoitsu:
		out := make([]tile.Tile, 0, 14)
		for _, t := range f.pairTiles {
			out = append(out, t, t)
		}
		return out
	default:
		out := make([]tile.Tile, 0, 14)
		for _, k := range tile.Kinds() {
			if k.IsTerminalOrHonor() {
				out = append(out, k)
			}
		}
		return append(out, f.pairTile)
	}
}

func (f *Form) allTilesSatisfy(pred func(tile.Tile) bool) bool {
	for _, t := range f.allTiles() {
		if !pred(t) {
			return false
		}
	}
	return true
}

func (f *Form) hasHonor() bool {
	for _, t := range f.allTiles() {
		if t.IsHonor() {
			return true
		}
	}
	return false
}

// oneNumericSuit reports whether the form's numeric tiles all share one
// suit; a form with no numeric tiles does not.
func (f *Form) oneNumericSuit() bool {
	var found tile.Suit
	for _, t := range f.allTiles() {
		if t.IsHonor() {
			continue
		}
		if found == 0 {
			found = t.Suit()
		} else if found != t.Suit() {
			return false
		}
	}
	return found != 0
}

func standardForms(tiles []tile.Tile, melds []hand.Meld, winningTile tile.Tile, s Situation) []Form {
	ds := decompositions(tiles, 4-len(melds))
	forms := make([]Form, 0, len(ds))
	for _, d := range ds {
		forms = append(forms, formsFrom(d, tiles, melds, winningTile, s)...)
	}
	return forms
}

func formsFrom(d decomposition, tiles []tile.Tile, melds []hand.Meld, winningTile tile.Tile, s Situation) []Form {
	pool := append([]tile.Tile(nil), tiles...)
	pairTile := popKind(&pool, d.pair)
	popKind(&pool, d.pair)

	closedMentsu := make([]Mentsu, 0, len(d.mentsu))
	for _, spec := range d.mentsu {
		mentsuTiles := make([]tile.Tile, 0, 3)
		for _, k := range spec.kinds() {
			mentsuTiles = append(mentsuTiles, popKind(&pool, k))
		}
		closedMentsu = append(closedMentsu, newMentsu(spec.kind, mentsuTiles, false))
	}
	openMentsu := make([]Mentsu, 0, len(melds))
	for _, m := range melds {
		openMentsu = append(openMentsu, mentsuFromMeld(m))
	}
	menzen := true
	for _, m := range melds {
		if m.IsCalled() {
			menzen = false
		}
	}

	waits := d.waits(winningTile.Kind())
	forms := make([]Form, 0, len(waits))
	for _, w := range waits {
		mentsu := make([]Mentsu, 0, 4)
		for i, m := range closedMentsu {
			if s.IsRon() && !w.onPair && w.mentsuIndex == i {
				m.open = true
			}
			mentsu = append(mentsu, m)
		}
		mentsu = append(mentsu, openMentsu...)
		forms = append(forms, Form{
			kind:        Standard,
			winningTile: winningTile,
			mentsu:      mentsu,
			pairTile:    pairTile,
			waitKind:    w.kind,
			menzen:      menzen,
		})
	}
	return forms
}

// popKind removes and returns the first tile of the kind from the pool.
func popKind(pool *[]tile.Tile, kind tile.Tile) tile.Tile {
	for i, t := range *pool {
		if t.Kind() == kind {
			*pool = append((*pool)[:i], (*pool)[i+1:]...)
			return t
		}
	}
	panic(fmt.Sprintf("winning: %v not in pool", kind))
}

func chiitoitsuForms(tiles []tile.Tile, melds []hand.Meld, winningTile tile.Tile) []Form {
	if len(melds) > 0 {
		return nil
	}
	var order []tile.Tile
	first := map[tile.Tile]tile.Tile{}
	counts := map[tile.Tile]int{}
	for _, t := range tiles {
		k := t.Kind()
		if _, seen := counts[k]; !seen {
			order = append(order, k)
			first[k] = t
		}
		counts[k]++
	}
	if len(order) != 7 {
		return nil
	}
	pairTiles := make([]tile.Tile, 0, 7)
	for _, k := range order {
		if counts[k] != 2 {
			return nil
		}
		pairTiles = append(pairTiles, first[k])
	}
	return []Form{{kind: Chiitoitsu, winningTile: winningTile, pairTiles: pairTiles, waitKind: Tanki, menzen: true}}
}

func kokushiForms(tiles []tile.Tile, melds []hand.Meld, winningTile tile.Tile) []Form {
	if len(melds) > 0 {
		return nil
	}
	counts := map[tile.Tile]int{}
	for _, t := range tiles {
		if !t.IsTerminalOrHonor() {
			return nil
		}
		counts[t.Kind()]++
	}
	if len(counts) != 13 {
		return nil
	}
	var duplicated tile.Tile
	for k, c := range counts {
		if c == 2 {
			if duplicated != 0 {
				return nil
			}
			duplicated = k
		}
	}
	if duplicated == 0 {
		return nil
	}
	var pairTile tile.Tile
	for _, t := range tiles {
		if t.Kind() == duplicated {
			pairTile = t
			break
		}
	}
	return []Form{{kind: Kokushi, winningTile: winningTile, pairTile: pairTile, waitKind: Tanki, menzen: true}}
}
