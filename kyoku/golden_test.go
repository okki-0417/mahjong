package kyoku_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/internal/goldentest"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

type goldenSetup struct {
	DealerSeat   int    `json:"dealer_seat"`
	RoundWind    string `json:"round_wind"`
	KyokuNumber  int    `json:"kyoku_number"`
	Honba        int    `json:"honba"`
	RiichiSticks int    `json:"riichi_sticks"`
	Scores       [4]int `json:"scores"`
}

type goldenStep struct {
	Awaiting []int             `json:"awaiting"`
	Legal    []json.RawMessage `json:"legal"`
	Sight    json.RawMessage   `json:"sight"`
	Chosen   kyoku.Action      `json:"chosen"`
}

type goldenKyoku struct {
	Wall    []string    `json:"wall"`
	Setup   goldenSetup `json:"setup"`
	RuleSet struct {
		Kuitan        bool `json:"kuitan"`
		RoundUpMangan bool `json:"round_up_mangan"`
		NagashiMangan bool `json:"nagashi_mangan"`
	} `json:"rule_set"`
	Steps   []goldenStep      `json:"steps"`
	Result  json.RawMessage   `json:"result"`
	AllLast bool              `json:"all_last"`
	Last    bool              `json:"last"`
	Sights  []json.RawMessage `json:"sights"`
	Next    *goldenSetup      `json:"next"`
}

func decode(t *testing.T, raw []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func encode(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return decode(t, raw)
}

func resultRecord(r *kyoku.Result) map[string]any {
	record := map[string]any{
		"kind":                  r.Kind().String(),
		"winner":                nil,
		"loser":                 nil,
		"deltas":                r.Deltas(),
		"scores":                r.Scores(),
		"dealer_continues":      r.DealerContinues(),
		"next_honba":            r.NextHonba(),
		"carried_riichi_sticks": r.CarriedRiichiSticks(),
		"revealed_hands":        emptyIfNilHands(r.RevealedHands()),
		"dora_indicators":       tile.Labels(r.DoraIndicators()),
		"uradora_indicators":    emptyLabels(r.UradoraIndicators()),
	}
	if w, ok := r.Winner(); ok {
		record["winner"] = w
	}
	if l, ok := r.Loser(); ok {
		record["loser"] = l
	}
	if score, ok := r.Score(); ok {
		yakus := []map[string]any{}
		for _, y := range score.Yakus() {
			yakus = append(yakus, map[string]any{"id": string(y.ID), "name": y.Name, "han": y.Han, "yakuman": y.Yakuman})
		}
		var rank any
		if score.RankLabel() != "" {
			rank = score.RankLabel()
		}
		record["score"] = map[string]any{
			"yakus": yakus, "han": score.Han(), "fu": score.Fu(), "yakuman_count": score.YakumanCount(),
			"dora_count": score.DoraCount(), "total": score.Total(), "rank_label": rank,
		}
	}
	return record
}

func emptyIfNilHands(hands []kyoku.RevealedHand) []kyoku.RevealedHand {
	if hands == nil {
		return []kyoku.RevealedHand{}
	}
	return hands
}

func emptyLabels(tiles []tile.Tile) []string {
	if len(tiles) == 0 {
		return []string{}
	}
	return tile.Labels(tiles)
}

func setupRecord(k *kyoku.Kyoku) goldenSetup {
	km := k.Kyokumen()
	var scores [4]int
	for i := range scores {
		scores[i] = km.Score(i)
	}
	return goldenSetup{
		DealerSeat: km.DealerSeat(), RoundWind: km.RoundWind().String(), KyokuNumber: km.KyokuNumber(),
		Honba: km.Honba(), RiichiSticks: km.RiichiSticks(), Scores: scores,
	}
}

// TestGolden replays hanchans recorded from the Ruby implementation: every
// legal action list, the acting seat's view, the result, and the following
// kyoku's setup must agree.
func TestGolden(t *testing.T) {
	goldentest.Each(t, "kyoku.jsonl", func(line int, g goldenKyoku) {
		wallTiles, err := tile.ParseAll(g.Wall)
		if err != nil {
			t.Fatal(err)
		}
		wall := kyoku.MustWall(wallTiles)
		roundWind, err := tile.ParseWind(g.Setup.RoundWind)
		if err != nil {
			t.Fatal(err)
		}
		rules := ruleset.Default().WithKuitan(g.RuleSet.Kuitan).WithRoundUpMangan(g.RuleSet.RoundUpMangan).WithNagashiMangan(g.RuleSet.NagashiMangan)
		scores := g.Setup.Scores
		k, err := kyoku.Deal(kyoku.Setup{
			Wall: &wall, DealerSeat: g.Setup.DealerSeat, RoundWind: roundWind, KyokuNumber: g.Setup.KyokuNumber,
			Scores: &scores, Honba: g.Setup.Honba, RiichiSticks: g.Setup.RiichiSticks, RuleSet: rules,
		})
		if err != nil {
			t.Fatalf("line %d: %v", line, err)
		}

		for i, step := range g.Steps {
			if !reflect.DeepEqual(k.AwaitingSeats(), step.Awaiting) {
				t.Errorf("line %d step %d: awaiting %v, want %v", line, i, k.AwaitingSeats(), step.Awaiting)
			}
			legal := k.Kyokumen().LegalActions()
			if len(legal) != len(step.Legal) {
				t.Errorf("line %d step %d: %d legal actions, want %d\n got %v", line, i, len(legal), len(step.Legal), legal)
				return
			}
			for j := range legal {
				if !reflect.DeepEqual(encode(t, legal[j]), decode(t, step.Legal[j])) {
					t.Errorf("line %d step %d: legal[%d] = %v, want %s", line, i, j, legal[j], step.Legal[j])
					return
				}
			}
			if step.Sight != nil {
				got := encode(t, k.SeenBy(step.Chosen.Seat()))
				if !reflect.DeepEqual(got, decode(t, step.Sight)) {
					t.Errorf("line %d step %d: sight differs\n got  %v\n want %s", line, i, got, step.Sight)
					return
				}
			}
			next, err := k.Take(step.Chosen)
			if err != nil {
				t.Errorf("line %d step %d: take %v: %v", line, i, step.Chosen, err)
				return
			}
			k = next
		}

		result, ok := k.Result()
		if !ok {
			t.Errorf("line %d: not finished after %d steps", line, len(g.Steps))
			return
		}
		if got, want := encode(t, resultRecord(result)), decode(t, g.Result); !reflect.DeepEqual(got, want) {
			t.Errorf("line %d: result\n got  %v\n want %v", line, got, want)
		}
		if k.IsAllLast() != g.AllLast || k.IsLast() != g.Last {
			t.Errorf("line %d: all_last %v last %v, want %v %v", line, k.IsAllLast(), k.IsLast(), g.AllLast, g.Last)
		}
		for seat, raw := range g.Sights {
			if got := encode(t, k.SeenBy(seat)); !reflect.DeepEqual(got, decode(t, raw)) {
				t.Errorf("line %d: final sight of seat %d differs\n got  %v\n want %s", line, seat, got, raw)
			}
		}
		if g.Next != nil {
			next, err := k.DealNext()
			if err != nil {
				t.Errorf("line %d: DealNext: %v", line, err)
				return
			}
			if got := setupRecord(next); got != *g.Next {
				t.Errorf("line %d: next setup %+v, want %+v", line, got, *g.Next)
			}
		} else if _, err := k.DealNext(); err == nil {
			t.Errorf("line %d: DealNext succeeded after the last kyoku", line)
		}
	})
}
