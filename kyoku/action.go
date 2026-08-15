// Package kyoku plays one kyoku (a single deal) of riichi mahjong: the wall,
// the four seats, every legal choice at each moment, and the result. A
// Kyoku is the wall plus the list of actions taken; the current position is
// derived by folding the actions, so a game can be replayed from its record.
package kyoku

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/tile"
)

// Seats is the number of seats at the table. Seats are numbered 0 to 3
// counter-clockwise; seat 0 is the initial dealer's seat by convention.
const Seats = 4

// ActionKind is what a player chose to do, or a marker of whose turn comes
// next.
type ActionKind uint8

// Every kind of action. ActionHandoff and ActionRinshan are not choices: they mark
// that play moved to a seat's draw (from the wall, or from the dead wall
// after a kan) and are recorded so the position can be read from its last
// action alone.
const (
	ActionDiscard ActionKind = iota + 1
	ActionRiichi
	ActionAnkan
	ActionKakan
	ActionChi
	ActionPon
	ActionMinkan
	ActionTsumo
	ActionRon
	ActionPass
	ActionHandoff
	ActionRinshan
	ActionKyushukyuhai
)

var actionKindNames = map[ActionKind]string{
	ActionDiscard: "discard", ActionRiichi: "riichi", ActionAnkan: "ankan", ActionKakan: "kakan",
	ActionChi: "chi", ActionPon: "pon", ActionMinkan: "minkan",
	ActionTsumo: "tsumo", ActionRon: "ron", ActionPass: "pass",
	ActionHandoff: "handoff", ActionRinshan: "rinshan", ActionKyushukyuhai: "kyushukyuhai",
}

var actionKindsByName = func() map[string]ActionKind {
	m := make(map[string]ActionKind, len(actionKindNames))
	for k, n := range actionKindNames {
		m[n] = k
	}
	return m
}()

// How many tiles from the hand each kind names. The called tile, the drawn
// tile, and the winning tile are fixed by what just happened, so they are
// not part of the choice; a minkan uses all three matching tiles so there is
// nothing to choose either.
var actionTileCounts = map[ActionKind]int{
	ActionDiscard: 1, ActionRiichi: 1, ActionAnkan: 1, ActionKakan: 1,
	ActionChi: 2, ActionPon: 2, ActionMinkan: 0,
	ActionTsumo: 0, ActionRon: 0, ActionPass: 0, ActionHandoff: 0, ActionRinshan: 0, ActionKyushukyuhai: 0,
}

// String returns the lowercase name.
func (k ActionKind) String() string {
	if n, ok := actionKindNames[k]; ok {
		return n
	}
	return fmt.Sprintf("ActionKind(%d)", uint8(k))
}

// ParseActionKind returns the kind for its lowercase name.
func ParseActionKind(name string) (ActionKind, error) {
	k, ok := actionKindsByName[name]
	if !ok {
		return 0, fmt.Errorf("%w: unknown kind %q", ErrInvalidAction, name)
	}
	return k, nil
}

func (k ActionKind) takesCalledTile() bool {
	return k == ActionChi || k == ActionPon || k == ActionMinkan
}

// ErrInvalidAction is returned when an action's parts do not fit its kind.
var ErrInvalidAction = errors.New("kyoku: invalid action")

// Action is one choice by one seat. Actions are the record of a kyoku, so
// they hold only what the rules do not already fix. Whether an action is
// legal right now is the position's business.
//
// Actions compare with ==; the tiles are kept sorted so the same choice is
// the same value however it was written.
type Action struct {
	kind   ActionKind
	seat   int
	n      uint8
	tiles  [2]tile.Tile
	called tile.Tile
}

// NewAction validates and builds an action of any kind. The constructors
// below are more convenient for a known kind.
func NewAction(kind ActionKind, seat int, tiles []tile.Tile, called tile.Tile) (Action, error) {
	want, ok := actionTileCounts[kind]
	if !ok {
		return Action{}, fmt.Errorf("%w: unknown kind %v", ErrInvalidAction, kind)
	}
	if seat < 0 || seat >= Seats {
		return Action{}, fmt.Errorf("%w: seat %d", ErrInvalidAction, seat)
	}
	if len(tiles) != want {
		return Action{}, fmt.Errorf("%w: %v takes %d tiles, got %d", ErrInvalidAction, kind, want, len(tiles))
	}
	for _, t := range tiles {
		if !t.IsValid() {
			return Action{}, fmt.Errorf("%w: tile %v", ErrInvalidAction, t)
		}
	}
	if kind.takesCalledTile() != called.IsValid() {
		return Action{}, fmt.Errorf("%w: %v and called tile %v", ErrInvalidAction, kind, called)
	}
	a := Action{kind: kind, seat: seat, n: uint8(len(tiles)), called: called}
	copy(a.tiles[:], tile.Sorted(tiles))
	return a, nil
}

func mustAction(kind ActionKind, seat int, tiles []tile.Tile, called tile.Tile) Action {
	a, err := NewAction(kind, seat, tiles, called)
	if err != nil {
		panic(err)
	}
	return a
}

// NewDiscard discards a tile.
func NewDiscard(seat int, t tile.Tile) Action {
	return mustAction(ActionDiscard, seat, []tile.Tile{t}, 0)
}

// NewRiichi declares riichi while discarding a tile.
func NewRiichi(seat int, t tile.Tile) Action {
	return mustAction(ActionRiichi, seat, []tile.Tile{t}, 0)
}

// NewAnkan declares a concealed kan of the tile's kind.
func NewAnkan(seat int, t tile.Tile) Action { return mustAction(ActionAnkan, seat, []tile.Tile{t}, 0) }

// NewKakan adds the tile to a pon.
func NewKakan(seat int, t tile.Tile) Action { return mustAction(ActionKakan, seat, []tile.Tile{t}, 0) }

// NewChi calls a chi on the discard with two tiles from the hand.
func NewChi(seat int, called tile.Tile, consumed []tile.Tile) Action {
	return mustAction(ActionChi, seat, consumed, called)
}

// NewPon calls a pon on the discard with two tiles from the hand.
func NewPon(seat int, called tile.Tile, consumed []tile.Tile) Action {
	return mustAction(ActionPon, seat, consumed, called)
}

// NewMinkan calls an open kan on the discard.
func NewMinkan(seat int, called tile.Tile) Action { return mustAction(ActionMinkan, seat, nil, called) }

// NewTsumo wins on the drawn tile.
func NewTsumo(seat int) Action { return mustAction(ActionTsumo, seat, nil, 0) }

// NewRon wins on the discard.
func NewRon(seat int) Action { return mustAction(ActionRon, seat, nil, 0) }

// NewPass declines to call the discard.
func NewPass(seat int) Action { return mustAction(ActionPass, seat, nil, 0) }

// NewHandoff marks that play moved to the seat's draw.
func NewHandoff(seat int) Action { return mustAction(ActionHandoff, seat, nil, 0) }

// NewRinshan marks that the seat draws from the dead wall after a kan.
func NewRinshan(seat int) Action { return mustAction(ActionRinshan, seat, nil, 0) }

// NewKyushukyuhai aborts the kyoku with nine terminals and honors.
func NewKyushukyuhai(seat int) Action { return mustAction(ActionKyushukyuhai, seat, nil, 0) }

// Kind returns what was chosen.
func (a Action) Kind() ActionKind { return a.kind }

// Seat returns who chose.
func (a Action) Seat() int { return a.seat }

// Tiles returns the tiles from the hand the action names, sorted.
func (a Action) Tiles() []tile.Tile {
	out := make([]tile.Tile, a.n)
	copy(out, a.tiles[:a.n])
	return out
}

// CalledTile returns the tile taken from another seat's discard for a chi,
// pon, or minkan; ok is false for other kinds.
func (a Action) CalledTile() (t tile.Tile, ok bool) {
	return a.called, a.kind.takesCalledTile()
}

// IsZero reports whether a is the zero Action rather than a real one.
func (a Action) IsZero() bool { return a.kind == 0 }

// String renders the action.
func (a Action) String() string {
	if a.kind.takesCalledTile() {
		return fmt.Sprintf("%v(seat %d, %v, called %v)", a.kind, a.seat, a.Tiles(), a.called)
	}
	return fmt.Sprintf("%v(seat %d, %v)", a.kind, a.seat, a.Tiles())
}

type actionJSON struct {
	Seat   int      `json:"seat"`
	Kind   string   `json:"kind"`
	Tiles  []string `json:"tiles"`
	Called *string  `json:"called,omitempty"`
}

// MarshalJSON renders {"seat", "kind", "tiles", "called"?}.
func (a Action) MarshalJSON() ([]byte, error) {
	j := actionJSON{Seat: a.seat, Kind: a.kind.String(), Tiles: tile.Labels(a.Tiles())}
	if a.kind.takesCalledTile() {
		label := a.called.String()
		j.Called = &label
	}
	return json.Marshal(j)
}

// UnmarshalJSON reads the form MarshalJSON writes.
func (a *Action) UnmarshalJSON(data []byte) error {
	var j actionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	kind, err := ParseActionKind(j.Kind)
	if err != nil {
		return err
	}
	tiles, err := tile.ParseAll(j.Tiles)
	if err != nil {
		return err
	}
	var called tile.Tile
	if j.Called != nil {
		if called, err = tile.Parse(*j.Called); err != nil {
			return err
		}
	}
	parsed, err := NewAction(kind, j.Seat, tiles, called)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}
