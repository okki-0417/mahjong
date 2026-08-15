package kyoku

import (
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

var (
	// ErrInvalidSetup is returned by Deal for a setup that cannot start a
	// kyoku.
	ErrInvalidSetup = errors.New("kyoku: invalid setup")
	// ErrFinished is returned by Take when the kyoku is over.
	ErrFinished = errors.New("kyoku: the kyoku is finished")
	// ErrNotFinished is returned by DealNext while the kyoku is still going.
	ErrNotFinished = errors.New("kyoku: the kyoku is not finished")
	// ErrHanchanOver is returned by DealNext when this was the last kyoku.
	ErrHanchanOver = errors.New("kyoku: the hanchan is over")
)

// The winds of a hanchan (east-south game). An east-only game would move
// this into the rule set.
var roundWinds = []tile.Wind{tile.EastWind, tile.SouthWind}

const kyokusPerWind = 4

// Setup is the starting conditions of a kyoku. Zero values mean: a freshly
// shuffled wall, seat 0 dealing, east round, kyoku 1, no honba or sticks,
// every seat at the rule set's starting score, and default rules.
type Setup struct {
	Wall         *Wall
	DealerSeat   int
	RoundWind    tile.Wind
	KyokuNumber  int
	Scores       *[Seats]int
	Honba        int
	RiichiSticks int
	RuleSet      ruleset.RuleSet
	// Actions replays a recorded kyoku up to its current position.
	Actions []Action
}

// Kyoku is one deal: the wall, the actions taken so far, and the position
// and result they lead to. It is immutable; Take returns a new Kyoku.
type Kyoku struct {
	wall    Wall
	initial Kyokumen
	actions []Action
	current Kyokumen
	result  *Result
}

// Deal starts a kyoku from a setup, folding any recorded actions.
func Deal(s Setup) (*Kyoku, error) {
	wall := ShuffledWall(nil)
	if s.Wall != nil {
		wall = *s.Wall
	}
	if s.DealerSeat < 0 || s.DealerSeat >= Seats {
		return nil, fmt.Errorf("%w: dealer seat %d", ErrInvalidSetup, s.DealerSeat)
	}
	roundWind := s.RoundWind
	if roundWind == 0 {
		roundWind = tile.EastWind
	}
	if !roundWind.IsValid() {
		return nil, fmt.Errorf("%w: round wind %v", ErrInvalidSetup, roundWind)
	}
	number := s.KyokuNumber
	if number == 0 {
		number = 1
	}
	if number < 1 || number > kyokusPerWind {
		return nil, fmt.Errorf("%w: kyoku number %d", ErrInvalidSetup, number)
	}
	if s.Honba < 0 || s.RiichiSticks < 0 {
		return nil, fmt.Errorf("%w: honba %d, riichi sticks %d", ErrInvalidSetup, s.Honba, s.RiichiSticks)
	}
	var scores [Seats]int
	if s.Scores != nil {
		scores = *s.Scores
	} else {
		for i := range scores {
			scores[i] = s.RuleSet.StartingScore()
		}
	}
	initial := dealKyokumen(wall, s.DealerSeat, roundWind, scores, number, s.Honba, s.RiichiSticks, s.RuleSet)
	k := &Kyoku{wall: wall, initial: initial}
	return k.fold(s.Actions)
}

// fold applies recorded actions in order from the opening.
func (k *Kyoku) fold(actions []Action) (*Kyoku, error) {
	current := &Kyoku{wall: k.wall, initial: k.initial, current: k.initial}
	current.result = current.drawResult()
	for i, a := range actions {
		next, err := current.step(a)
		if err != nil {
			return nil, fmt.Errorf("%w (action %d)", err, i)
		}
		current = next
	}
	return current, nil
}

// step applies one legal action. The result comes from the position the
// terminal action was taken in; an exhaustive draw is not anyone's action,
// so it is read off the new position.
func (k *Kyoku) step(a Action) (*Kyoku, error) {
	if k.IsFinished() {
		return nil, ErrFinished
	}
	if !containsAction(k.current.LegalActions(), a) {
		return nil, fmt.Errorf("%w: %v", ErrIllegalAction, a)
	}
	advanced := k.current.advance(a)
	next := &Kyoku{
		wall: k.wall, initial: k.initial, actions: append(append([]Action(nil), k.actions...), a),
		current: advanced,
	}
	if r, ok := resultOf(k.current.takenOn(a), advanced.last); ok {
		next.result = r
	} else {
		next.result = next.drawResult()
	}
	return next, nil
}

func (k *Kyoku) drawResult() *Result {
	if d, ok := k.current.DrawKind(); ok {
		return newResult(k.current, resultKindOfDraw(d), -1)
	}
	return nil
}

func containsAction(actions []Action, a Action) bool {
	for _, x := range actions {
		if x == a {
			return true
		}
	}
	return false
}

// Take appends one action and returns the kyoku after it. The action must be
// among the current legal actions.
func (k *Kyoku) Take(a Action) (*Kyoku, error) {
	return k.step(a)
}

// Wall returns the wall the kyoku was dealt from.
func (k *Kyoku) Wall() Wall { return k.wall }

// Actions returns the actions taken so far.
func (k *Kyoku) Actions() []Action { return append([]Action(nil), k.actions...) }

// Kyokumen returns the current position.
func (k *Kyoku) Kyokumen() *Kyokumen {
	current := k.current
	return &current
}

// Result returns the result, once the kyoku is finished.
func (k *Kyoku) Result() (*Result, bool) { return k.result, k.result != nil }

// IsFinished reports whether the kyoku has a result.
func (k *Kyoku) IsFinished() bool { return k.result != nil }

// AwaitingSeats lists the seats that must choose next: the seat on turn,
// or the seats that have not answered a claimable tile.
func (k *Kyoku) AwaitingSeats() []int {
	if k.IsFinished() {
		return nil
	}
	var out []int
	for _, a := range k.current.LegalActions() {
		if !containsSeat(out, a.seat) {
			out = append(out, a.seat)
		}
	}
	return out
}

// SeenBy projects the current position onto what one seat can see.
func (k *Kyoku) SeenBy(seat int) Sight { return k.Kyokumen().SeenBy(seat) }

// IsAllLast reports whether this is the last kyoku of the hanchan by
// name: the last wind, with the last dealer dealing. It is known from the
// deal, before the kyoku is played out.
func (k *Kyoku) IsAllLast() bool {
	return k.initial.roundWind == roundWinds[len(roundWinds)-1] && k.initial.dealerSeat == k.lastDealerSeat()
}

// IsLast reports whether the hanchan ended with this kyoku: someone went
// below zero, or the all-last dealer lost the deal. It is false while the
// kyoku is unfinished.
func (k *Kyoku) IsLast() bool {
	if !k.IsFinished() {
		return false
	}
	return k.busted() || (k.IsAllLast() && !k.result.DealerContinues())
}

// DealNext deals the following kyoku from this one's result, with a fresh
// wall.
func (k *Kyoku) DealNext() (*Kyoku, error) {
	if !k.IsFinished() {
		return nil, ErrNotFinished
	}
	if k.IsLast() {
		return nil, ErrHanchanOver
	}
	roundWind, number, dealer := k.succeeding()
	scores := k.result.Scores()
	return Deal(Setup{
		DealerSeat: dealer, RoundWind: roundWind, KyokuNumber: number,
		Honba: k.result.NextHonba(), RiichiSticks: k.result.CarriedRiichiSticks(),
		Scores: &scores, RuleSet: k.initial.rules,
	})
}

// chiichaSeat is the first dealer, found by winding the deal back.
func (k *Kyoku) chiichaSeat() int {
	return (k.initial.dealerSeat - (k.initial.kyokuNumber - 1) + Seats*Seats) % Seats
}

func (k *Kyoku) lastDealerSeat() int {
	return (k.chiichaSeat() - 1 + Seats) % Seats
}

func (k *Kyoku) busted() bool {
	for _, s := range k.result.Scores() {
		if s < 0 {
			return true
		}
	}
	return false
}

func (k *Kyoku) succeeding() (tile.Wind, int, int) {
	if k.result.DealerContinues() {
		return k.initial.roundWind, k.initial.kyokuNumber, k.initial.dealerSeat
	}
	next := k.initial.Shimocha(k.initial.dealerSeat)
	if k.initial.kyokuNumber < kyokusPerWind {
		return k.initial.roundWind, k.initial.kyokuNumber + 1, next
	}
	for i, w := range roundWinds {
		if w == k.initial.roundWind {
			return roundWinds[i+1], 1, next
		}
	}
	panic("kyoku: no wind after the last")
}
