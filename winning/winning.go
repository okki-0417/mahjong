// Package winning models a win: a tenpai hand completed by a winning tile
// under a situation. It reads the hand into forms, finds the yaku, and
// scores fu and points.
package winning

import (
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

var (
	// ErrNotWinning is returned by New when the hand plus the winning tile is
	// not a win: it is not a complete shape, or it has no yaku. Errors about
	// the arguments themselves (an invalid situation, a fifth tile) do not
	// wrap it.
	ErrNotWinning = errors.New("winning: not a winning hand")
	// ErrNoForm wraps ErrNotWinning when the tiles do not form a complete hand.
	ErrNoForm = fmt.Errorf("%w: not a complete hand", ErrNotWinning)
	// ErrNoYaku wraps ErrNotWinning when the hand is complete but has no yaku.
	ErrNoYaku = fmt.Errorf("%w: no yaku", ErrNotWinning)
	// ErrTileExhausted is returned when the winning tile would be a fifth of
	// its kind.
	ErrTileExhausted = errors.New("winning: all four of the winning tile are already held")
)

// Winning is a win that has happened: a hand, the tile that completed it,
// and the situation. A hand that is not complete, or has no yaku, is not a
// win and cannot be built, so a Winning always scores.
type Winning struct {
	hand        hand.Hand
	winningTile tile.Tile
	situation   Situation
	ruleSet     ruleset.RuleSet
	readings    []reading
}

// reading is one form with the yaku it scores. Forms with no yaku are not
// readings.
type reading struct {
	form  *Form
	yakus []Yaku
}

// New builds the win, or reports why it is not one. Pass ruleset.Default()
// for the usual rules.
func New(h hand.Hand, winningTile tile.Tile, s Situation, rs ruleset.RuleSet) (*Winning, error) {
	if !winningTile.IsValid() {
		return nil, fmt.Errorf("%w: %v", hand.ErrInvalidTile, winningTile)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if !h.CanHold(winningTile) {
		return nil, fmt.Errorf("%w: %v", ErrTileExhausted, winningTile)
	}
	forms := Forms(h, winningTile, s)
	if len(forms) == 0 {
		return nil, ErrNoForm
	}
	var readings []reading
	for i := range forms {
		f := &forms[i]
		if yakus := appliedYaku(f, s, rs); len(yakus) > 0 {
			readings = append(readings, reading{form: f, yakus: yakus})
		}
	}
	if len(readings) == 0 {
		return nil, ErrNoYaku
	}
	return &Winning{hand: h, winningTile: winningTile, situation: s, ruleSet: rs, readings: readings}, nil
}

// Hand returns the hand before the winning tile.
func (w *Winning) Hand() hand.Hand {
	return w.hand
}

// WinningTile returns the tile that completed the hand.
func (w *Winning) WinningTile() tile.Tile {
	return w.winningTile
}

// Situation returns the situation of the win.
func (w *Winning) Situation() Situation {
	return w.situation
}

// Fu counts the fu of every reading and returns the highest.
func (w *Winning) Fu() Fu {
	var best Fu
	for i, r := range w.readings {
		fu := fuOf(r.form, w.situation, w.ruleSet)
		if i == 0 || fu.Total() > best.Total() {
			best = fu
		}
	}
	return best
}

// Score scores every reading with the given dora count and returns the
// highest by yakuman, then han, then fu, then points.
func (w *Winning) Score(doraCount int) Score {
	var best Score
	for i, r := range w.readings {
		s := scoreOf(r.yakus, fuOf(r.form, w.situation, w.ruleSet), w.situation, doraCount, w.ruleSet)
		if i == 0 || scoreLess(best, s) {
			best = s
		}
	}
	return best
}

func scoreLess(a, b Score) bool {
	if a.YakumanCount() != b.YakumanCount() {
		return a.YakumanCount() < b.YakumanCount()
	}
	if a.Han() != b.Han() {
		return a.Han() < b.Han()
	}
	if a.Fu() != b.Fu() {
		return a.Fu() < b.Fu()
	}
	return a.Total() < b.Total()
}
