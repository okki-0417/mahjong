package winning_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/internal/goldentest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

type goldenMentsu struct {
	Kind  string   `json:"kind"`
	Tiles []string `json:"tiles"`
	Open  bool     `json:"open"`
}

type goldenForm struct {
	Kind      string         `json:"kind"`
	WaitKind  string         `json:"wait_kind"`
	Menzen    bool           `json:"menzen"`
	Mentsu    []goldenMentsu `json:"mentsu"`
	PairTile  *string        `json:"pair_tile"`
	PairTiles []string       `json:"pair_tiles"`
}

type goldenSituation struct {
	WinKind      string `json:"win_kind"`
	RoundWind    string `json:"round_wind"`
	SeatWind     string `json:"seat_wind"`
	Riichi       bool   `json:"riichi"`
	DoubleRiichi bool   `json:"double_riichi"`
	Ippatsu      bool   `json:"ippatsu"`
	Haitei       bool   `json:"haitei"`
	Houtei       bool   `json:"houtei"`
	Rinshan      bool   `json:"rinshan"`
	Chankan      bool   `json:"chankan"`
	Tenhou       bool   `json:"tenhou"`
	Chiihou      bool   `json:"chiihou"`
}

type goldenFuSource struct {
	Kind  string   `json:"kind"`
	Label string   `json:"label"`
	Fu    int      `json:"fu"`
	Tiles []string `json:"tiles"`
}

type goldenFu struct {
	Subtotal int              `json:"subtotal"`
	Total    int              `json:"total"`
	Sources  []goldenFuSource `json:"sources"`
	Form     goldenForm       `json:"form"`
}

type goldenYaku struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Han     int    `json:"han"`
	Yakuman int    `json:"yakuman"`
}

type goldenScore struct {
	Yakus        []goldenYaku `json:"yakus"`
	Han          int          `json:"han"`
	Fu           int          `json:"fu"`
	YakumanCount int          `json:"yakuman_count"`
	DoraCount    int          `json:"dora_count"`
	BasePoints   int          `json:"base_points"`
	Total        int          `json:"total"`
	RankLabel    *string      `json:"rank_label"`
	Payments     struct {
		FromLoser     *int `json:"from_loser"`
		FromDealer    *int `json:"from_dealer"`
		FromNonDealer *int `json:"from_non_dealer"`
	} `json:"payments"`
	Form goldenForm `json:"form"`
}

type goldenWinning struct {
	Closed      []string          `json:"closed"`
	Melds       []goldentest.Meld `json:"melds"`
	WinningTile string            `json:"winning_tile"`
	Situation   goldenSituation   `json:"situation"`
	DoraCount   int               `json:"dora_count"`
	RuleSet     struct {
		Kuitan        bool `json:"kuitan"`
		RoundUpMangan bool `json:"round_up_mangan"`
	} `json:"rule_set"`
	Forms []goldenForm `json:"forms"`
	Fu    *goldenFu    `json:"fu"`
	Score *goldenScore `json:"score"`
	Error string       `json:"error"`
}

func (g goldenSituation) situation(t *testing.T) winning.Situation {
	t.Helper()
	winKind, err := winning.ParseWinKind(g.WinKind)
	if err != nil {
		t.Fatal(err)
	}
	roundWind, err := tile.ParseWind(g.RoundWind)
	if err != nil {
		t.Fatal(err)
	}
	seatWind, err := tile.ParseWind(g.SeatWind)
	if err != nil {
		t.Fatal(err)
	}
	return winning.Situation{
		WinKind: winKind, RoundWind: roundWind, SeatWind: seatWind,
		Riichi: g.Riichi, DoubleRiichi: g.DoubleRiichi, Ippatsu: g.Ippatsu,
		Haitei: g.Haitei, Houtei: g.Houtei, Rinshan: g.Rinshan, Chankan: g.Chankan,
		Tenhou: g.Tenhou, Chiihou: g.Chiihou,
	}
}

func formRecord(f *winning.Form) goldenForm {
	g := goldenForm{Kind: f.Kind().String(), WaitKind: f.WaitKind().String(), Menzen: f.IsMenzen()}
	for _, m := range f.Mentsu() {
		g.Mentsu = append(g.Mentsu, goldenMentsu{Kind: m.Kind().String(), Tiles: tile.Labels(m.Tiles()), Open: m.IsOpen()})
	}
	if pair, ok := f.PairTile(); ok {
		label := pair.String()
		g.PairTile = &label
	}
	g.PairTiles = tile.Labels(f.PairTiles())
	return g
}

func sameForm(got, want goldenForm) bool {
	if got.Kind != want.Kind || got.WaitKind != want.WaitKind || got.Menzen != want.Menzen {
		return false
	}
	if len(got.Mentsu) != len(want.Mentsu) {
		return false
	}
	for i := range got.Mentsu {
		if got.Mentsu[i].Kind != want.Mentsu[i].Kind || got.Mentsu[i].Open != want.Mentsu[i].Open ||
			!goldentest.SameLabels(got.Mentsu[i].Tiles, want.Mentsu[i].Tiles) {
			return false
		}
	}
	if (got.PairTile == nil) != (want.PairTile == nil) || (got.PairTile != nil && *got.PairTile != *want.PairTile) {
		return false
	}
	return goldentest.SameLabels(got.PairTiles, want.PairTiles)
}

func intOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// TestGolden replays wins recorded from the Ruby implementation and checks
// forms, fu, and score agree.
func TestGolden(t *testing.T) {
	goldentest.Each(t, "winning.jsonl", func(line int, g goldenWinning) {
		h := goldentest.Hand(t, g.Closed, g.Melds)
		winningTile := tile.MustParse(g.WinningTile)
		s := g.Situation.situation(t)
		rs := ruleset.Default().WithKuitan(g.RuleSet.Kuitan).WithRoundUpMangan(g.RuleSet.RoundUpMangan)

		forms := winning.Forms(h, winningTile, winning.Situation{})
		if len(forms) != len(g.Forms) {
			t.Errorf("line %d %v+%v: %d forms, want %d", line, h, winningTile, len(forms), len(g.Forms))
		} else {
			for i := range forms {
				if got := formRecord(&forms[i]); !sameForm(got, g.Forms[i]) {
					t.Errorf("line %d %v+%v: form %d = %+v, want %+v", line, h, winningTile, i, got, g.Forms[i])
				}
			}
		}

		w, err := winning.New(h, winningTile, s, rs)
		switch g.Error {
		case "no_form":
			if !errors.Is(err, winning.ErrNoForm) {
				t.Errorf("line %d %v+%v: err = %v, want ErrNoForm", line, h, winningTile, err)
			}
			return
		case "no_yaku":
			if !errors.Is(err, winning.ErrNoYaku) {
				t.Errorf("line %d %v+%v: err = %v, want ErrNoYaku", line, h, winningTile, err)
			}
			return
		case "argument":
			if err == nil || errors.Is(err, winning.ErrNotWinning) {
				t.Errorf("line %d %v+%v: err = %v, want an argument error", line, h, winningTile, err)
			}
			return
		}
		if err != nil {
			t.Errorf("line %d %v+%v: %v", line, h, winningTile, err)
			return
		}

		fu := w.Fu()
		if fu.Subtotal() != g.Fu.Subtotal || fu.Total() != g.Fu.Total {
			t.Errorf("line %d %v+%v: fu %d/%d, want %d/%d", line, h, winningTile, fu.Subtotal(), fu.Total(), g.Fu.Subtotal, g.Fu.Total)
		}
		var sources []goldenFuSource
		for _, src := range fu.Sources() {
			sources = append(sources, goldenFuSource{Kind: string(src.Kind), Label: src.Label, Fu: src.Fu, Tiles: labelsOrNil(src.Tiles)})
		}
		if len(sources)+len(g.Fu.Sources) > 0 && !reflect.DeepEqual(sources, g.Fu.Sources) {
			t.Errorf("line %d %v+%v: fu sources %+v, want %+v", line, h, winningTile, sources, g.Fu.Sources)
		}
		if got := formRecord(fu.Form()); !sameForm(got, g.Fu.Form) {
			t.Errorf("line %d %v+%v: fu form %+v, want %+v", line, h, winningTile, got, g.Fu.Form)
		}

		score := w.Score(g.DoraCount)
		var yakus []goldenYaku
		for _, y := range score.Yakus() {
			yakus = append(yakus, goldenYaku{ID: string(y.ID), Name: y.Name, Han: y.Han, Yakuman: y.Yakuman})
		}
		if !reflect.DeepEqual(yakus, g.Score.Yakus) {
			t.Errorf("line %d %v+%v: yakus %+v, want %+v", line, h, winningTile, yakus, g.Score.Yakus)
		}
		if score.Han() != g.Score.Han || score.Fu() != g.Score.Fu || score.YakumanCount() != g.Score.YakumanCount ||
			score.DoraCount() != g.Score.DoraCount || score.BasePoints() != g.Score.BasePoints || score.Total() != g.Score.Total {
			t.Errorf("line %d %v+%v: score han %d fu %d yakuman %d dora %d base %d total %d, want %d %d %d %d %d %d",
				line, h, winningTile, score.Han(), score.Fu(), score.YakumanCount(), score.DoraCount(), score.BasePoints(), score.Total(),
				g.Score.Han, g.Score.Fu, g.Score.YakumanCount, g.Score.DoraCount, g.Score.BasePoints, g.Score.Total)
		}
		wantRank := ""
		if g.Score.RankLabel != nil {
			wantRank = *g.Score.RankLabel
		}
		if score.RankLabel() != wantRank {
			t.Errorf("line %d %v+%v: rank %q, want %q", line, h, winningTile, score.RankLabel(), wantRank)
		}
		p := score.Payments()
		if p.FromLoser != intOrZero(g.Score.Payments.FromLoser) || p.FromDealer != intOrZero(g.Score.Payments.FromDealer) ||
			p.FromNonDealer != intOrZero(g.Score.Payments.FromNonDealer) {
			t.Errorf("line %d %v+%v: payments %+v, want %+v", line, h, winningTile, p, g.Score.Payments)
		}
		if got := formRecord(score.Form()); !sameForm(got, g.Score.Form) {
			t.Errorf("line %d %v+%v: score form %+v, want %+v", line, h, winningTile, got, g.Score.Form)
		}
	})
}

func labelsOrNil(tiles []tile.Tile) []string {
	if len(tiles) == 0 {
		return nil
	}
	return tile.Labels(tiles)
}
