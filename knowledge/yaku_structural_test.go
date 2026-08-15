package knowledge_test

import (
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// structuralWin is a ron by the south seat with the east round wind, the
// starting point of every structural yaku example.
func structuralWin(closed, win string, kind winning.WinKind, melds ...hand.Meld) (*winning.Winning, error) {
	return winOf(closed, win, melds, sit(kind, tile.EastWind, tile.SouthWind), ruleset.Default())
}

func ronNames(closed, win string, melds ...hand.Meld) []string {
	return yakuNames(structuralWin(closed, win, winning.Ron, melds...))
}

func ronHan(name, closed, win string, melds ...hand.Meld) int {
	return hanOf(name)(structuralWin(closed, win, winning.Ron, melds...))
}

func expectNames(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func expectHas(t *testing.T, names []string, name string) {
	t.Helper()
	if !has(names, name) {
		t.Errorf("%v lacks %s", names, name)
	}
}

func expectLacks(t *testing.T, names []string, name string) {
	t.Helper()
	if has(names, name) {
		t.Errorf("%v has %s", names, name)
	}
}

func expectHan(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("han %d, want %d", got, want)
	}
}

// 手牌の面子構成そのものから生まれる役（状況・鳴きの有無で翻が変わるものを含む）。
// ここには「どういう和了形なら役がつき、何翻になるか」という麻雀のルールだけを描く。
func TestStructuralYaku(t *testing.T) {
	t.Run("平和（ピンフ）", func(t *testing.T) {
		// 平和が崩れても和了れるように、一気通貫を並べた形で不成立を見る。
		besideIttsu := func(tail, win string) []string {
			return ronNames("1m 2m 3m 4m 5m 6m 7m 8m 9m "+tail, win)
		}

		t.Run("4順子 + 非役牌の雀頭 + 両面待ち + 門前 のとき成立し 1翻 となること", func(t *testing.T) {
			expectNames(t, ronNames("2m 3m 4m 5m 6m 7m 1p 2p 3p 4p 5p 9s 9s", "6p"), "平和")
			expectHan(t, ronHan("平和", "2m 3m 4m 5m 6m 7m 1p 2p 3p 4p 5p 9s 9s", "6p"), 1)
		})
		t.Run("門前ロンは 30符、門前ツモは 20符固定（ツモの2符が付かない）となること", func(t *testing.T) {
			ron, _ := structuralWin("2m 3m 4m 5m 6m 7m 1p 2p 3p 4p 5p 9s 9s", "6p", winning.Ron)
			tsumo, _ := structuralWin("2m 3m 4m 5m 6m 7m 1p 2p 3p 4p 5p 9s 9s", "6p", winning.Tsumo)
			if ron.Score(0).Fu() != 30 || tsumo.Score(0).Fu() != 20 {
				t.Fatalf("ron %d tsumo %d", ron.Score(0).Fu(), tsumo.Score(0).Fu())
			}
		})
		t.Run("刻子を含むとき不成立となること", func(t *testing.T) {
			expectNames(t, ronNames("2m 2m 2m 5m 6m 7m 2p 3p 4p 4p 5p 3s 3s", "6p"), "断么九")
		})
		t.Run("雀頭が三元牌のとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("4p 5p 5z 5z", "6p"), "一気通貫")
		})
		t.Run("雀頭が場風のとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("4p 5p 1z 1z", "6p"), "一気通貫")
		})
		t.Run("雀頭が自風のとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("4p 5p 2z 2z", "6p"), "一気通貫")
		})
		t.Run("嵌張待ちのとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("4p 6p 9s 9s", "5p"), "一気通貫")
		})
		t.Run("辺張待ちのとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("1p 2p 9s 9s", "3p"), "一気通貫")
		})
		t.Run("単騎待ちのとき不成立となること", func(t *testing.T) {
			expectNames(t, besideIttsu("4p 5p 6p 9s", "9s"), "一気通貫")
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			expectNames(t, ronNames("2m 3m 4m 5m 6m 7m 4p 5p 3s 3s", "6p", mt.Chi("2p 3p 4p")), "断么九")
		})
	})

	t.Run("一盃口（イーペーコー）", func(t *testing.T) {
		t.Run("同一の順子2組を含み門前のとき成立し 1翻 となること", func(t *testing.T) {
			expectHas(t, ronNames("2m 3m 4m 2m 3m 4m 5m 6m 7m 1p 2p 9s 9s", "3p"), "一盃口")
			expectHan(t, ronHan("一盃口", "2m 3m 4m 2m 3m 4m 5m 6m 7m 1p 2p 9s 9s", "3p"), 1)
		})
		t.Run("同一の順子が2組そろわないとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2m 3m 4m 3m 4m 5m 5m 6m 7m 1p 2p 9s 9s", "3p"), "一盃口")
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			expectNames(t, ronNames("2m 3m 4m 2m 3m 4m 2p 3p 3s 3s", "4p", mt.Chi("4p 5p 6p")), "断么九")
		})
		t.Run("二盃口が成立するときは一盃口として数えないこと", func(t *testing.T) {
			names := ronNames("2m 3m 4m 2m 3m 4m 6m 7m 8m 6m 7m 8m 9s", "9s")
			expectHas(t, names, "二盃口")
			expectLacks(t, names, "一盃口")
		})
	})

	t.Run("二盃口（リャンペーコー）", func(t *testing.T) {
		t.Run("同一の順子2組を2セット含み門前のとき成立し 3翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("二盃口", "2m 3m 4m 2m 3m 4m 6m 7m 8m 6m 7m 8m 9s", "9s"), 3)
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			expectNames(t, ronNames("2m 3m 4m 6m 7m 8m 6m 7m 8m 3s", "3s", mt.Chi("2m 3m 4m")), "断么九")
		})
		t.Run("七対子として解釈できる形でも、二盃口を優先して数えること", func(t *testing.T) {
			names := ronNames("2m 2m 3m 3m 4m 4m 6m 6m 7m 7m 8m 8m 9s", "9s")
			expectHas(t, names, "二盃口")
			expectLacks(t, names, "七対子")
		})
	})

	t.Run("三色同順（サンショクドウジュン）", func(t *testing.T) {
		t.Run("萬子・筒子・索子で同じ並びの順子3組を含むとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 2m 3m 1p 2p 3p 1s 2s 3s 5m 6m 7m 9s", "9s"), "三色同順")
		})
		t.Run("門前で 2翻、副露で 1翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("三色同順", "1m 2m 3m 1p 2p 3p 1s 2s 3s 5m 6m 7m 9s", "9s"), 2)
			expectHan(t, ronHan("三色同順", "1m 2m 3m 1s 2s 3s 5m 6m 7m 9s", "9s", mt.Chi("1p 2p 3p")), 1)
		})
		t.Run("1色でも並びが異なるとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2m 3m 4m 2p 3p 4p 3s 4s 5s 6m 7m 8m 5s", "5s"), "三色同順")
		})
	})

	t.Run("一気通貫（イッキツウカン）", func(t *testing.T) {
		t.Run("同一色で 123・456・789 の3順子を含むとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s"), "一気通貫")
		})
		t.Run("門前で 2翻、副露で 1翻（食い下がり）となること", func(t *testing.T) {
			expectHan(t, ronHan("一気通貫", "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s"), 2)
			expectHan(t, ronHan("一気通貫", "4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s", mt.Chi("1m 2m 3m")), 1)
		})
		t.Run("123・456・789 が同一色にそろわないとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2m 3m 4m 5m 6m 7m 2p 3p 4p 5p 6p 7p 3s", "3s"), "一気通貫")
		})
	})

	t.Run("三暗刻（サンアンコウ）", func(t *testing.T) {
		t.Run("暗刻を3つ含むとき成立し 2翻 となること", func(t *testing.T) {
			expectHan(t, ronHan("三暗刻", "2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 3p 4p 9s", "9s"), 2)
		})
		t.Run("暗槓も暗刻として数えること", func(t *testing.T) {
			expectHas(t, ronNames("5m 5m 5m 8m 8m 8m 2p 3p 4p 9s", "9s", mt.Ankan("2m 2m 2m 2m")), "三暗刻")
		})
		t.Run("ロン和了で完成した刻子は明刻として扱い、暗刻が2つなら不成立となること", func(t *testing.T) {
			closed, win := "2m 2m 2m 5m 5m 5m 8m 8m 3s 3s 2p 3p 4p", "8m"
			expectLacks(t, yakuNames(structuralWin(closed, win, winning.Ron)), "三暗刻")
			expectHas(t, yakuNames(structuralWin(closed, win, winning.Tsumo)), "三暗刻")
		})
		t.Run("暗刻が2つ以下のとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2m 2m 2m 5m 5m 5m 2p 3p 4p 5p 6p 7p 3s", "3s"), "三暗刻")
		})
	})
}
