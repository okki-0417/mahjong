package kyoku

import (
	"encoding/json"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

const riichiStick = 1000

// Discard is one tile a seat threw: what it was and how. A discard that was
// called leaves the river but stays in the record, because furiten rests on
// having thrown the tile, not on it lying there.
type Discard struct {
	tile              tile.Tile
	tsumogiri         bool
	riichiDeclaration bool
	calledBy          int
}

func newDiscard(t tile.Tile, tsumogiri, riichiDeclaration bool) Discard {
	return Discard{tile: t, tsumogiri: tsumogiri, riichiDeclaration: riichiDeclaration, calledBy: -1}
}

// Tile returns the discarded tile.
func (d Discard) Tile() tile.Tile { return d.tile }

// IsTsumogiri reports whether the drawn tile was discarded as drawn.
func (d Discard) IsTsumogiri() bool { return d.tsumogiri }

// IsRiichiDeclaration reports whether the tile was turned sideways to
// declare riichi.
func (d Discard) IsRiichiDeclaration() bool { return d.riichiDeclaration }

// CalledBy returns the seat that called the tile, if any.
func (d Discard) CalledBy() (seat int, ok bool) {
	return d.calledBy, d.calledBy >= 0
}

// IsInRiver reports whether the tile still lies in the river (was not
// called).
func (d Discard) IsInRiver() bool { return d.calledBy < 0 }

func (d Discard) called(by int) Discard {
	d.calledBy = by
	return d
}

// MarshalJSON renders {"tile", "tsumogiri", "riichi_declaration", "called_by"}.
func (d Discard) MarshalJSON() ([]byte, error) {
	var calledBy *int
	if d.calledBy >= 0 {
		v := d.calledBy
		calledBy = &v
	}
	return json.Marshal(struct {
		Tile              string `json:"tile"`
		Tsumogiri         bool   `json:"tsumogiri"`
		RiichiDeclaration bool   `json:"riichi_declaration"`
		CalledBy          *int   `json:"called_by"`
	}{d.tile.String(), d.tsumogiri, d.riichiDeclaration, calledBy})
}

type riichiState struct {
	declared bool
	junme    int
	ippatsu  bool
	double   bool
}

func (r riichiState) withoutIppatsu() riichiState {
	r.ippatsu = false
	return r
}

type liability struct {
	yaku winning.YakuID
	from int
}

// SeatState is one seat's state: the hand, and everything that only the
// progress of the kyoku produces (score, discards, riichi, furiten). It is
// an immutable value; every change returns a new SeatState.
type SeatState struct {
	hand              hand.Hand
	score             int
	discards          []Discard
	riichi            riichiState
	missedThisRound   bool
	missedAfterRiichi bool
	liabilities       []liability
}

// Hand returns the hand after its last discard (never holding the drawn
// tile).
func (s SeatState) Hand() hand.Hand { return s.hand }

// Score returns the seat's points.
func (s SeatState) Score() int { return s.score }

// Discards returns every tile the seat threw, called ones included.
func (s SeatState) Discards() []Discard { return append([]Discard(nil), s.discards...) }

// IsRiichi reports whether the seat has declared riichi.
func (s SeatState) IsRiichi() bool { return s.riichi.declared }

// IsDoubleRiichi reports whether the riichi was declared on the seat's first
// uninterrupted turn.
func (s SeatState) IsDoubleRiichi() bool { return s.riichi.double }

// IsIppatsu reports whether the seat can still win with ippatsu.
func (s SeatState) IsIppatsu() bool { return s.riichi.ippatsu }

// IsMenzen reports whether the hand is concealed.
func (s SeatState) IsMenzen() bool { return s.hand.IsMenzen() }

// IsTenpai reports whether the hand is one tile from winning.
func (s SeatState) IsTenpai() bool { return s.hand.IsTenpai() }

// IsFuriten reports whether the seat may not ron: it has discarded one of
// its waits, or missed a ron this turn or after riichi.
func (s SeatState) IsFuriten() bool {
	if s.missedThisRound || s.missedAfterRiichi {
		return true
	}
	for _, w := range s.hand.Waits() {
		if s.hasDiscarded(w) {
			return true
		}
	}
	return false
}

// IsNagashiMangan reports whether every tile the seat threw is a terminal
// or honor and none was called. A seat that has not discarded is not.
func (s SeatState) IsNagashiMangan() bool {
	if len(s.discards) == 0 {
		return false
	}
	for _, d := range s.discards {
		if !d.tile.IsTerminalOrHonor() || !d.IsInRiver() {
			return false
		}
	}
	return true
}

func (s SeatState) singleRiichi() bool { return s.riichi.declared && !s.riichi.double }

func (s SeatState) hasDiscarded(t tile.Tile) bool {
	for _, d := range s.discards {
		if d.tile.SameKind(t) {
			return true
		}
	}
	return false
}

// openTiles are the tiles everyone can see: melds and the river. A called
// discard is counted with the caller's meld, not here.
func (s SeatState) openTiles() []tile.Tile {
	var out []tile.Tile
	for _, m := range s.hand.Melds() {
		out = append(out, m.Tiles()...)
	}
	for _, d := range s.discards {
		if d.IsInRiver() {
			out = append(out, d.tile)
		}
	}
	return out
}

func (s SeatState) liableSeatFor(id winning.YakuID) (int, bool) {
	for _, l := range s.liabilities {
		if l.yaku == id {
			return l.from, true
		}
	}
	return 0, false
}

func (s SeatState) withHand(h hand.Hand) SeatState {
	s.hand = h
	return s
}

func (s SeatState) placeDiscard(h hand.Hand, t tile.Tile, tsumogiri, declaration bool) SeatState {
	s.hand = h
	s.discards = append(append([]Discard(nil), s.discards...), newDiscard(t, tsumogiri, declaration))
	return s
}

func (s SeatState) markLastCalled(by int) SeatState {
	discards := append([]Discard(nil), s.discards...)
	discards[len(discards)-1] = discards[len(discards)-1].called(by)
	s.discards = discards
	return s
}

func (s SeatState) declareRiichi(junme int, double bool) SeatState {
	s.score -= riichiStick
	s.riichi = riichiState{declared: true, junme: junme, ippatsu: true, double: double}
	return s
}

func (s SeatState) withoutIppatsu() SeatState {
	s.riichi = s.riichi.withoutIppatsu()
	return s
}

func (s SeatState) onDraw() SeatState {
	s.missedThisRound = false
	return s
}

func (s SeatState) missedRon() SeatState {
	s.missedThisRound = true
	s.missedAfterRiichi = s.missedAfterRiichi || s.riichi.declared
	return s
}

func (s SeatState) withLiability(id winning.YakuID, from int) SeatState {
	kept := make([]liability, 0, len(s.liabilities)+1)
	for _, l := range s.liabilities {
		if l.yaku != id {
			kept = append(kept, l)
		}
	}
	kept = append(kept, liability{yaku: id, from: from})
	s.liabilities = kept
	return s
}
