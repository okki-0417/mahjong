package winning_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

func southRon() winning.Situation {
	return winning.Situation{WinKind: winning.Ron, RoundWind: tile.EastWind, SeatWind: tile.SouthWind}
}

func build(t *testing.T, closed, win string, melds ...hand.Meld) *winning.Winning {
	t.Helper()
	w, err := winning.New(mt.Hand(closed, melds...), mt.T(win), southRon(), ruleset.Default())
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func sourceKinds(fu winning.Fu) []winning.FuSourceKind {
	var kinds []winning.FuSourceKind
	for _, s := range fu.Sources() {
		kinds = append(kinds, s.Kind)
	}
	return kinds
}

func hasKind(kinds []winning.FuSourceKind, kind winning.FuSourceKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func TestNew(t *testing.T) {
	t.Run("accepts 13 tiles and a winning tile", func(t *testing.T) {
		build(t, "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s", "5s")
	})
	t.Run("keeps menzen with an ankan and breaks it with a minkan", func(t *testing.T) {
		ankan := build(t, "2m 3m 4m 5p 6p 7p 1s 2s 3s 9s", "9s", mt.Ankan("5z 5z 5z 5z"))
		minkan := build(t, "2m 3m 4m 5p 6p 7p 1s 2s 3s 9s", "9s", mt.Minkan("5z 5z 5z 5z"))
		if !hasKind(sourceKinds(ankan.Fu()), winning.FuMenzenRon) || hasKind(sourceKinds(minkan.Fu()), winning.FuMenzenRon) {
			t.Fatal("menzen ron fu")
		}
	})
	t.Run("rejects an invalid winning tile", func(t *testing.T) {
		_, err := winning.New(mt.Hand("1m 2m 3m 4m 5p 6p 7p 1s 2s 3s 5s 5s 7z"), 0, southRon(), ruleset.Default())
		if !errors.Is(err, hand.ErrInvalidTile) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects a contradictory situation", func(t *testing.T) {
		s := southRon()
		s.Ippatsu = true
		_, err := winning.New(mt.Hand("1m 2m 3m 4m 5p 6p 7p 1s 2s 3s 5s 5s 7z"), mt.T("7z"), s, ruleset.Default())
		if !errors.Is(err, winning.ErrInvalidSituation) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects a winning tile that would be a fifth copy", func(t *testing.T) {
		_, err := winning.New(mt.Hand("1m 1m 1m 1m 5p 6p 7p 1s 2s 3s 5s 5s 7z"), mt.T("1m"), southRon(), ruleset.Default())
		if !errors.Is(err, winning.ErrTileExhausted) || errors.Is(err, winning.ErrNotWinning) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("returns ErrNoForm for an incomplete hand", func(t *testing.T) {
		_, err := winning.New(mt.Hand("1m 3m 5m 7m 9m 1p 3p 5p 7p 9p 1s 3s 5s"), mt.T("7z"), southRon(), ruleset.Default())
		if !errors.Is(err, winning.ErrNoForm) || !errors.Is(err, winning.ErrNotWinning) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("returns ErrNoYaku for a complete hand without yaku", func(t *testing.T) {
		_, err := winning.New(mt.Hand("1p 2p 3p 4p 5p 6p 7s 8s 9s 2z", mt.Chi("1m 2m 3m")), mt.T("2z"), southRon(), ruleset.Default())
		if !errors.Is(err, winning.ErrNoYaku) || !errors.Is(err, winning.ErrNotWinning) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("exposes the hand, winning tile, and situation", func(t *testing.T) {
		w := build(t, "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s", "5s")
		if !w.Hand().Equal(mt.Hand("2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s")) || w.WinningTile() != tile.S5 || w.Situation() != southRon() {
			t.Fatal("accessors")
		}
	})
}

func TestIsForm(t *testing.T) {
	t.Run("is true for a complete hand even without yaku", func(t *testing.T) {
		if !winning.IsForm(mt.Hand("2m 3m 4m 5m 6m 7m 2p 3p 4p 5p 6p 7p 1z"), tile.East) {
			t.Fatal("not a form")
		}
	})
	t.Run("is false for an incomplete hand", func(t *testing.T) {
		if winning.IsForm(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s"), tile.S5) {
			t.Fatal("a form")
		}
	})
}

func TestFu(t *testing.T) {
	t.Run("answers the total and the sources", func(t *testing.T) {
		fu := build(t, "1m 2m 3m 4p 5p 6p 7s 8s 9s 2m 3m 2p 2p", "4m").Fu()
		if fu.Total() <= 0 || len(fu.Sources()) == 0 {
			t.Fatal("empty fu")
		}
	})
	t.Run("exposes the reading the fu were counted on", func(t *testing.T) {
		fu := build(t, "2m 2m 2m 3m 3m 3m 4m 4m 4m 5p 6p 7p 8s", "8s").Fu()
		koutsu := 0
		for _, m := range fu.Form().Mentsu() {
			if m.Kind() == winning.Koutsu {
				koutsu++
			}
		}
		if koutsu != 3 {
			t.Fatalf("got %d koutsu", koutsu)
		}
	})
	t.Run("picks the highest reading among those with yaku", func(t *testing.T) {
		if got := build(t, "2m 2m 2m 3m 3m 3m 4m 4m 4m 5p 6p 7p 8s", "8s").Fu().Total(); got != 50 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("Sources returns copies", func(t *testing.T) {
		fu := build(t, "1z 1z 1z 1p 3p 5p 5p 4m 5m 6m 2s 3s 4s", "2p").Fu()
		fu.Sources()[0].Fu = 999
		if fu.Sources()[0].Fu == 999 {
			t.Fatal("shared")
		}
	})
}

type expected struct {
	yakus        []winning.YakuID
	includeYakus []winning.YakuID
	han          *int
	fu           *int
	total        *int
	yakumanCount *int
	rankLabel    *string
	payments     *winning.Payments
}

func n(v int) *int         { return &v }
func str(v string) *string { return &v }

func ids(score winning.Score) []winning.YakuID {
	var out []winning.YakuID
	for _, y := range score.Yakus() {
		out = append(out, y.ID)
	}
	return out
}

func sortedIDs(in []winning.YakuID) []winning.YakuID {
	out := append([]winning.YakuID(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func expectScore(t *testing.T, score winning.Score, want expected) {
	t.Helper()
	got := ids(score)
	if want.yakus != nil && !reflect.DeepEqual(sortedIDs(got), sortedIDs(want.yakus)) {
		t.Errorf("yakus %v, want %v", got, want.yakus)
	}
	for _, id := range want.includeYakus {
		found := false
		for _, g := range got {
			if g == id {
				found = true
			}
		}
		if !found {
			t.Errorf("yakus %v lack %v", got, id)
		}
	}
	if want.han != nil && score.Han() != *want.han {
		t.Errorf("han %d, want %d", score.Han(), *want.han)
	}
	if want.fu != nil && score.Fu() != *want.fu {
		t.Errorf("fu %d, want %d", score.Fu(), *want.fu)
	}
	if want.total != nil && score.Total() != *want.total {
		t.Errorf("total %d, want %d", score.Total(), *want.total)
	}
	if want.yakumanCount != nil && score.YakumanCount() != *want.yakumanCount {
		t.Errorf("yakuman %d, want %d", score.YakumanCount(), *want.yakumanCount)
	}
	if want.rankLabel != nil && score.RankLabel() != *want.rankLabel {
		t.Errorf("rank %q, want %q", score.RankLabel(), *want.rankLabel)
	}
	if want.payments != nil && score.Payments() != *want.payments {
		t.Errorf("payments %+v, want %+v", score.Payments(), *want.payments)
	}
}

// TestScore covers Winning.Score over realistic hands and situations: the
// choice of reading, the propagation of dora and rules, and the payment
// table. Whether each yaku applies is the knowledge tests' job.
func TestScore(t *testing.T) {
	type scoreCase struct {
		name    string
		closed  string
		win     string
		melds   []hand.Meld
		kind    winning.WinKind
		seat    tile.Wind
		edit    func(*winning.Situation)
		dora    int
		roundUp bool
		want    expected
	}
	run := func(t *testing.T, c scoreCase) winning.Score {
		t.Helper()
		s := southRon()
		if c.kind != 0 {
			s.WinKind = c.kind
		}
		if c.seat != 0 {
			s.SeatWind = c.seat
		}
		if c.edit != nil {
			c.edit(&s)
		}
		w, err := winning.New(mt.Hand(c.closed, c.melds...), mt.T(c.win), s, ruleset.Default().WithRoundUpMangan(c.roundUp))
		if err != nil {
			t.Fatal(err)
		}
		return w.Score(c.dora)
	}
	riichiIppatsu := func(s *winning.Situation) { s.Riichi = true; s.Ippatsu = true }
	riichi := func(s *winning.Situation) { s.Riichi = true }

	t.Run("choice of reading", func(t *testing.T) {
		t.Run("a yakuman drops regular yaku", func(t *testing.T) {
			expectScore(t, run(t, scoreCase{closed: "1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z", win: "7z"}),
				expected{yakumanCount: n(1), han: n(0), fu: n(0)})
		})
		t.Run("a chuuren wait picks junsei chuurenpoutou", func(t *testing.T) {
			expectScore(t, run(t, scoreCase{closed: "1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m", win: "5m"}),
				expected{yakumanCount: n(2), includeYakus: []winning.YakuID{winning.YakuJunseiChuurenpoutou}})
		})
	})

	t.Run("propagation", func(t *testing.T) {
		t.Run("dora adds to han", func(t *testing.T) {
			expectScore(t, run(t, scoreCase{closed: "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s", win: "5s", dora: 2}), expected{han: n(3)})
		})
		t.Run("dora alone reaches mangan", func(t *testing.T) {
			expectScore(t, run(t, scoreCase{closed: "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s", win: "5s", dora: 4}),
				expected{han: n(5), rankLabel: str("満貫"), total: n(8000)})
		})
		t.Run("round-up mangan lifts 4 han 30 fu", func(t *testing.T) {
			expectScore(t, run(t, scoreCase{closed: "3m 4m 5m 6m 7m 8m 1p 2p 3p 4p 5p 7s 7s", win: "6p", edit: riichiIppatsu, dora: 1, roundUp: true}),
				expected{han: n(4), fu: n(30), total: n(8000), rankLabel: str("満貫")})
		})
		t.Run("exposes the scored form", func(t *testing.T) {
			form := run(t, scoreCase{closed: "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s", win: "5s"}).Form()
			pair, _ := form.PairTile()
			if form.WaitKind() != winning.Tanki || pair != tile.S5 {
				t.Fatalf("wait %v pair %v", form.WaitKind(), pair)
			}
		})
		t.Run("a yakuman ignores dora", func(t *testing.T) {
			score := run(t, scoreCase{closed: "2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 2p 9s 9s", win: "2p", kind: winning.Tsumo, dora: 3})
			if score.YakumanCount() != 1 || score.DoraCount() != 0 {
				t.Fatalf("yakuman %d dora %d", score.YakumanCount(), score.DoraCount())
			}
		})
	})

	pinfu := "3m 4m 5m 6m 7m 8m 1p 2p 3p 4p 5p 7s 7s"
	tanyao := "2m 3m 4m 5m 6m 7m 3p 4p 5p 6p 7p 8p 5s"
	haku := "1m 2m 3m 4m 5m 6m 7p 8p 9p 4z 4z 5z 5z"
	chiitoi := "1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z"
	kokushi := "1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z"
	cases := []scoreCase{
		{name: "[pinfu] child ron", closed: pinfu, win: "6p",
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu}, han: n(1), fu: n(30), total: n(1000)}},
		{name: "[pinfu] child tsumo, fixed 20 fu", closed: pinfu, win: "6p", kind: winning.Tsumo,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu, winning.YakuMenzenTsumo}, han: n(2), fu: n(20), total: n(1500),
				payments: &winning.Payments{FromDealer: 700, FromNonDealer: 400}}},
		{name: "[pinfu] dealer ron", closed: pinfu, win: "6p", seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu}, han: n(1), fu: n(30), total: n(1500)}},
		{name: "[pinfu] dealer tsumo", closed: pinfu, win: "6p", kind: winning.Tsumo, seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu, winning.YakuMenzenTsumo}, han: n(2), fu: n(20), total: n(2100),
				payments: &winning.Payments{FromNonDealer: 700}}},
		{name: "[pinfu] child ron + dora 1", closed: pinfu, win: "6p", dora: 1, want: expected{han: n(2), fu: n(30), total: n(2000)}},
		{name: "[pinfu] child ron + riichi", closed: pinfu, win: "6p", edit: riichi,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu, winning.YakuRiichi}, han: n(2), fu: n(30), total: n(2000)}},
		{name: "[pinfu] child ron + riichi ippatsu", closed: pinfu, win: "6p", edit: riichiIppatsu,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu, winning.YakuRiichi, winning.YakuIppatsu}, han: n(3), fu: n(30), total: n(3900)}},
		{name: "[pinfu] 4 han 30 fu without round-up is 7700", closed: pinfu, win: "6p", edit: riichiIppatsu, dora: 1,
			want: expected{han: n(4), fu: n(30), total: n(7700)}},
		{name: "[pinfu] child tsumo + riichi ippatsu, 4 han 20 fu", closed: pinfu, win: "6p", kind: winning.Tsumo, edit: riichiIppatsu,
			want: expected{yakus: []winning.YakuID{winning.YakuPinfu, winning.YakuMenzenTsumo, winning.YakuRiichi, winning.YakuIppatsu}, han: n(4), fu: n(20), total: n(5200)}},

		{name: "[tanyao] child ron tanki", closed: tanyao, win: "5s",
			want: expected{yakus: []winning.YakuID{winning.YakuTanyao}, han: n(1), fu: n(40), total: n(1300)}},
		{name: "[tanyao] child tsumo tanki", closed: tanyao, win: "5s", kind: winning.Tsumo,
			want: expected{yakus: []winning.YakuID{winning.YakuTanyao, winning.YakuMenzenTsumo}, han: n(2), fu: n(30), total: n(2000),
				payments: &winning.Payments{FromDealer: 1000, FromNonDealer: 500}}},
		{name: "[tanyao] dealer ron tanki", closed: tanyao, win: "5s", seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuTanyao}, han: n(1), fu: n(40), total: n(2000)}},
		{name: "[tanyao] dealer tsumo tanki", closed: tanyao, win: "5s", kind: winning.Tsumo, seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuTanyao, winning.YakuMenzenTsumo}, han: n(2), fu: n(30), total: n(3000),
				payments: &winning.Payments{FromNonDealer: 1000}}},
		{name: "[tanyao] child ron + dora 2, 3 han 40 fu", closed: tanyao, win: "5s", dora: 2, want: expected{han: n(3), fu: n(40), total: n(5200)}},

		{name: "[haku] child ron shanpon", closed: haku, win: "5z",
			want: expected{yakus: []winning.YakuID{winning.YakuHaku}, han: n(1), fu: n(40), total: n(1300)}},
		{name: "[haku] child tsumo shanpon, the set stays concealed", closed: haku, win: "5z", kind: winning.Tsumo,
			want: expected{yakus: []winning.YakuID{winning.YakuHaku, winning.YakuMenzenTsumo}, han: n(2), fu: n(30), total: n(2000),
				payments: &winning.Payments{FromDealer: 1000, FromNonDealer: 500}}},
		{name: "[haku] dealer ron shanpon", closed: haku, win: "5z", seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuHaku}, han: n(1), fu: n(40), total: n(2000)}},
		{name: "[haku] child ron + riichi, 2 han 40 fu", closed: haku, win: "5z", edit: riichi,
			want: expected{yakus: []winning.YakuID{winning.YakuHaku, winning.YakuRiichi}, han: n(2), fu: n(40), total: n(2600)}},

		{name: "[toitoi] open child ron with sanankou", closed: "2m 2m 2m 3s 3s 3s 7s 7s 7s 1z", win: "1z", melds: []hand.Meld{mt.Pon("5p 5p 5p")},
			want: expected{yakus: []winning.YakuID{winning.YakuToitoi, winning.YakuSanankou}, han: n(4)}},
		{name: "[toitoi] open child tsumo", closed: "2m 2m 2m 3s 3s 3s 7s 7s 7s 1z", win: "1z", melds: []hand.Meld{mt.Pon("5p 5p 5p")}, kind: winning.Tsumo,
			want: expected{includeYakus: []winning.YakuID{winning.YakuToitoi, winning.YakuSanankou}}},

		{name: "[honitsu] closed child ron", closed: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1z 1z 5z 5z", win: "5z",
			want: expected{includeYakus: []winning.YakuID{winning.YakuHonitsu, winning.YakuHaku}}},
		{name: "[honitsu] closed child ron + riichi", closed: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1z 1z 5z 5z", win: "5z", edit: riichi,
			want: expected{includeYakus: []winning.YakuID{winning.YakuHonitsu, winning.YakuHaku, winning.YakuRiichi}}},
		{name: "[honitsu] open child ron", closed: "1m 2m 3m 1m 2m 3m 4m 5m 1z 1z", win: "6m", melds: []hand.Meld{mt.Pon("5z 5z 5z")},
			want: expected{includeYakus: []winning.YakuID{winning.YakuHonitsu, winning.YakuHaku}}},
		{name: "[chinitsu] closed child ron", closed: "2m 3m 4m 5m 6m 7m 3m 4m 5m 6m 7m 8m 5m", win: "8m",
			want: expected{includeYakus: []winning.YakuID{winning.YakuChinitsu}}},

		{name: "[chiitoitsu] child ron, fixed 25 fu", closed: chiitoi, win: "7z",
			want: expected{includeYakus: []winning.YakuID{winning.YakuChiitoitsu}, han: n(2), fu: n(25), total: n(1600)}},
		{name: "[chiitoitsu] child ron + riichi, 3 han 25 fu", closed: chiitoi, win: "7z", edit: riichi,
			want: expected{includeYakus: []winning.YakuID{winning.YakuChiitoitsu, winning.YakuRiichi}, han: n(3), fu: n(25), total: n(3200)}},
		{name: "[chiitoitsu] child tsumo", closed: chiitoi, win: "7z", kind: winning.Tsumo,
			want: expected{includeYakus: []winning.YakuID{winning.YakuChiitoitsu, winning.YakuMenzenTsumo}, fu: n(25)}},

		{name: "[kokushi] child ron tanki", closed: kokushi, win: "7z",
			want: expected{yakus: []winning.YakuID{winning.YakuKokushimusou}, yakumanCount: n(1), total: n(32000), rankLabel: str("役満")}},
		{name: "[kokushi] dealer ron", closed: kokushi, win: "7z", seat: tile.EastWind,
			want: expected{yakus: []winning.YakuID{winning.YakuKokushimusou}, yakumanCount: n(1), total: n(48000), rankLabel: str("役満")}},
		{name: "[kokushi 13-sided] child ron, double yakuman", closed: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", win: "1z",
			want: expected{yakus: []winning.YakuID{winning.YakuKokushimusouJuusanmen}, yakumanCount: n(2), total: n(64000), rankLabel: str("ダブル役満")}},

		{name: "[suuankou] child tsumo shanpon", closed: "2m 2m 2m 5p 5p 5p 7s 7s 7s 9s 9s 4z 4z", win: "4z", kind: winning.Tsumo,
			want: expected{yakus: []winning.YakuID{winning.YakuSuuankou}, yakumanCount: n(1), total: n(32000),
				payments: &winning.Payments{FromDealer: 16000, FromNonDealer: 8000}}},
		{name: "[suuankou tanki] child ron, double yakuman", closed: "2m 2m 2m 5p 5p 5p 7s 7s 7s 9s 9s 9s 4z", win: "4z",
			want: expected{yakus: []winning.YakuID{winning.YakuSuuankouTanki}, yakumanCount: n(2), total: n(64000)}},

		{name: "[haitei + tanyao + menzen tsumo] 3 han", closed: tanyao, win: "5s", kind: winning.Tsumo, edit: func(s *winning.Situation) { s.Haitei = true },
			want: expected{yakus: []winning.YakuID{winning.YakuHaitei, winning.YakuTanyao, winning.YakuMenzenTsumo}, han: n(3)}},
		{name: "[rinshan + menzen tsumo with an ankan]", closed: "1m 2m 3m 4p 5p 6p 7s 8s 9s 4z", win: "4z", melds: []hand.Meld{mt.Ankan("5m 5m 5m 5m")},
			kind: winning.Tsumo, edit: func(s *winning.Situation) { s.Rinshan = true },
			want: expected{includeYakus: []winning.YakuID{winning.YakuRinshan, winning.YakuMenzenTsumo}}},
		{name: "[chankan + haku shanpon]", closed: haku, win: "5z", edit: func(s *winning.Situation) { s.Chankan = true },
			want: expected{yakus: []winning.YakuID{winning.YakuChankan, winning.YakuHaku}, han: n(2)}},

		{name: "[round-up mangan] 3 han 60 fu", closed: "6z 6z 6z 1m 2m 3m 1p 2p 3p 4z", win: "4z", melds: []hand.Meld{mt.Ankan("5m 5m 5m 5m")},
			edit: riichiIppatsu, roundUp: true,
			want: expected{yakus: []winning.YakuID{winning.YakuRiichi, winning.YakuIppatsu, winning.YakuHatsu}, han: n(3), fu: n(60), total: n(8000), rankLabel: str("満貫")}},
		{name: "[kazoe yakuman] riichi ippatsu pinfu tanyao chinitsu + dora 4 = 14 han", closed: "2m 3m 3m 4m 4m 5m 5m 6m 6m 7m 7m 8m 8m", win: "8m",
			edit: riichiIppatsu, dora: 4,
			want: expected{yakus: []winning.YakuID{winning.YakuRiichi, winning.YakuIppatsu, winning.YakuTanyao, winning.YakuPinfu, winning.YakuChinitsu},
				han: n(14), rankLabel: str("数え役満"), total: n(32000)}},

		{name: "[table] child ron + dora 5, 6 han haneman", closed: pinfu, win: "6p", dora: 5, want: expected{han: n(6), rankLabel: str("跳満"), total: n(12000)}},
		{name: "[table] child ron + dora 7, 8 han baiman", closed: pinfu, win: "6p", dora: 7, want: expected{han: n(8), rankLabel: str("倍満"), total: n(16000)}},
		{name: "[table] child ron + dora 10, 11 han sanbaiman", closed: pinfu, win: "6p", dora: 10, want: expected{han: n(11), rankLabel: str("三倍満"), total: n(24000)}},
		{name: "[table] dealer ron + dora 4, mangan 12000", closed: pinfu, win: "6p", seat: tile.EastWind, dora: 4, want: expected{han: n(5), rankLabel: str("満貫"), total: n(12000)}},
		{name: "[table] child tsumo + dora 1, 3 han 30 fu", closed: tanyao, win: "5s", kind: winning.Tsumo, dora: 1,
			want: expected{han: n(3), fu: n(30), total: n(4000), payments: &winning.Payments{FromDealer: 2000, FromNonDealer: 1000}}},
		{name: "[table] dealer tsumo + dora 1, 2000 all", closed: tanyao, win: "5s", kind: winning.Tsumo, seat: tile.EastWind, dora: 1,
			want: expected{han: n(3), fu: n(30), total: n(6000), payments: &winning.Payments{FromNonDealer: 2000}}},
		{name: "[table] child tsumo + dora 3, mangan 4000/2000", closed: tanyao, win: "5s", kind: winning.Tsumo, dora: 3,
			want: expected{han: n(5), rankLabel: str("満貫"), total: n(8000), payments: &winning.Payments{FromDealer: 4000, FromNonDealer: 2000}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expectScore(t, run(t, c), c.want)
		})
	}
}

func TestPayments(t *testing.T) {
	cases := []struct {
		p    winning.Payments
		want int
	}{
		{winning.Payments{FromLoser: 1300}, 1300},
		{winning.Payments{FromNonDealer: 700}, 2100},
		{winning.Payments{FromDealer: 700, FromNonDealer: 400}, 1500},
	}
	for _, c := range cases {
		if c.p.Total() != c.want {
			t.Errorf("%+v: %d, want %d", c.p, c.p.Total(), c.want)
		}
	}
}

func TestEnums(t *testing.T) {
	if winning.Ron.String() != "ron" || winning.Tsumo.String() != "tsumo" || winning.WinKind(9).String() != "WinKind(9)" {
		t.Error("WinKind")
	}
	if k, err := winning.ParseWinKind("tsumo"); err != nil || k != winning.Tsumo {
		t.Error("ParseWinKind")
	}
	if _, err := winning.ParseWinKind("draw"); !errors.Is(err, winning.ErrInvalidSituation) {
		t.Error("ParseWinKind error")
	}
	if winning.Standard.String() != "standard" || winning.Chiitoitsu.String() != "chiitoitsu" || winning.Kokushi.String() != "kokushi" || winning.FormKind(9).String() != "FormKind(9)" {
		t.Error("FormKind")
	}
	if winning.Ryanmen.String() != "ryanmen" || winning.Shanpon.String() != "shanpon" || winning.WaitKind(9).String() != "WaitKind(9)" {
		t.Error("WaitKind")
	}
	if winning.Shuntsu.String() != "shuntsu" || winning.Kantsu.String() != "kantsu" || winning.MentsuKind(9).String() != "MentsuKind(9)" {
		t.Error("MentsuKind")
	}
	if !(winning.Yaku{Yakuman: 1}).IsYakuman() || (winning.Yaku{Han: 1}).IsYakuman() {
		t.Error("Yaku.IsYakuman")
	}
}
