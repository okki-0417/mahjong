package knowledge_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// 和了の成立。テンパイの13枚に和了牌1枚が加わり、14枚が和了形になること。
func TestWinning(t *testing.T) {
	win := func(closed, winTile string, melds ...hand.Meld) (*winning.Winning, error) {
		return winOf(closed, winTile, melds, sit(winning.Tsumo, tile.EastWind, tile.EastWind), ruleset.Default())
	}

	t.Run("和了形", func(t *testing.T) {
		t.Run("標準形（4面子1雀頭）で和了となること", func(t *testing.T) {
			if _, err := win("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s"); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("七対子（異なる7対子）で和了となること", func(t *testing.T) {
			if _, err := win("1m 1m 3m 3m 5m 5m 7m 7m 9m 9m 1p 1p 3p", "3p"); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("国士無双（13種の么九牌 + いずれか1種の対子）で和了となること", func(t *testing.T) {
			if _, err := win("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", "1m"); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("面子がそろっていない手は和了とならないこと", func(t *testing.T) {
			if _, err := win("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s", "5s"); !errors.Is(err, winning.ErrNoForm) {
				t.Fatalf("err = %v", err)
			}
		})
	})

	t.Run("和了牌", func(t *testing.T) {
		t.Run("和了牌は手牌の外から来て、テンパイの13枚に加わり14枚になること", func(t *testing.T) {
			w, err := win("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s")
			if err != nil {
				t.Fatal(err)
			}
			form := w.Fu().Form()
			count := 0
			for _, m := range form.Mentsu() {
				count += len(m.Tiles())
			}
			if pair, ok := form.PairTile(); ok && pair != 0 {
				count += 2
			}
			if count != 14 {
				t.Fatalf("got %d tiles", count)
			}
		})
		t.Run("同種は4枚までで、和了牌を加えて5枚になる形は和了として成立しないこと", func(t *testing.T) {
			_, err := win("1m 1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m", "1m")
			if !errors.Is(err, winning.ErrTileExhausted) {
				t.Fatalf("err = %v", err)
			}
		})
	})

	t.Run("門前の判定", func(t *testing.T) {
		t.Run("暗槓は門前を保つこと", func(t *testing.T) {
			if !mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", mt.Ankan("1z 1z 1z 1z")).IsMenzen() {
				t.Fatal("not menzen")
			}
		})
		t.Run("明槓・ポン・チーは門前を崩すこと", func(t *testing.T) {
			for _, m := range []hand.Meld{mt.Minkan("1z 1z 1z 1z"), mt.Pon("1z 1z 1z"), mt.Chi("1p 2p 3p")} {
				if mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", m).IsMenzen() {
					t.Errorf("%v keeps menzen", m)
				}
			}
		})
	})

	t.Run("役の要求", func(t *testing.T) {
		t.Run("和了形でも役が1つも無ければ和了として成立しないこと", func(t *testing.T) {
			// 鳴いた手なので門前清自摸和が付かず、北は場風でも自風でもないので役牌にならない。
			_, err := winOf("2m 3m 4m 6m 7m 9p 9p", "8m", []hand.Meld{mt.Pon("4z 4z 4z"), mt.Chi("1p 2p 3p")},
				sit(winning.Tsumo, tile.EastWind, tile.SouthWind), ruleset.Default())
			if !errors.Is(err, winning.ErrNoYaku) {
				t.Fatalf("err = %v", err)
			}
		})
	})
}

// 和了状況どうしの整合性。ある状況の組み合わせは麻雀のルール上ありえない。
func TestSituation(t *testing.T) {
	// 手は常に和了形（一気通貫の単騎和了）に固定し、状況フラグだけを差し替える。
	situation := func(edit func(*winning.Situation)) error {
		s := sit(winning.Tsumo, tile.EastWind, tile.EastWind)
		edit(&s)
		_, err := winOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s", nil, s, ruleset.Default())
		return err
	}
	accepted := func(t *testing.T, edit func(*winning.Situation)) {
		t.Helper()
		if err := situation(edit); err != nil {
			t.Errorf("rejected: %v", err)
		}
	}
	rejected := func(t *testing.T, edit func(*winning.Situation)) {
		t.Helper()
		if err := situation(edit); !errors.Is(err, winning.ErrInvalidSituation) {
			t.Errorf("accepted (err = %v)", err)
		}
	}

	t.Run("一発は立直（またはダブル立直）しているときのみ成立すること", func(t *testing.T) {
		rejected(t, func(s *winning.Situation) { s.Ippatsu = true })
		accepted(t, func(s *winning.Situation) { s.Riichi = true; s.Ippatsu = true })
		accepted(t, func(s *winning.Situation) { s.DoubleRiichi = true; s.Ippatsu = true })
	})
	t.Run("立直とダブル立直は同時に成立しないこと", func(t *testing.T) {
		rejected(t, func(s *winning.Situation) { s.Riichi = true; s.DoubleRiichi = true })
	})
	t.Run("海底摸月（ハイテイ）はツモのときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.Haitei = true })
		rejected(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Haitei = true })
	})
	t.Run("河底撈魚（ホウテイ）はロンのときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Houtei = true })
		rejected(t, func(s *winning.Situation) { s.Houtei = true })
	})
	t.Run("嶺上開花はツモのときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.Rinshan = true })
		rejected(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Rinshan = true })
	})
	t.Run("槍槓はロンのときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Chankan = true })
		rejected(t, func(s *winning.Situation) { s.Chankan = true })
	})
	t.Run("天和と地和は同時に成立しないこと", func(t *testing.T) {
		rejected(t, func(s *winning.Situation) { s.Tenhou = true; s.Chiihou = true })
	})
	t.Run("天和・地和はツモのときのみ成立すること", func(t *testing.T) {
		rejected(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Tenhou = true })
		rejected(t, func(s *winning.Situation) { s.WinKind = winning.Ron; s.Chiihou = true; s.SeatWind = tile.SouthWind })
	})
	t.Run("天和・地和は立直・ダブル立直と同時に成立しないこと", func(t *testing.T) {
		rejected(t, func(s *winning.Situation) { s.Tenhou = true; s.Riichi = true })
		rejected(t, func(s *winning.Situation) { s.Chiihou = true; s.SeatWind = tile.SouthWind; s.DoubleRiichi = true })
	})
	t.Run("天和は親（東家）のときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.Tenhou = true })
		rejected(t, func(s *winning.Situation) { s.SeatWind = tile.SouthWind; s.Tenhou = true })
	})
	t.Run("地和は子（東家以外）のときのみ成立すること", func(t *testing.T) {
		accepted(t, func(s *winning.Situation) { s.SeatWind = tile.SouthWind; s.Chiihou = true })
		rejected(t, func(s *winning.Situation) { s.Chiihou = true })
	})
}
