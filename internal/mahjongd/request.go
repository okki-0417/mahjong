// Package mahjongd is the HTTP face of the engine: stateless JSON endpoints
// for hand analysis, scoring, and driving a kyoku against CPU seats.
package mahjongd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// meldRequest is a meld as clients send it: {"kind": "pon", "tiles": [...]}.
type meldRequest struct {
	Kind  string   `json:"kind"`
	Tiles []string `json:"tiles"`
}

// handRequest is the hand part of every analysis request.
type handRequest struct {
	ClosedTiles []string      `json:"closed_tiles"`
	OpenMelds   []meldRequest `json:"open_melds"`
}

func (r handRequest) melds() ([]hand.Meld, error) {
	melds := make([]hand.Meld, 0, len(r.OpenMelds))
	for _, m := range r.OpenMelds {
		kind, err := hand.ParseMeldKind(m.Kind)
		if err != nil {
			return nil, err
		}
		tiles, err := tile.ParseAll(m.Tiles)
		if err != nil {
			return nil, err
		}
		meld, err := hand.NewMeld(kind, tiles)
		if err != nil {
			return nil, err
		}
		melds = append(melds, meld)
	}
	return melds, nil
}

func (r handRequest) hand() (hand.Hand, error) {
	closed, err := tile.ParseAll(r.ClosedTiles)
	if err != nil {
		return hand.Hand{}, err
	}
	melds, err := r.melds()
	if err != nil {
		return hand.Hand{}, err
	}
	return hand.New(closed, melds)
}

// winningRequest adds the winning tile and situation to a hand.
type winningRequest struct {
	handRequest
	WinningTile   string `json:"winning_tile"`
	WinKind       string `json:"win_kind"`
	RoundWind     string `json:"round_wind"`
	SeatWind      string `json:"seat_wind"`
	Riichi        bool   `json:"riichi"`
	DoubleRiichi  bool   `json:"double_riichi"`
	Ippatsu       bool   `json:"ippatsu"`
	Haitei        bool   `json:"haitei"`
	Houtei        bool   `json:"houtei"`
	Rinshan       bool   `json:"rinshan"`
	Chankan       bool   `json:"chankan"`
	Tenhou        bool   `json:"tenhou"`
	Chiihou       bool   `json:"chiihou"`
	DoraCount     int    `json:"dora_count"`
	RoundUpMangan bool   `json:"round_up_mangan"`
	Kuitan        *bool  `json:"kuitan"`
}

func (r winningRequest) winning() (*winning.Winning, error) {
	h, err := r.hand()
	if err != nil {
		return nil, err
	}
	winningTile, err := tile.Parse(r.WinningTile)
	if err != nil {
		return nil, err
	}
	winKind, err := winning.ParseWinKind(r.WinKind)
	if err != nil {
		return nil, err
	}
	roundWind, err := tile.ParseWind(r.RoundWind)
	if err != nil {
		return nil, fmt.Errorf("round_wind: %w", err)
	}
	seatWind, err := tile.ParseWind(r.SeatWind)
	if err != nil {
		return nil, fmt.Errorf("seat_wind: %w", err)
	}
	rules := ruleset.Default().WithRoundUpMangan(r.RoundUpMangan)
	if r.Kuitan != nil {
		rules = rules.WithKuitan(*r.Kuitan)
	}
	return winning.New(h, winningTile, winning.Situation{
		WinKind: winKind, RoundWind: roundWind, SeatWind: seatWind,
		Riichi: r.Riichi, DoubleRiichi: r.DoubleRiichi, Ippatsu: r.Ippatsu,
		Haitei: r.Haitei, Houtei: r.Houtei, Rinshan: r.Rinshan, Chankan: r.Chankan,
		Tenhou: r.Tenhou, Chiihou: r.Chiihou,
	}, rules)
}

// rulesRequest is the table rules; a missing field keeps the default.
type rulesRequest struct {
	Kuitan        *bool `json:"kuitan"`
	RoundUpMangan *bool `json:"round_up_mangan"`
	NagashiMangan *bool `json:"nagashi_mangan"`
	StartingScore *int  `json:"starting_score"`
}

func (r rulesRequest) ruleSet() (ruleset.RuleSet, error) {
	rules := ruleset.Default()
	if r.Kuitan != nil {
		rules = rules.WithKuitan(*r.Kuitan)
	}
	if r.RoundUpMangan != nil {
		rules = rules.WithRoundUpMangan(*r.RoundUpMangan)
	}
	if r.NagashiMangan != nil {
		rules = rules.WithNagashiMangan(*r.NagashiMangan)
	}
	if r.StartingScore != nil {
		var err error
		if rules, err = rules.WithStartingScore(*r.StartingScore); err != nil {
			return ruleset.RuleSet{}, err
		}
	}
	return rules, nil
}

// kyokuRequest describes a kyoku to replay: its wall (or none, to shuffle),
// starting conditions, rules, and recorded actions.
type kyokuRequest struct {
	Wall         []string          `json:"wall"`
	DealerSeat   int               `json:"dealer_seat"`
	RoundWind    string            `json:"round_wind"`
	KyokuNumber  int               `json:"kyoku_number"`
	Scores       *[kyoku.Seats]int `json:"scores"`
	Honba        int               `json:"honba"`
	RiichiSticks int               `json:"riichi_sticks"`
	Rules        rulesRequest      `json:"rules"`
	Actions      []kyoku.Action    `json:"actions"`
}

func (r kyokuRequest) deal() (*kyoku.Kyoku, error) {
	setup := kyoku.Setup{
		DealerSeat: r.DealerSeat, KyokuNumber: r.KyokuNumber, Scores: r.Scores,
		Honba: r.Honba, RiichiSticks: r.RiichiSticks, Actions: r.Actions,
	}
	if len(r.Wall) > 0 {
		tiles, err := tile.ParseAll(r.Wall)
		if err != nil {
			return nil, err
		}
		wall, err := kyoku.NewWall(tiles)
		if err != nil {
			return nil, err
		}
		setup.Wall = &wall
	}
	if r.RoundWind != "" {
		wind, err := tile.ParseWind(r.RoundWind)
		if err != nil {
			return nil, err
		}
		setup.RoundWind = wind
	}
	rules, err := r.Rules.ruleSet()
	if err != nil {
		return nil, err
	}
	setup.RuleSet = rules
	return kyoku.Deal(setup)
}

// errBadRequest marks input that could not be read at all, as opposed to
// input the domain rejects.
var errBadRequest = errors.New("mahjongd: bad request")

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", errBadRequest, err)
	}
	return nil
}
