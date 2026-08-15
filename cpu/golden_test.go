package cpu_test

import (
	"testing"

	"github.com/okki-0417/mahjong/cpu"
	"github.com/okki-0417/mahjong/internal/goldentest"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

type goldenCPUKyoku struct {
	Wall  []string `json:"wall"`
	Setup struct {
		DealerSeat   int    `json:"dealer_seat"`
		RoundWind    string `json:"round_wind"`
		KyokuNumber  int    `json:"kyoku_number"`
		Honba        int    `json:"honba"`
		RiichiSticks int    `json:"riichi_sticks"`
		Scores       [4]int `json:"scores"`
	} `json:"setup"`
	RuleSet struct {
		Kuitan        bool `json:"kuitan"`
		RoundUpMangan bool `json:"round_up_mangan"`
		NagashiMangan bool `json:"nagashi_mangan"`
	} `json:"rule_set"`
	Actions    []kyoku.Action `json:"actions"`
	ResultKind string         `json:"result_kind"`
}

// TestGolden replays kyokus played by the Ruby simulator's CPU at all four
// seats and checks that Choose makes the same choice at every step.
func TestGolden(t *testing.T) {
	goldentest.Each(t, "cpu_kyoku.jsonl", func(line int, g goldenCPUKyoku) {
		wallTiles, err := tile.ParseAll(g.Wall)
		if err != nil {
			t.Fatal(err)
		}
		wall := kyoku.MustWall(wallTiles)
		roundWind, _ := tile.ParseWind(g.Setup.RoundWind)
		scores := g.Setup.Scores
		k, err := kyoku.Deal(kyoku.Setup{
			Wall: &wall, DealerSeat: g.Setup.DealerSeat, RoundWind: roundWind, KyokuNumber: g.Setup.KyokuNumber,
			Scores: &scores, Honba: g.Setup.Honba, RiichiSticks: g.Setup.RiichiSticks,
			RuleSet: ruleset.Default().WithKuitan(g.RuleSet.Kuitan).WithRoundUpMangan(g.RuleSet.RoundUpMangan).WithNagashiMangan(g.RuleSet.NagashiMangan),
		})
		if err != nil {
			t.Fatal(err)
		}
		for i, want := range g.Actions {
			seat := k.AwaitingSeats()[0]
			got, ok := cpu.Choose(k.SeenBy(seat))
			if !ok || got != want {
				t.Errorf("line %d step %d: chose %v, want %v", line, i, got, want)
				return
			}
			if k, err = k.Take(got); err != nil {
				t.Fatalf("line %d step %d: %v", line, i, err)
			}
		}
		if r, ok := k.Result(); !ok || r.Kind().String() != g.ResultKind {
			t.Errorf("line %d: result %v, want %s", line, r, g.ResultKind)
		}
	})
}
