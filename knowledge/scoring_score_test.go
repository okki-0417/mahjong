package knowledge_test

import (
	"testing"

	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// 翻・符から最終的な点数と支払いを決めるルール。親子・ロン/ツモで分配が変わる。
func TestScore(t *testing.T) {
	// 支払いは100点単位へ切り上げられる。
	ceil100 := func(points int) int { return (points + 99) / 100 * 100 }

	// 一通の手を土台にして、翻・符を変えた採点を組み立てる。
	scoreOf := func(t *testing.T, closed, win string, kind winning.WinKind, seat tile.Wind, rs ruleset.RuleSet, doraCount int) winning.Score {
		t.Helper()
		w, err := winOf(closed, win, nil, sit(kind, tile.EastWind, seat), rs)
		if err != nil {
			t.Fatal(err)
		}
		return w.Score(doraCount)
	}
	// 子・門前ロン・一通のみ（2翻40符）。
	ittsu := func(t *testing.T, kind winning.WinKind, seat tile.Wind, doraCount int) winning.Score {
		t.Helper()
		return scoreOf(t, "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s", kind, seat, ruleset.Default(), doraCount)
	}
	child := func(t *testing.T, doraCount int) winning.Score {
		t.Helper()
		return ittsu(t, winning.Ron, tile.SouthWind, doraCount)
	}
	dealer := func(t *testing.T, kind winning.WinKind, doraCount int) winning.Score {
		t.Helper()
		return ittsu(t, kind, tile.EastWind, doraCount)
	}

	t.Run("基本点", func(t *testing.T) {
		t.Run("基本点 = 符 × 2の(翻+2)乗 となること", func(t *testing.T) {
			s := child(t, 0)
			if s.BasePoints() != s.Fu()*(1<<(2+s.Han())) {
				t.Fatalf("base %d fu %d han %d", s.BasePoints(), s.Fu(), s.Han())
			}
		})
		t.Run("基本点は 2000（満貫）で頭打ちとなること", func(t *testing.T) {
			if child(t, 10).BasePoints() > 8000 || child(t, 3).BasePoints() != 2000 {
				t.Fatalf("base %d / %d", child(t, 10).BasePoints(), child(t, 3).BasePoints())
			}
		})
	})

	t.Run("翻数によるランク", func(t *testing.T) {
		t.Run("満貫未満はランクが付かないこと", func(t *testing.T) {
			if got := child(t, 0).RankLabel(); got != "" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("5翻、または基本点2000到達で 満貫 となること", func(t *testing.T) {
			if got := child(t, 3).RankLabel(); got != "満貫" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("6〜7翻は 跳満、8〜10翻は 倍満、11〜12翻は 三倍満、13翻以上は 数え役満 となること", func(t *testing.T) {
			for dora, want := range map[int]string{4: "跳満", 6: "倍満", 9: "三倍満", 11: "数え役満"} {
				if got := child(t, dora).RankLabel(); got != want {
					t.Errorf("dora %d: %q, want %q", dora, got, want)
				}
			}
		})
	})

	t.Run("子の点数", func(t *testing.T) {
		t.Run("子ロンは 基本点 × 4 となること", func(t *testing.T) {
			s := child(t, 0)
			if s.Total() != ceil100(s.BasePoints()*4) || s.Payments().FromLoser != s.Total() {
				t.Fatalf("total %d base %d payments %+v", s.Total(), s.BasePoints(), s.Payments())
			}
		})
		t.Run("子ツモは 親から 基本点×2、子から 基本点×1 となること", func(t *testing.T) {
			s := ittsu(t, winning.Tsumo, tile.SouthWind, 0)
			p := s.Payments()
			if p.FromDealer != ceil100(s.BasePoints()*2) || p.FromNonDealer != ceil100(s.BasePoints()) {
				t.Fatalf("payments %+v base %d", p, s.BasePoints())
			}
		})
		t.Run("子の満貫は 8000点 となること", func(t *testing.T) {
			if got := child(t, 3).Total(); got != 8000 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("親の点数", func(t *testing.T) {
		t.Run("親ロンは 基本点 × 6 となること", func(t *testing.T) {
			s := dealer(t, winning.Ron, 0)
			if s.Total() != ceil100(s.BasePoints()*2)*3 {
				t.Fatalf("total %d base %d", s.Total(), s.BasePoints())
			}
		})
		t.Run("親ツモは 各家から 基本点×2 となること", func(t *testing.T) {
			s := dealer(t, winning.Tsumo, 0)
			if s.Payments().FromNonDealer != ceil100(s.BasePoints()*2) {
				t.Fatalf("payments %+v base %d", s.Payments(), s.BasePoints())
			}
		})
		t.Run("親の満貫は 12000点 となること", func(t *testing.T) {
			if got := dealer(t, winning.Ron, 3).Total(); got != 12000 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("端数処理", func(t *testing.T) {
		t.Run("各支払いは 100点単位へ切り上げること", func(t *testing.T) {
			for _, s := range []winning.Score{child(t, 0), ittsu(t, winning.Tsumo, tile.SouthWind, 0)} {
				p := s.Payments()
				if p.FromLoser%100 != 0 || p.FromDealer%100 != 0 || p.FromNonDealer%100 != 0 {
					t.Errorf("payments %+v", p)
				}
			}
		})
	})

	t.Run("役満の点数", func(t *testing.T) {
		kokushi := func(t *testing.T, seat tile.Wind, doraCount int) winning.Score {
			t.Helper()
			return scoreOf(t, "1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z", "7z", winning.Ron, seat, ruleset.Default(), doraCount)
		}

		t.Run("子の役満は 32000点、親の役満は 48000点 となること", func(t *testing.T) {
			if kokushi(t, tile.SouthWind, 0).Total() != 32000 || kokushi(t, tile.EastWind, 0).Total() != 48000 {
				t.Fatalf("child %d dealer %d", kokushi(t, tile.SouthWind, 0).Total(), kokushi(t, tile.EastWind, 0).Total())
			}
		})
		t.Run("役満成立時はドラを数えないこと", func(t *testing.T) {
			if got := kokushi(t, tile.SouthWind, 5).Total(); got != 32000 {
				t.Fatalf("got %d", got)
			}
		})
	})

	// 採用するかどうかは卓ごとの取り決めで、麻雀としてはどちらもありうる。
	t.Run("切り上げ満貫（採用ルール）", func(t *testing.T) {
		t.Run("採用しなければ満貫に届かない手をそのまま計算すること", func(t *testing.T) {
			if got := child(t, 1).RankLabel(); got != "" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("採用すれば、満貫にわずかに届かない基本点を満貫として扱うこと", func(t *testing.T) {
			near := scoreOf(t, "2m 3m 4m 6m 7m 8m 1p 2p 3p 5p 6p 9s 9s", "7p", winning.Ron, tile.SouthWind,
				ruleset.Default().WithRoundUpMangan(true), 0)
			if near.BasePoints() > 2000 {
				t.Fatalf("base %d", near.BasePoints())
			}
		})
	})
}
