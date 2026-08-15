package kyoku_test

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

const (
	ittsu = "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"
	noten = "1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 2z 3z 4z"
)

func seededWall() kyoku.Wall {
	return kyoku.ShuffledWall(rand.New(rand.NewPCG(7, 7)), true)
}

func TestNewWall(t *testing.T) {
	t.Run("accepts an ordering of the full set", func(t *testing.T) {
		if _, err := kyoku.NewWall(tile.FullSet(true)); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rejects the wrong count", func(t *testing.T) {
		if _, err := kyoku.NewWall(tile.FullSet(true)[:135]); !errors.Is(err, kyoku.ErrInvalidWall) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects a set that is not four of each kind", func(t *testing.T) {
		tiles := tile.FullSet(true)
		tiles[0] = tile.M2
		if _, err := kyoku.NewWall(tiles); !errors.Is(err, kyoku.ErrInvalidWall) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects an invalid tile", func(t *testing.T) {
		tiles := tile.FullSet(true)
		tiles[0] = 0
		if _, err := kyoku.NewWall(tiles); !errors.Is(err, kyoku.ErrInvalidWall) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("MustWall panics on error", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		kyoku.MustWall(nil)
	})
	t.Run("Tiles returns a copy", func(t *testing.T) {
		w := seededWall()
		w.Tiles()[0] = 0
		if !w.Tiles()[0].IsValid() {
			t.Fatal("shared")
		}
	})
	t.Run("ShuffledWall with nil uses the package source", func(t *testing.T) {
		if len(kyoku.ShuffledWall(nil, false).Tiles()) != 136 {
			t.Fatal("size")
		}
	})
}

func TestDeal(t *testing.T) {
	t.Run("defaults to seat 0 dealing east 1 at the starting score", func(t *testing.T) {
		k, err := kyoku.Deal(kyoku.Setup{})
		if err != nil {
			t.Fatal(err)
		}
		km := k.Kyokumen()
		if km.DealerSeat() != 0 || km.RoundWind() != tile.EastWind || km.KyokuNumber() != 1 || km.Honba() != 0 || km.RiichiSticks() != 0 || km.Score(2) != 25000 {
			t.Fatalf("got %+v", km)
		}
		if km.RemainingDraws() != 70 || len(km.DoraIndicators()) != 1 || !km.IsOpening() {
			t.Fatal("opening")
		}
	})
	t.Run("keeps the wall it was dealt from", func(t *testing.T) {
		w := seededWall()
		k, _ := kyoku.Deal(kyoku.Setup{Wall: &w})
		if !reflect.DeepEqual(k.Wall().Tiles(), w.Tiles()) {
			t.Fatal("wall")
		}
	})
	t.Run("deals from the dealer on", func(t *testing.T) {
		w := seededWall()
		k, _ := kyoku.Deal(kyoku.Setup{Wall: &w, DealerSeat: 2})
		if !reflect.DeepEqual(k.Kyokumen().Seat(2).Hand().ClosedTiles(), w.Hands()[0]) {
			t.Fatal("dealer hand")
		}
		seat, _ := k.Kyokumen().DiscardingSeat()
		if seat != 2 || k.Kyokumen().SeatWind(2) != tile.EastWind || k.Kyokumen().SeatWind(3) != tile.SouthWind {
			t.Fatal("dealer turn or winds")
		}
	})
	t.Run("uses the rule set's starting score", func(t *testing.T) {
		rs, _ := ruleset.Default().WithStartingScore(30000)
		k, _ := kyoku.Deal(kyoku.Setup{RuleSet: rs})
		if k.Kyokumen().Score(0) != 30000 || k.Kyokumen().RuleSet() != rs {
			t.Fatal("scores")
		}
	})
	t.Run("assigns given scores by seat", func(t *testing.T) {
		scores := [4]int{1, 2, 3, 4}
		k, _ := kyoku.Deal(kyoku.Setup{Scores: &scores})
		if k.Kyokumen().Score(3) != 4 {
			t.Fatal("scores")
		}
	})
	t.Run("passes the round conditions through", func(t *testing.T) {
		k, _ := kyoku.Deal(kyoku.Setup{DealerSeat: 1, RoundWind: tile.SouthWind, KyokuNumber: 3, Honba: 2, RiichiSticks: 1})
		km := k.Kyokumen()
		if km.DealerSeat() != 1 || km.RoundWind() != tile.SouthWind || km.KyokuNumber() != 3 || km.Honba() != 2 || km.RiichiSticks() != 1 {
			t.Fatal("conditions")
		}
	})
	rejected := []struct {
		name  string
		setup kyoku.Setup
	}{
		{"a dealer seat out of range", kyoku.Setup{DealerSeat: 4}},
		{"an invalid round wind", kyoku.Setup{RoundWind: 9}},
		{"a kyoku number out of range", kyoku.Setup{KyokuNumber: 5}},
		{"a negative honba", kyoku.Setup{Honba: -1}},
		{"negative riichi sticks", kyoku.Setup{RiichiSticks: -1}},
	}
	for _, c := range rejected {
		t.Run("rejects "+c.name, func(t *testing.T) {
			if _, err := kyoku.Deal(c.setup); !errors.Is(err, kyoku.ErrInvalidSetup) {
				t.Fatalf("err = %v", err)
			}
		})
	}
	t.Run("replays recorded actions and rejects an illegal one", func(t *testing.T) {
		w := seededWall()
		played, _ := kyoku.Deal(kyoku.Setup{Wall: &w})
		for i := 0; i < 8; i++ {
			played, _ = played.Take(played.Kyokumen().LegalActions()[0])
		}
		replayed, err := kyoku.Deal(kyoku.Setup{Wall: &w, Actions: played.Actions()})
		if err != nil || !reflect.DeepEqual(replayed.Kyokumen(), played.Kyokumen()) {
			t.Fatalf("replay: %v", err)
		}
		if _, err := kyoku.Deal(kyoku.Setup{Wall: &w, Actions: []kyoku.Action{kyoku.NewTsumo(3)}}); !errors.Is(err, kyoku.ErrIllegalAction) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestTake(t *testing.T) {
	k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsu}, Draws: "5s"})

	t.Run("appends the action and folds it", func(t *testing.T) {
		next, err := k.Take(kyoku.NewTsumo(0))
		if err != nil {
			t.Fatal(err)
		}
		if len(next.Actions()) != 1 || !next.IsFinished() {
			t.Fatal("not folded")
		}
	})
	t.Run("leaves the original untouched", func(t *testing.T) {
		if _, err := k.Take(kyoku.NewTsumo(0)); err != nil {
			t.Fatal(err)
		}
		if len(k.Actions()) != 0 || k.IsFinished() {
			t.Fatal("mutated")
		}
	})
	t.Run("rejects an action that is not legal now", func(t *testing.T) {
		if _, err := k.Take(kyoku.NewDiscard(0, tile.Chun)); !errors.Is(err, kyoku.ErrIllegalAction) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects another seat's action", func(t *testing.T) {
		if _, err := k.Take(kyoku.NewDiscard(1, tile.M1)); !errors.Is(err, kyoku.ErrIllegalAction) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects anything after the end", func(t *testing.T) {
		done, _ := k.Take(kyoku.NewTsumo(0))
		if _, err := done.Take(kyoku.NewDiscard(0, tile.S5)); !errors.Is(err, kyoku.ErrFinished) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("accepts every offered action", func(t *testing.T) {
		for _, a := range k.Kyokumen().LegalActions() {
			if _, err := k.Take(a); err != nil {
				t.Errorf("%v: %v", a, err)
			}
		}
	})
	t.Run("Actions returns a copy", func(t *testing.T) {
		next, _ := k.Take(kyoku.NewDiscard(0, tile.S5))
		next.Actions()[0] = kyoku.NewPass(3)
		if next.Actions()[0].Kind() != kyoku.ActionDiscard {
			t.Fatal("shared")
		}
	})
}

func TestAwaitingSeats(t *testing.T) {
	t.Run("only the seat on turn", func(t *testing.T) {
		if got := mt.BuildKyoku(mt.KyokuSpec{}).AwaitingSeats(); !reflect.DeepEqual(got, []int{0}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("every seat that can answer a discard", func(t *testing.T) {
		k := mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 2: ittsu, 3: "5s 0s 1m 2m 3m 4m 6m 7m 8m 9m 1p 2p 3p"},
			Draws: "5s", Actions: []kyoku.Action{kyoku.NewDiscard(0, tile.S5)},
		})
		if got := k.AwaitingSeats(); !reflect.DeepEqual(got, []int{2, 3}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("nobody after the end", func(t *testing.T) {
		k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsu}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}})
		if len(k.AwaitingSeats()) != 0 {
			t.Fatal("awaiting")
		}
	})
	t.Run("nobody after an exhaustive draw, though the last window stays open", func(t *testing.T) {
		hands := map[int]string{}
		for s := 0; s < 4; s++ {
			hands[s] = noten
		}
		k := mt.BuildKyoku(mt.KyokuSpec{Hands: hands, Wall: mt.WallOf(0)})
		if _, open := k.Kyokumen().ClaimedTile(); !open || len(k.AwaitingSeats()) != 0 || !k.IsFinished() {
			t.Fatal("awaiting")
		}
	})
}

func TestResultAccessors(t *testing.T) {
	won := mt.BuildKyoku(mt.KyokuSpec{
		Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 2: ittsu},
		Draws: "5s", Actions: []kyoku.Action{kyoku.NewDiscard(0, tile.S5), kyoku.NewRon(2)},
	})
	r, ok := won.Result()
	if !ok {
		t.Fatal("no result")
	}
	if r.Kind() != kyoku.ResultRon || !r.Kind().IsWin() {
		t.Error("kind")
	}
	if w, ok := r.Winner(); !ok || w != 2 {
		t.Error("winner")
	}
	if l, ok := r.Loser(); !ok || l != 0 {
		t.Error("loser")
	}
	score, ok := r.Score()
	if !ok || score.Total() != r.Deltas()[2] || r.Scores()[2] != 25000+r.Deltas()[2] {
		t.Error("score")
	}
	hands := r.RevealedHands()
	if len(hands) != 1 || hands[0].Seat != 2 || !hands[0].HasWinning || hands[0].WinningTile != tile.S5 {
		t.Errorf("revealed %+v", hands)
	}
	if len(r.DoraIndicators()) != 1 || r.UradoraIndicators() != nil {
		t.Error("indicators")
	}
	if r.DealerContinues() || r.NextHonba() != 0 || r.CarriedRiichiSticks() != 0 {
		t.Error("continuation")
	}

	drawn := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"}, Draws: "1m", Actions: []kyoku.Action{kyoku.NewKyushukyuhai(0)}})
	dr, _ := drawn.Result()
	if _, ok := dr.Winner(); ok || dr.Kind() != kyoku.ResultKyushukyuhai || dr.Kind().IsWin() {
		t.Error("draw result")
	}
	if _, ok := dr.Score(); ok {
		t.Error("draw scored")
	}
	if len(dr.RevealedHands()) != 0 {
		t.Error("draw revealed hands")
	}
}

func TestDealNext(t *testing.T) {
	t.Run("refuses while the kyoku is going", func(t *testing.T) {
		if _, err := mt.BuildKyoku(mt.KyokuSpec{}).DealNext(); !errors.Is(err, kyoku.ErrNotFinished) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("hands the deal to the next seat after a child's win", func(t *testing.T) {
		k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{1: ittsu}, Draws: "5z 5s", Actions: []kyoku.Action{kyoku.NewDiscard(0, tile.Haku), kyoku.NewTsumo(1)}})
		next, err := k.DealNext()
		if err != nil {
			t.Fatal(err)
		}
		km := next.Kyokumen()
		r, _ := k.Result()
		if km.DealerSeat() != 1 || km.KyokuNumber() != 2 || km.Honba() != 0 || km.Score(1) != r.Scores()[1] {
			t.Fatalf("got %+v", km)
		}
	})
	t.Run("IsAllLast is the last dealer's turn of the last wind", func(t *testing.T) {
		if !mt.BuildKyoku(mt.KyokuSpec{Dealer: 0, RoundWind: tile.SouthWind, KyokuNumber: 4}).IsAllLast() {
			t.Error("south 4 is all last")
		}
		if !mt.BuildKyoku(mt.KyokuSpec{Dealer: 3, RoundWind: tile.SouthWind, KyokuNumber: 4}).IsAllLast() {
			t.Error("south 4 is all last whoever started as dealer")
		}
		if mt.BuildKyoku(mt.KyokuSpec{Dealer: 0, RoundWind: tile.SouthWind, KyokuNumber: 3}).IsAllLast() {
			t.Error("south 3 is not all last")
		}
		if mt.BuildKyoku(mt.KyokuSpec{Dealer: 3, RoundWind: tile.EastWind, KyokuNumber: 4}).IsAllLast() {
			t.Error("east 4 is not all last")
		}
	})
}

func TestSightJSON(t *testing.T) {
	k := mt.BuildKyoku(mt.KyokuSpec{
		Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
		Draws: "5z", Actions: []kyoku.Action{kyoku.NewDiscard(0, tile.M2), kyoku.NewPass(1), kyoku.NewPon(2, tile.M2, mt.Tiles("2m 2m"))},
	})
	raw, err := json.Marshal(k.SeenBy(2))
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"seat", "closed_tiles", "drawn", "legal_actions", "seats", "dealer_seat", "round_wind", "kyoku_number", "honba", "riichi_sticks", "junme", "remaining_draws", "dora_indicators", "discarding_seat", "claimed_tile", "claimed_from"} {
		if _, ok := v[key]; !ok {
			t.Errorf("missing %s", key)
		}
	}
	if v["drawn"] != nil || v["discarding_seat"] != float64(2) || len(v["closed_tiles"].([]any)) != 11 {
		t.Errorf("drawn %v discarding %v closed %v", v["drawn"], v["discarding_seat"], v["closed_tiles"])
	}
	seats := v["seats"].([]any)
	if seats[2].(map[string]any)["pending_call"] == nil || seats[0].(map[string]any)["pending_call"] != nil {
		t.Error("pending_call")
	}
	discard := seats[0].(map[string]any)["discards"].([]any)[0].(map[string]any)
	if discard["called_by"] != float64(2) || discard["tile"] != "2m" {
		t.Errorf("discard %v", discard)
	}
}
