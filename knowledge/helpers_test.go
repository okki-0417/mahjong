package knowledge_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

func labels(tiles []tile.Tile) string {
	return strings.Join(tile.Labels(tiles), " ")
}

func meldKinds(h hand.Hand) []hand.MeldKind {
	out := make([]hand.MeldKind, 0, len(h.Melds()))
	for _, m := range h.Melds() {
		out = append(out, m.Kind())
	}
	return out
}

func must(t *testing.T) func(hand.Hand, error) hand.Hand {
	return func(h hand.Hand, err error) hand.Hand {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
}

func lastParamIsTile(fn any, numIn int) bool {
	typ := reflect.TypeOf(fn)
	return typ.NumIn() == numIn && typ.In(numIn-1) == reflect.TypeOf(tile.Tile(0))
}

func contains(tiles []tile.Tile, x tile.Tile) bool {
	for _, t := range tiles {
		if t == x {
			return true
		}
	}
	return false
}

func has(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// sit builds the situation every winning example starts from.
func sit(kind winning.WinKind, round, seat tile.Wind) winning.Situation {
	return winning.Situation{WinKind: kind, RoundWind: round, SeatWind: seat}
}

func winOf(closed, win string, melds []hand.Meld, s winning.Situation, rs ruleset.RuleSet) (*winning.Winning, error) {
	return winning.New(mt.Hand(closed, melds...), mt.T(win), s, rs)
}

// yakuNames lists the names of the yaku that scored, or nothing when the
// hand is not a win.
func yakuNames(w *winning.Winning, err error) []string {
	if err != nil {
		return nil
	}
	var names []string
	for _, y := range w.Score(0).Yakus() {
		names = append(names, y.Name)
	}
	return names
}

// hanOf returns the han of the named yaku, or -1 when the hand is not a win
// or the yaku did not score.
func hanOf(name string) func(*winning.Winning, error) int {
	return func(w *winning.Winning, err error) int {
		if err != nil {
			return -1
		}
		for _, y := range w.Score(0).Yakus() {
			if y.Name == name {
				return y.Han
			}
		}
		return -1
	}
}
