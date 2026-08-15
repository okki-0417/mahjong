package winning

import (
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/tile"
)

// WinKind is how the winning tile arrived.
type WinKind uint8

const (
	// Ron wins on another player's discard.
	Ron WinKind = iota + 1
	// Tsumo wins on a self-drawn tile.
	Tsumo
)

// String returns "ron" or "tsumo".
func (k WinKind) String() string {
	switch k {
	case Ron:
		return "ron"
	case Tsumo:
		return "tsumo"
	}
	return fmt.Sprintf("WinKind(%d)", uint8(k))
}

// ParseWinKind returns the kind for "ron" or "tsumo".
func ParseWinKind(name string) (WinKind, error) {
	switch name {
	case "ron":
		return Ron, nil
	case "tsumo":
		return Tsumo, nil
	}
	return 0, fmt.Errorf("%w: win kind %q", ErrInvalidSituation, name)
}

// ErrInvalidSituation is returned when a Situation contradicts itself.
var ErrInvalidSituation = errors.New("winning: invalid situation")

// Situation is everything about a win that the hand's tiles cannot tell:
// how it was won, the winds, and the events (riichi, ippatsu, haitei, ...)
// that become yaku. Validate checks the combinations that cannot happen
// together.
type Situation struct {
	WinKind      WinKind
	RoundWind    tile.Wind
	SeatWind     tile.Wind
	Riichi       bool
	DoubleRiichi bool
	Ippatsu      bool
	Haitei       bool
	Houtei       bool
	Rinshan      bool
	Chankan      bool
	Tenhou       bool
	Chiihou      bool
}

// Validate reports the first contradiction in the situation, if any.
func (s Situation) Validate() error {
	switch {
	case s.WinKind != Ron && s.WinKind != Tsumo:
		return fmt.Errorf("%w: win kind %v", ErrInvalidSituation, s.WinKind)
	case !s.RoundWind.IsValid():
		return fmt.Errorf("%w: round wind %v", ErrInvalidSituation, s.RoundWind)
	case !s.SeatWind.IsValid():
		return fmt.Errorf("%w: seat wind %v", ErrInvalidSituation, s.SeatWind)
	case s.Riichi && s.DoubleRiichi:
		return fmt.Errorf("%w: riichi and double riichi together", ErrInvalidSituation)
	case s.Ippatsu && !s.Riichi && !s.DoubleRiichi:
		return fmt.Errorf("%w: ippatsu without riichi", ErrInvalidSituation)
	case s.Haitei && !s.IsTsumo():
		return fmt.Errorf("%w: haitei is a tsumo", ErrInvalidSituation)
	case s.Houtei && !s.IsRon():
		return fmt.Errorf("%w: houtei is a ron", ErrInvalidSituation)
	case s.Rinshan && !s.IsTsumo():
		return fmt.Errorf("%w: rinshan is a tsumo", ErrInvalidSituation)
	case s.Chankan && !s.IsRon():
		return fmt.Errorf("%w: chankan is a ron", ErrInvalidSituation)
	case s.Tenhou && s.Chiihou:
		return fmt.Errorf("%w: tenhou and chiihou together", ErrInvalidSituation)
	case (s.Tenhou || s.Chiihou) && !s.IsTsumo():
		return fmt.Errorf("%w: tenhou/chiihou is a tsumo", ErrInvalidSituation)
	case (s.Tenhou || s.Chiihou) && (s.Riichi || s.DoubleRiichi):
		return fmt.Errorf("%w: tenhou/chiihou with riichi", ErrInvalidSituation)
	case s.Tenhou && !s.IsDealer():
		return fmt.Errorf("%w: tenhou is the dealer's", ErrInvalidSituation)
	case s.Chiihou && s.IsDealer():
		return fmt.Errorf("%w: chiihou is a non-dealer's", ErrInvalidSituation)
	}
	return nil
}

// IsRon reports whether the win was on a discard.
func (s Situation) IsRon() bool {
	return s.WinKind == Ron
}

// IsTsumo reports whether the win was self-drawn.
func (s Situation) IsTsumo() bool {
	return s.WinKind == Tsumo
}

// IsDealer reports whether the winner is the dealer (seat wind east).
func (s Situation) IsDealer() bool {
	return s.SeatWind == tile.EastWind
}

// IsRoundWind reports whether t is the round wind's tile.
func (s Situation) IsRoundWind(t tile.Tile) bool {
	return t == s.RoundWind.Tile()
}

// IsSeatWind reports whether t is the seat wind's tile.
func (s Situation) IsSeatWind(t tile.Tile) bool {
	return t == s.SeatWind.Tile()
}

// IsYakuhai reports whether t is a dragon, the round wind, or the seat wind.
// It decides yakuhai yaku and the pair fu.
func (s Situation) IsYakuhai(t tile.Tile) bool {
	return t.IsDragon() || s.IsRoundWind(t) || s.IsSeatWind(t)
}
