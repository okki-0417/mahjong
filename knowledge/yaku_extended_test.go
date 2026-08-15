package knowledge_test

import (
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// 牌の色・種類・面子種別の偏りから生まれる役。多くは副露で食い下がる。
func TestExtendedYaku(t *testing.T) {
	t.Run("断么九（タンヤオ）", func(t *testing.T) {
		// 鳴いた断么九。喰い断の設定でだけ成否が変わる形。
		meldedTanyao := func(rs ruleset.RuleSet) []string {
			return yakuNames(winOf("2m 3m 4m 5m 6m 7m 5p 6p 3s 3s", "7p", []hand.Meld{mt.Chi("2p 3p 4p")},
				sit(winning.Ron, tile.EastWind, tile.SouthWind), rs))
		}

		t.Run("2〜8の中張牌のみで構成されるとき成立し 1翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("断么九", "2m 3m 4m 5m 6m 7m 2p 3p 4p 5p 6p 3s 3s", "7p"), 1)
		})
		t.Run("么九牌（1・9・字牌）を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 2m 3m 5m 6m 7m 2p 3p 4p 5p 6p 3s 3s", "7p"), "断么九")
		})
		t.Run("赤5は通常の5として中張牌に数えること", func(t *testing.T) {
			expectHas(t, ronNames("2m 3m 4m 0m 6m 7m 2p 3p 4p 5p 6p 3s 3s", "7p"), "断么九")
		})
		t.Run("副露していても成立すること（喰い断あり）", func(t *testing.T) {
			expectNames(t, meldedTanyao(ruleset.Default()), "断么九")
		})
		t.Run("喰い断の有無はルール設定で切り替えられること", func(t *testing.T) {
			expectNames(t, meldedTanyao(ruleset.Default().WithKuitan(true)), "断么九")
			if got := meldedTanyao(ruleset.Default().WithKuitan(false)); len(got) != 0 {
				t.Errorf("got %v", got)
			}
		})
		t.Run("喰い断なしでも、門前なら断么九は成立すること", func(t *testing.T) {
			names := yakuNames(winOf("2m 3m 4m 5m 6m 7m 2p 3p 4p 5p 6p 3s 3s", "7p", nil,
				sit(winning.Ron, tile.EastWind, tile.SouthWind), ruleset.Default().WithKuitan(false)))
			expectHas(t, names, "断么九")
		})
	})

	t.Run("対々和（トイトイ）", func(t *testing.T) {
		t.Run("4面子すべてが刻子（槓子を含む）+ 雀頭 のとき成立し 2翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("対々和", "5m 5m 5m 8m 8m 8m 2p 2p 2p 3s", "3s", mt.Pon("2m 2m 2m")), 2)
		})
		t.Run("順子を1つでも含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 3p 4p 3s", "3s"), "対々和")
		})
		t.Run("副露していても 2翻のまま（食い下がりなし）となること", func(t *testing.T) {
			expectHan(t, ronHan("対々和", "8m 8m 8m 2p 2p 2p 3s", "3s", mt.Pon("2m 2m 2m"), mt.Pon("5m 5m 5m")), 2)
		})
	})

	t.Run("七対子（チートイツ）", func(t *testing.T) {
		t.Run("異なる7種の対子で構成され門前のとき成立し 2翻・25符固定 となること", func(t *testing.T) {
			w, err := structuralWin("1m 1m 3m 3m 5m 5m 7m 7m 9m 9m 1p 1p 3p", "3p", winning.Ron)
			if err != nil {
				t.Fatal(err)
			}
			score := w.Score(0)
			expectNames(t, yakuNames(w, nil), "七対子")
			if score.Han() != 2 || score.Fu() != 25 {
				t.Fatalf("han %d fu %d", score.Han(), score.Fu())
			}
		})
		t.Run("同種4枚を2対子として数えないこと（不成立）", func(t *testing.T) {
			if got := ronNames("1m 1m 1m 1m 3m 3m 5m 5m 7m 7m 9m 9m 1p", "1p"); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			if got := ronNames("3m 3m 5m 5m 7m 7m 9m 9m 1p 1p", "1p", mt.Pon("1m 1m 1m")); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("二盃口として解釈できる形では二盃口を優先すること", func(t *testing.T) {
			names := ronNames("2m 2m 3m 3m 4m 4m 6m 6m 7m 7m 8m 8m 9s", "9s")
			expectHas(t, names, "二盃口")
			expectLacks(t, names, "七対子")
		})
	})

	t.Run("混一色（ホンイツ）", func(t *testing.T) {
		t.Run("1種類の数牌 + 字牌のみで構成されるとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 2m 3m 4m 5m 6m 7m 8m 9m 5z 5z 5z 1z", "1z"), "混一色")
		})
		t.Run("門前で 3翻、副露で 2翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("混一色", "1m 2m 3m 4m 5m 6m 7m 8m 9m 5z 5z 5z 1z", "1z"), 3)
			expectHan(t, ronHan("混一色", "4m 5m 6m 7m 8m 9m 5z 5z 5z 1z", "1z", mt.Chi("1m 2m 3m")), 2)
		})
		t.Run("2種以上の数牌を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 2m 3m 4m 5m 6m 7p 8p 9p 5z 5z 5z 1z", "1z"), "混一色")
		})
	})

	t.Run("清一色（チンイツ）", func(t *testing.T) {
		t.Run("1種類の数牌のみ（字牌なし）で構成されるとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 1m 1m 2m 3m 4m 2m 3m 4m 5m 6m 7m 9m", "9m"), "清一色")
		})
		t.Run("門前で 6翻、副露で 5翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("清一色", "1m 1m 1m 2m 3m 4m 2m 3m 4m 5m 6m 7m 9m", "9m"), 6)
			expectHan(t, ronHan("清一色", "1m 1m 1m 2m 3m 4m 5m 6m 7m 9m", "9m", mt.Chi("2m 3m 4m")), 5)
		})
		t.Run("字牌を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 2m 3m 4m 5m 6m 7m 8m 9m 5z 5z 5z 1z", "1z"), "清一色")
		})
	})

	t.Run("混全帯么九（チャンタ）", func(t *testing.T) {
		t.Run("すべての面子と雀頭が么九牌（1・9・字牌）を含み、字牌を含むとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 2m 3m 7m 8m 9m 1p 2p 3p 1z 1z 1z 9s", "9s"), "混全帯么九")
		})
		t.Run("門前で 2翻、副露で 1翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("混全帯么九", "1m 2m 3m 7m 8m 9m 1p 2p 3p 1z 1z 1z 9s", "9s"), 2)
			expectHan(t, ronHan("混全帯么九", "7m 8m 9m 1p 2p 3p 1z 1z 1z 9s", "9s", mt.Chi("1m 2m 3m")), 1)
		})
		t.Run("中張牌だけの面子を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 2m 3m 7m 8m 9m 4p 5p 6p 1z 1z 1z 9s", "9s"), "混全帯么九")
		})
		t.Run("刻子・順子はあってよいが、么九牌を含まない面子があると不成立となること", func(t *testing.T) {
			expectHas(t, ronNames("1m 1m 1m 7m 8m 9m 1p 2p 3p 1z 1z 1z 9s", "9s"), "混全帯么九")
			expectLacks(t, ronNames("1m 1m 1m 7m 8m 9m 5p 5p 5p 1z 1z 1z 9s", "9s"), "混全帯么九")
		})
	})

	t.Run("純全帯么九（純チャン）", func(t *testing.T) {
		t.Run("すべての面子と雀頭が老頭牌（1・9）を含み、字牌を含まないとき成立すること", func(t *testing.T) {
			names := ronNames("1m 2m 3m 7m 8m 9m 1p 2p 3p 9p 9p 9p 9s", "9s")
			expectHas(t, names, "純全帯么九")
			expectLacks(t, names, "混全帯么九")
		})
		t.Run("門前で 3翻、副露で 2翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("純全帯么九", "1m 2m 3m 7m 8m 9m 1p 2p 3p 9p 9p 9p 9s", "9s"), 3)
			expectHan(t, ronHan("純全帯么九", "7m 8m 9m 1p 2p 3p 9p 9p 9p 9s", "9s", mt.Chi("1m 2m 3m")), 2)
		})
	})

	t.Run("混老頭（ホンロウトウ）", func(t *testing.T) {
		honroutou := func() (*winning.Winning, error) {
			return structuralWin("9m 9m 9m 1z 1z 1z 5z 5z 5z 9s", "9s", winning.Ron, mt.Pon("1m 1m 1m"))
		}

		t.Run("老頭牌（1・9）と字牌のみで構成されるとき成立し 2翻 となること", func(t *testing.T) {
			w, err := honroutou()
			expectHan(t, hanOf("混老頭")(w, err), 2)
		})
		t.Run("対々和 または 七対子 と必ず複合すること", func(t *testing.T) {
			w, err := honroutou()
			expectHas(t, yakuNames(w, err), "対々和")
			names := ronNames("1m 1m 9m 9m 1p 1p 9p 9p 1s 1s 1z 1z 5z", "5z")
			expectHas(t, names, "混老頭")
			expectHas(t, names, "七対子")
		})
		t.Run("順子を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 2m 3m 7m 8m 9m 1p 2p 3p 1z 1z 1z 9s", "9s"), "混老頭")
		})
	})

	t.Run("三色同刻（サンショクドウコウ）", func(t *testing.T) {
		t.Run("萬子・筒子・索子で同じ番号の刻子3組を含むとき成立し 2翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("三色同刻", "2m 2m 2m 2p 2p 2p 2s 2s 2s 5m 6m 7m 3s", "3s"), 2)
		})
	})

	t.Run("三槓子（サンカンツ）", func(t *testing.T) {
		t.Run("槓子を3つ含むとき成立し 2翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("三槓子", "2m 3m 4m 3s", "3s", mt.Minkan("1z 1z 1z 1z"), mt.Minkan("2z 2z 2z 2z"), mt.Ankan("5z 5z 5z 5z")), 2)
		})
	})

	t.Run("小三元（ショウサンゲン）", func(t *testing.T) {
		shousangen := func() (*winning.Winning, error) {
			return structuralWin("5z 5z 5z 6z 6z 6z 1m 2m 3m 4m 5m 6m 7z", "7z", winning.Ron)
		}

		t.Run("三元牌のうち2種を刻子、残り1種を雀頭にしたとき成立し 2翻 となること", func(t *testing.T) {
			w, err := shousangen()
			expectHan(t, hanOf("小三元")(w, err), 2)
		})
		t.Run("役牌2つ分（成立している三元牌の刻子）と複合すること", func(t *testing.T) {
			w, err := shousangen()
			names := yakuNames(w, err)
			expectHas(t, names, "役牌（白）")
			expectHas(t, names, "役牌（發）")
		})
	})
}
