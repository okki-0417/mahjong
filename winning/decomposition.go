package winning

import (
	"sort"

	"github.com/okki-0417/mahjong/tile"
)

// mentsuSpec is a set described by kind and starting number only; Form
// assigns the actual tiles.
type mentsuSpec struct {
	kind  MentsuKind
	suit  tile.Suit
	start int
}

func (s mentsuSpec) kinds() []tile.Tile {
	first := tile.Of(s.suit, s.start)
	if s.kind == Shuntsu {
		return []tile.Tile{first, tile.Of(s.suit, s.start+1), tile.Of(s.suit, s.start+2)}
	}
	return []tile.Tile{first, first, first}
}

func (s mentsuSpec) contains(number int) bool {
	if s.kind == Shuntsu {
		return number >= s.start && number <= s.start+2
	}
	return number == s.start
}

// decomposition is one reading of the tiles as a pair plus sets.
type decomposition struct {
	pair   tile.Tile
	mentsu []mentsuSpec
}

type wait struct {
	kind        WaitKind
	onPair      bool
	mentsuIndex int
}

// waits lists where the winning tile sits in this reading. The same reading
// can hold it in more than one place: 22234m won on 2m is both tanki and
// ryanmen.
func (d decomposition) waits(winningKind tile.Tile) []wait {
	var out []wait
	if d.pair == winningKind {
		out = append(out, wait{kind: Tanki, onPair: true})
	}
	number := winningKind.EffectiveNumber()
	for i, spec := range d.mentsu {
		if spec.suit != winningKind.Suit() || !spec.contains(number) {
			continue
		}
		out = append(out, wait{kind: waitKindIn(spec, number), mentsuIndex: i})
	}
	return out
}

func waitKindIn(spec mentsuSpec, number int) WaitKind {
	if spec.kind != Shuntsu {
		return Shanpon
	}
	switch number {
	case spec.start:
		if spec.start == 7 {
			return Penchan
		}
		return Ryanmen
	case spec.start + 2:
		if spec.start == 1 {
			return Penchan
		}
		return Ryanmen
	default:
		return Kanchan
	}
}

var decompositionSuits = [...]tile.Suit{tile.Man, tile.Pin, tile.Sou, tile.Honor}

type suitTally struct {
	counts [10]uint8
	order  []int
}

// decompositions lists every reading of tiles as one pair plus mentsuCount
// sets. Pair candidates are tried suit by suit in the order their number
// first appears, so readings come out in a stable order for the same input.
func decompositions(tiles []tile.Tile, mentsuCount int) []decomposition {
	var tally [4]suitTally
	for _, t := range tiles {
		s := &tally[suitIndex(t.Suit())]
		n := t.EffectiveNumber()
		if s.counts[n] == 0 {
			s.order = append(s.order, n)
		}
		s.counts[n]++
	}

	var out []decomposition
	for si, s := range tally {
		for _, n := range s.order {
			if s.counts[n] < 2 {
				continue
			}
			sub := tally
			sub[si].counts[n] -= 2
			perSuit, ok := decomposeSuits(sub)
			if !ok {
				continue
			}
			for _, combo := range product(perSuit) {
				if len(combo) != mentsuCount {
					continue
				}
				out = append(out, decomposition{pair: tile.Of(decompositionSuits[si], n), mentsu: combo})
			}
		}
	}
	return out
}

func suitIndex(s tile.Suit) int {
	return int(s - tile.Man)
}

func decomposeSuits(tally [4]suitTally) ([4][][]mentsuSpec, bool) {
	var perSuit [4][][]mentsuSpec
	for si, s := range tally {
		total := 0
		for _, c := range s.counts {
			total += int(c)
		}
		if total%3 != 0 {
			return perSuit, false
		}
		perSuit[si] = decomposeSuit(decompositionSuits[si], s.counts)
		if len(perSuit[si]) == 0 {
			return perSuit, false
		}
	}
	return perSuit, true
}

// decomposeSuit lists the ways one suit splits into sets, each way sorted
// and deduplicated so order of discovery does not produce repeats.
func decomposeSuit(suit tile.Suit, counts [10]uint8) [][]mentsuSpec {
	raw := decomposeRecursive(suit, counts)
	seen := map[string]bool{}
	out := make([][]mentsuSpec, 0, len(raw))
	for _, d := range raw {
		sort.Slice(d, func(i, j int) bool {
			if d[i].start != d[j].start {
				return d[i].start < d[j].start
			}
			return d[i].kind == Koutsu && d[j].kind == Shuntsu
		})
		key := specKey(d)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, d)
	}
	return out
}

func specKey(d []mentsuSpec) string {
	key := make([]byte, 0, len(d)*2)
	for _, s := range d {
		key = append(key, byte(s.kind), byte(s.start))
	}
	return string(key)
}

func decomposeRecursive(suit tile.Suit, counts [10]uint8) [][]mentsuSpec {
	n := 1
	for n < len(counts) && counts[n] == 0 {
		n++
	}
	if n == len(counts) {
		return [][]mentsuSpec{{}}
	}
	var out [][]mentsuSpec
	if counts[n] >= 3 {
		sub := counts
		sub[n] -= 3
		for _, rest := range decomposeRecursive(suit, sub) {
			out = append(out, append([]mentsuSpec{{kind: Koutsu, suit: suit, start: n}}, rest...))
		}
	}
	if suit.IsNumeric() && n <= 7 && counts[n+1] >= 1 && counts[n+2] >= 1 {
		sub := counts
		sub[n]--
		sub[n+1]--
		sub[n+2]--
		for _, rest := range decomposeRecursive(suit, sub) {
			out = append(out, append([]mentsuSpec{{kind: Shuntsu, suit: suit, start: n}}, rest...))
		}
	}
	return out
}

// product combines one way per suit into a flat list of sets, man first
// then pin, sou, honors.
func product(perSuit [4][][]mentsuSpec) [][]mentsuSpec {
	combos := [][]mentsuSpec{{}}
	for _, ways := range perSuit {
		var next [][]mentsuSpec
		for _, combo := range combos {
			for _, way := range ways {
				merged := make([]mentsuSpec, 0, len(combo)+len(way))
				merged = append(merged, combo...)
				merged = append(merged, way...)
				next = append(next, merged)
			}
		}
		combos = next
	}
	return combos
}
