package kyoku

import (
	"encoding/json"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

// Sight is the position as one seat sees it: its own hand, everything on
// the table (rivers, melds, dora indicators, scores), and its own legal
// actions. Other hands, the wall, uradora, and rinshan tiles are not
// answered.
type Sight struct {
	k    *Kyokumen
	seat int
}

// Seat returns whose view this is.
func (s Sight) Seat() int { return s.seat }

// Hand returns the seat's hand, without the drawn tile.
func (s Sight) Hand() hand.Hand { return s.k.seats[s.seat].hand }

// ConcealedTiles returns the tiles still in hand; tiles promised to a
// pending call are already on the table.
func (s Sight) ConcealedTiles() []tile.Tile { return s.k.ConcealedTiles(s.seat) }

// Drawn returns the tile just drawn, only on the seat's own turn.
func (s Sight) Drawn() (tile.Tile, bool) {
	if !s.IsMyTurn() {
		return 0, false
	}
	return s.k.Drawn()
}

// IsMyTurn reports whether the seat is the one to discard.
func (s Sight) IsMyTurn() bool {
	seat, ok := s.k.DiscardingSeat()
	return ok && seat == s.seat
}

// LegalActions lists what this seat may do now; other seats' options are
// not shown.
func (s Sight) LegalActions() []Action {
	var out []Action
	for _, a := range s.k.LegalActions() {
		if a.seat == s.seat {
			out = append(out, a)
		}
	}
	return out
}

// TileSupply returns the supply as seen from this seat.
func (s Sight) TileSupply() tile.Supply { return s.k.TileSupply(s.seat) }

// Discards returns every tile a seat threw, called ones marked.
func (s Sight) Discards(of int) []Discard { return s.k.seats[of].Discards() }

// Melds returns a seat's exposed melds.
func (s Sight) Melds(of int) []hand.Meld { return s.k.seats[of].hand.Melds() }

// PendingCall returns a seat's declared but unfinished chi or pon.
func (s Sight) PendingCall(of int) (Action, bool) { return s.k.PendingCall(of) }

// Score returns a seat's points.
func (s Sight) Score(of int) int { return s.k.Score(of) }

// IsRiichi reports whether a seat has declared riichi.
func (s Sight) IsRiichi(of int) bool { return s.k.seats[of].IsRiichi() }

// SeatWind returns a seat's wind.
func (s Sight) SeatWind(of int) tile.Wind { return s.k.SeatWind(of) }

// IsDealer reports whether a seat is the dealer.
func (s Sight) IsDealer(of int) bool { return s.k.IsDealer(of) }

// ConcealedCount returns how many tiles a seat holds in hand.
func (s Sight) ConcealedCount(of int) int { return s.k.ConcealedCount(of) }

// DealerSeat returns the dealer's seat.
func (s Sight) DealerSeat() int { return s.k.dealerSeat }

// RoundWind returns the prevailing wind.
func (s Sight) RoundWind() tile.Wind { return s.k.roundWind }

// KyokuNumber returns which kyoku of the wind this is.
func (s Sight) KyokuNumber() int { return s.k.kyokuNumber }

// Honba returns the repeat counter.
func (s Sight) Honba() int { return s.k.honba }

// RiichiSticks returns the riichi sticks on the table.
func (s Sight) RiichiSticks() int { return s.k.riichiSticks }

// Junme returns the turn number.
func (s Sight) Junme() int { return s.k.junme }

// RemainingDraws returns how many tiles are left to draw.
func (s Sight) RemainingDraws() int { return s.k.RemainingDraws() }

// DoraIndicators returns the revealed dora indicators.
func (s Sight) DoraIndicators() []tile.Tile { return s.k.DoraIndicators() }

// DiscardingSeat returns the seat to discard next, if any.
func (s Sight) DiscardingSeat() (int, bool) { return s.k.DiscardingSeat() }

// ClaimedTile returns the tile open to calls, if any.
func (s Sight) ClaimedTile() (tile.Tile, bool) { return s.k.ClaimedTile() }

// ClaimedFrom returns the seat whose tile is open to calls, if any.
func (s Sight) ClaimedFrom() (int, bool) { return s.k.ClaimedFrom() }

type sightSeatJSON struct {
	Seat           int         `json:"seat"`
	Discards       []Discard   `json:"discards"`
	Melds          []hand.Meld `json:"melds"`
	PendingCall    *Action     `json:"pending_call"`
	ConcealedCount int         `json:"concealed_count"`
	Score          int         `json:"score"`
	Riichi         bool        `json:"riichi"`
	SeatWind       string      `json:"seat_wind"`
}

type sightJSON struct {
	Seat           int             `json:"seat"`
	ClosedTiles    []string        `json:"closed_tiles"`
	Drawn          *string         `json:"drawn"`
	LegalActions   []Action        `json:"legal_actions"`
	Seats          []sightSeatJSON `json:"seats"`
	DealerSeat     int             `json:"dealer_seat"`
	RoundWind      string          `json:"round_wind"`
	KyokuNumber    int             `json:"kyoku_number"`
	Honba          int             `json:"honba"`
	RiichiSticks   int             `json:"riichi_sticks"`
	Junme          int             `json:"junme"`
	RemainingDraws int             `json:"remaining_draws"`
	DoraIndicators []string        `json:"dora_indicators"`
	DiscardingSeat *int            `json:"discarding_seat"`
	ClaimedTile    *string         `json:"claimed_tile"`
	ClaimedFrom    *int            `json:"claimed_from"`
}

func optionalTile(t tile.Tile, ok bool) *string {
	if !ok {
		return nil
	}
	label := t.String()
	return &label
}

func optionalInt(v int, ok bool) *int {
	if !ok {
		return nil
	}
	return &v
}

// MarshalJSON renders the view for a client. Rivers are given as the full
// discard record so a called tile still shows who threw it.
func (s Sight) MarshalJSON() ([]byte, error) {
	j := sightJSON{
		Seat:           s.seat,
		ClosedTiles:    tile.Labels(s.ConcealedTiles()),
		Drawn:          optionalTile(s.Drawn()),
		LegalActions:   emptyIfNil(s.LegalActions()),
		DealerSeat:     s.DealerSeat(),
		RoundWind:      s.RoundWind().String(),
		KyokuNumber:    s.KyokuNumber(),
		Honba:          s.Honba(),
		RiichiSticks:   s.RiichiSticks(),
		Junme:          s.Junme(),
		RemainingDraws: s.RemainingDraws(),
		DoraIndicators: tile.Labels(s.DoraIndicators()),
		DiscardingSeat: optionalInt(s.DiscardingSeat()),
		ClaimedTile:    optionalTile(s.ClaimedTile()),
		ClaimedFrom:    optionalInt(s.ClaimedFrom()),
	}
	for seat := 0; seat < Seats; seat++ {
		var pending *Action
		if call, ok := s.PendingCall(seat); ok {
			pending = &call
		}
		melds := s.Melds(seat)
		if melds == nil {
			melds = []hand.Meld{}
		}
		discards := s.Discards(seat)
		if discards == nil {
			discards = []Discard{}
		}
		j.Seats = append(j.Seats, sightSeatJSON{
			Seat: seat, Discards: discards, Melds: melds, PendingCall: pending,
			ConcealedCount: s.ConcealedCount(seat), Score: s.Score(seat), Riichi: s.IsRiichi(seat),
			SeatWind: s.SeatWind(seat).String(),
		})
	}
	return json.Marshal(j)
}

func emptyIfNil(actions []Action) []Action {
	if actions == nil {
		return []Action{}
	}
	return actions
}
