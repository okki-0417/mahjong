package hand

import "github.com/okki-0417/mahjong/tile"

const (
	maxMentsu         = 4
	standardFormBase  = 8
	numberSlots       = 10
	honorNumbers      = 7
	numericNumbers    = 9
	standardFormSuits = 4
)

type suitCounts [numberSlots]uint8

type blocks struct{ mentsu, taatsu int }

// standardFormShanten is 8 - 2M - T - P for the best decomposition into M
// complete sets (melds included), T partial sets, and P = 1 with a pair.
//
// T is capped at 4 - M: a partial set beyond the four set slots is no
// progress, otherwise a hand with four sets and a partial set but no pair
// would score below tenpai.
//
// Every pair candidate (plus "no pair") is tried; the rest is decomposed
// per suit into Pareto-best (M, T) options and combined across suits.
func standardFormShanten(closed []tile.Tile, meldCount int) Shanten {
	var counts [standardFormSuits]suitCounts
	for _, t := range closed {
		counts[suitIndex(t.Suit())][t.EffectiveNumber()]++
	}
	d := newSuitDecomposer()

	best := d.evaluate(counts, meldCount, false)
	for s := range counts {
		for n := 1; n < numberSlots; n++ {
			if counts[s][n] < 2 {
				continue
			}
			sub := counts
			sub[s][n] -= 2
			best = min(best, d.evaluate(sub, meldCount, true))
		}
	}
	return best
}

func suitIndex(s tile.Suit) int {
	return int(s - tile.Man)
}

type decomposeKey struct {
	counts suitCounts
	honor  bool
}

type suitDecomposer struct {
	memo map[decomposeKey][]blocks
}

func newSuitDecomposer() *suitDecomposer {
	return &suitDecomposer{memo: make(map[decomposeKey][]blocks)}
}

func (d *suitDecomposer) evaluate(counts [standardFormSuits]suitCounts, meldCount int, hasPair bool) Shanten {
	var options [standardFormSuits][]blocks
	for s := range counts {
		options[s] = d.decompose(counts[s], s == suitIndex(tile.Honor))
	}
	pair := 0
	if hasPair {
		pair = 1
	}
	best := Shanten(standardFormBase)
	for _, a := range options[0] {
		for _, b := range options[1] {
			for _, c := range options[2] {
				for _, z := range options[3] {
					m := min(a.mentsu+b.mentsu+c.mentsu+z.mentsu+meldCount, maxMentsu)
					t := min(a.taatsu+b.taatsu+c.taatsu+z.taatsu, maxMentsu-m)
					best = min(best, Shanten(standardFormBase-2*m-t-pair))
				}
			}
		}
	}
	return best
}

// decompose lists the (M, T) options for one suit, keeping only the largest
// T for each M. Honors form no sequences or partial sequences.
func (d *suitDecomposer) decompose(counts suitCounts, honor bool) []blocks {
	key := decomposeKey{counts, honor}
	if cached, ok := d.memo[key]; ok {
		return cached
	}
	var found []blocks
	enumerate(counts, honor, 0, 0, &found)

	bestTaatsu := map[int]int{}
	for _, b := range found {
		if t, ok := bestTaatsu[b.mentsu]; !ok || b.taatsu > t {
			bestTaatsu[b.mentsu] = b.taatsu
		}
	}
	pareto := make([]blocks, 0, len(bestTaatsu))
	for m, t := range bestTaatsu {
		pareto = append(pareto, blocks{m, t})
	}
	d.memo[key] = pareto
	return pareto
}

func enumerate(counts suitCounts, honor bool, mentsu, taatsu int, found *[]blocks) {
	n := 1
	for n < numberSlots && counts[n] == 0 {
		n++
	}
	if n == numberSlots {
		*found = append(*found, blocks{mentsu, taatsu})
		return
	}
	limit := numericNumbers
	if honor {
		limit = honorNumbers
	}

	if counts[n] >= 3 {
		sub := counts
		sub[n] -= 3
		enumerate(sub, honor, mentsu+1, taatsu, found)
	}
	if !honor && n+2 <= limit && counts[n+1] >= 1 && counts[n+2] >= 1 {
		sub := counts
		sub[n]--
		sub[n+1]--
		sub[n+2]--
		enumerate(sub, honor, mentsu+1, taatsu, found)
	}
	if counts[n] >= 2 {
		sub := counts
		sub[n] -= 2
		enumerate(sub, honor, mentsu, taatsu+1, found)
	}
	if !honor && n+1 <= limit && counts[n+1] >= 1 {
		sub := counts
		sub[n]--
		sub[n+1]--
		enumerate(sub, honor, mentsu, taatsu+1, found)
	}
	if !honor && n+2 <= limit && counts[n+2] >= 1 {
		sub := counts
		sub[n]--
		sub[n+2]--
		enumerate(sub, honor, mentsu, taatsu+1, found)
	}
	sub := counts
	sub[n]--
	enumerate(sub, honor, mentsu, taatsu, found)
}
