// Package cpu is a computer player that chooses from what its seat can see.
//
// It plays for maximum tile efficiency with a concealed hand: it wins when it
// can, declares riichi when tenpai, and otherwise discards by the shanten
// and ukeire of the hand left behind. Calling needs value and defence
// judgement it does not have, so it always declines a call.
package cpu

import (
	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
)

// Discard order among equally efficient tiles: give up the tiles with the
// fewest uses first. A red five is worth points, so it is kept longest.
const (
	priorityRed      = 0
	prioritySimple   = 1
	priorityTerminal = 3
	priorityHonor    = 4
)

// Choose picks the seat's action in the current position. ok is false when
// the seat has nothing to choose.
func Choose(sight kyoku.Sight) (kyoku.Action, bool) {
	legal := sight.LegalActions()
	if len(legal) == 0 {
		return kyoku.Action{}, false
	}
	for _, kind := range []kyoku.ActionKind{kyoku.ActionTsumo, kyoku.ActionRon, kyoku.ActionPass, kyoku.ActionKyushukyuhai} {
		if a, ok := firstOf(legal, kind); ok {
			return a, true
		}
	}
	if a, ok := efficient(sight, legal, kyoku.ActionRiichi); ok {
		return a, true
	}
	if a, ok := efficient(sight, legal, kyoku.ActionDiscard); ok {
		return a, true
	}
	return legal[0], true
}

func firstOf(legal []kyoku.Action, kind kyoku.ActionKind) (kyoku.Action, bool) {
	for _, a := range legal {
		if a.Kind() == kind {
			return a, true
		}
	}
	return kyoku.Action{}, false
}

// efficient picks, among actions of the kind, the one whose remaining hand
// is closest to winning; ties go to the most ukeire, then to the discard
// priority. Ukeire needs every wait, so it is only counted for the fastest.
func efficient(sight kyoku.Sight, legal []kyoku.Action, kind kyoku.ActionKind) (kyoku.Action, bool) {
	drawn, hasDrawn := sight.Drawn()
	type candidate struct {
		action kyoku.Action
		after  hand.Hand
	}
	var candidates []candidate
	for _, a := range legal {
		if a.Kind() != kind {
			continue
		}
		if !hasDrawn {
			return a, true
		}
		after, err := sight.Hand().Discard(a.Tiles()[0], drawn)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{a, after})
	}
	if len(candidates) == 0 {
		return kyoku.Action{}, false
	}
	fastest := candidates[0].after.Shanten()
	for _, c := range candidates[1:] {
		fastest = min(fastest, c.after.Shanten())
	}
	supply := sight.TileSupply()
	best, bestScore := kyoku.Action{}, [2]int{-1, -1}
	found := false
	for _, c := range candidates {
		if c.after.Shanten() != fastest {
			continue
		}
		score := [2]int{ukeire.Of(c.after, supply).RemainingTotal(), discardPriority(c.action.Tiles()[0])}
		if !found || score[0] > bestScore[0] || (score[0] == bestScore[0] && score[1] > bestScore[1]) {
			best, bestScore, found = c.action, score, true
		}
	}
	return best, found
}

func discardPriority(t tile.Tile) int {
	switch {
	case t.IsRed():
		return priorityRed
	case t.IsHonor():
		return priorityHonor
	case t.IsTerminal():
		return priorityTerminal
	default:
		return prioritySimple
	}
}
