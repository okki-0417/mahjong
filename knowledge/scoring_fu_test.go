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

// 着目する面子以外は順子と数牌の単騎だけの手。
const (
	tanyaoRest  = "3p 4p 5p 6p 7p 8p 5s 6s 7s 8s"
	ittsuRest   = "1p 2p 3p 4p 5p 6p 7p 8p 9p 9s"
	yakuhaiRest = "1p 2p 3p 7p 8p 9p 5s 6s 7s 9s"
)

// 符計算のルール。和了形の各要素（面子・雀頭・待ち・和了方法）に付く符を積み上げ、
// 最後に10符単位へ切り上げる。翻数ではなく「符」がどう決まるかだけを描く。
func TestFu(t *testing.T) {
	// 一意に読める実際の和了から符を取る。門前手はリーチで役を確保し、
	// 副露手は断么九か役牌で確保する。リーチは符に影響しない。
	fuOf := func(t *testing.T, closed, win string, kind winning.WinKind, round, seat tile.Wind, melds ...hand.Meld) winning.Fu {
		t.Helper()
		s := sit(kind, round, seat)
		s.Riichi = len(melds) == 0
		w, err := winOf(closed, win, melds, s, ruleset.Default())
		if err != nil {
			t.Fatal(err)
		}
		return w.Fu()
	}
	ron := func(t *testing.T, closed, win string, melds ...hand.Meld) winning.Fu {
		t.Helper()
		return fuOf(t, closed, win, winning.Ron, tile.EastWind, tile.SouthWind, melds...)
	}
	tsumo := func(t *testing.T, closed, win string, melds ...hand.Meld) winning.Fu {
		t.Helper()
		return fuOf(t, closed, win, winning.Tsumo, tile.EastWind, tile.SouthWind, melds...)
	}
	// 発生源が無ければ 0符。
	fuFrom := func(fu winning.Fu, kind winning.FuSourceKind) int {
		total := 0
		for _, s := range fu.Sources() {
			if s.Kind == kind {
				total += s.Fu
			}
		}
		return total
	}
	expectFu := func(t *testing.T, got, want int) {
		t.Helper()
		if got != want {
			t.Errorf("fu %d, want %d", got, want)
		}
	}

	t.Run("副底", func(t *testing.T) {
		t.Run("すべての和了は副底 20符から始まること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, "2m 2m 2m "+ittsuRest, "9s"), winning.FuBase), 20)
		})
	})

	t.Run("面子の符", func(t *testing.T) {
		t.Run("順子は 0符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, "2m 3m 4m 5m 6m 7m 2p 3p 4p 5p 6p 7p 9s", "9s"), winning.FuMentsu), 0)
		})
		t.Run("中張牌の明刻は 2符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, tanyaoRest, "8s", mt.Pon("2m 2m 2m")), winning.FuMentsu), 2)
		})
		t.Run("中張牌の暗刻は 4符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, "2m 2m 2m "+ittsuRest, "9s"), winning.FuMentsu), 4)
		})
		t.Run("么九牌の明刻は 4符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, yakuhaiRest, "9s", mt.Pon("1z 1z 1z")), winning.FuMentsu), 4)
		})
		t.Run("么九牌の暗刻は 8符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, "1z 1z 1z "+ittsuRest, "9s"), winning.FuMentsu), 8)
		})
		t.Run("中張牌の明槓は 8符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, tanyaoRest, "8s", mt.Minkan("2m 2m 2m 2m")), winning.FuMentsu), 8)
		})
		t.Run("中張牌の暗槓は 16符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, ittsuRest, "9s", mt.Ankan("2m 2m 2m 2m")), winning.FuMentsu), 16)
		})
		t.Run("么九牌の明槓は 16符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, yakuhaiRest, "9s", mt.Minkan("1z 1z 1z 1z")), winning.FuMentsu), 16)
		})
		t.Run("么九牌の暗槓は 32符 となること", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, ittsuRest, "9s", mt.Ankan("1z 1z 1z 1z")), winning.FuMentsu), 32)
		})
	})

	t.Run("雀頭の符", func(t *testing.T) {
		// 一通の3順子 + 234s に雀頭の単騎和了を足した手。
		pairFu := func(t *testing.T, pair string, round, seat tile.Wind) int {
			t.Helper()
			return fuFrom(fuOf(t, "1p 2p 3p 4p 5p 6p 7p 8p 9p 2s 3s 4s "+pair, pair, winning.Ron, round, seat), winning.FuPair)
		}

		t.Run("数牌・非役牌の字牌の雀頭は 0符 となること", func(t *testing.T) {
			expectFu(t, pairFu(t, "9s", tile.EastWind, tile.SouthWind), 0)
			expectFu(t, pairFu(t, "4z", tile.EastWind, tile.SouthWind), 0)
		})
		t.Run("三元牌の雀頭は 2符 となること", func(t *testing.T) {
			expectFu(t, pairFu(t, "5z", tile.EastWind, tile.SouthWind), 2)
		})
		t.Run("場風の雀頭は 2符 となること", func(t *testing.T) {
			expectFu(t, pairFu(t, "1z", tile.EastWind, tile.SouthWind), 2)
		})
		t.Run("自風の雀頭は 2符 となること", func(t *testing.T) {
			expectFu(t, pairFu(t, "2z", tile.EastWind, tile.SouthWind), 2)
		})
		t.Run("場風かつ自風の連風牌の雀頭も 2符 となること（連風でも加算しない）", func(t *testing.T) {
			expectFu(t, pairFu(t, "1z", tile.EastWind, tile.EastWind), 2)
		})
	})

	t.Run("待ちの符", func(t *testing.T) {
		waitFu := func(t *testing.T, closed, win string) int {
			t.Helper()
			return fuFrom(ron(t, closed, win), winning.FuWait)
		}

		t.Run("両面待ちは 0符 となること", func(t *testing.T) {
			expectFu(t, waitFu(t, "2m 3m 5p 5p 1p 2p 3p 7p 8p 9p 2s 3s 4s", "4m"), 0)
		})
		t.Run("双碰（シャンポン）待ちは 0符 となること", func(t *testing.T) {
			expectFu(t, waitFu(t, "5z 5z 2z 2z 1p 2p 3p 4p 5p 6p 7p 8p 9p", "5z"), 0)
		})
		t.Run("嵌張待ちは 2符 となること", func(t *testing.T) {
			expectFu(t, waitFu(t, "1p 3p 5p 5p 4m 5m 6m 7m 8m 9m 2s 3s 4s", "2p"), 2)
		})
		t.Run("辺張待ちは 2符 となること", func(t *testing.T) {
			expectFu(t, waitFu(t, "1p 2p 5p 5p 4m 5m 6m 7m 8m 9m 2s 3s 4s", "3p"), 2)
		})
		t.Run("単騎待ちは 2符 となること", func(t *testing.T) {
			expectFu(t, waitFu(t, "1p 2p 3p 4p 5p 6p 7p 8p 9p 2s 3s 4s 9s", "9s"), 2)
		})
	})

	t.Run("和了方法の符", func(t *testing.T) {
		t.Run("門前ロンは 門前加符 10符が付くこと", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, "2m 2m 2m "+ittsuRest, "9s"), winning.FuMenzenRon), 10)
		})
		t.Run("副露していると門前加符は付かないこと", func(t *testing.T) {
			expectFu(t, fuFrom(ron(t, tanyaoRest, "8s", mt.Pon("2m 2m 2m")), winning.FuMenzenRon), 0)
		})
		t.Run("ツモは 門前・副露を問わず 2符が付くこと", func(t *testing.T) {
			expectFu(t, fuFrom(tsumo(t, "2m 2m 2m "+ittsuRest, "9s"), winning.FuTsumo), 2)
			expectFu(t, fuFrom(tsumo(t, tanyaoRest, "8s", mt.Pon("2m 2m 2m")), winning.FuTsumo), 2)
		})
	})

	t.Run("定型の固定符", func(t *testing.T) {
		t.Run("七対子は 25符固定（切り上げなし）となること", func(t *testing.T) {
			expectFu(t, ron(t, "1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z", "7z").Total(), 25)
		})
		t.Run("平和ツモは 20符固定（ツモ符が付かない）となること", func(t *testing.T) {
			expectFu(t, tsumo(t, "3m 4m 5m 6m 7m 8m 1p 2p 3p 4p 5p 7s 7s", "6p").Total(), 20)
		})
		t.Run("平和ロンは 副底20 + 門前加符10 = 30符 となること", func(t *testing.T) {
			expectFu(t, ron(t, "3m 4m 5m 6m 7m 8m 1p 2p 3p 4p 5p 7s 7s", "6p").Total(), 30)
		})
		t.Run("副露して符源が無い形（食い平和形）のロンは 20符ではなく 30符となること", func(t *testing.T) {
			expectFu(t, ron(t, "3p 4p 5p 6p 7p 8p 5s 5s 6s 7s", "8s", mt.Chi("2m 3m 4m")).Total(), 30)
		})
	})

	t.Run("端数処理", func(t *testing.T) {
		t.Run("積み上げた符は 10符単位へ切り上げること（例 32符 → 40符）", func(t *testing.T) {
			// 副底20 + 字牌暗刻8 + 嵌張2 + ツモ2 = 32
			fu := tsumo(t, "1z 1z 1z 1p 3p 5p 5p 4m 5m 6m 2s 3s 4s", "2p")
			if fu.Subtotal() != 32 || fu.Total() != 40 {
				t.Fatalf("subtotal %d total %d", fu.Subtotal(), fu.Total())
			}
		})
	})

	t.Run("和了の符", func(t *testing.T) {
		// 222m333m444m は暗刻3つとも順子3つとも読める。暗刻に読めば中張暗刻が4符ずつ付く。
		multiForm := func(t *testing.T) *winning.Winning {
			t.Helper()
			w, err := winOf("2m 2m 2m 3m 3m 3m 4m 4m 4m 5p 6p 7p 8s", "8s", nil, sit(winning.Ron, tile.EastWind, tile.SouthWind), ruleset.Default())
			if err != nil {
				t.Fatal(err)
			}
			return w
		}

		t.Run("面子の読み方が複数ある場合", func(t *testing.T) {
			t.Run("最も符の高い読み方の符となること（暗刻3つ 50符 / 順子3つ 40符）", func(t *testing.T) {
				expectFu(t, multiForm(t).Fu().Total(), 50)
			})
			t.Run("符の発生源も、その読み方のものとなること", func(t *testing.T) {
				expectFu(t, fuFrom(multiForm(t).Fu(), winning.FuMentsu), 12)
			})
		})
		t.Run("和了できない手には符が無いこと（役が無ければ和了ではない）", func(t *testing.T) {
			_, err := winOf("3m 3m 3m 4p 5p 6p 7s 8s 9s 2m 3m 4m 2p", "2p", nil, sit(winning.Ron, tile.EastWind, tile.SouthWind), ruleset.Default())
			if !errors.Is(err, winning.ErrNotWinning) {
				t.Fatalf("err = %v", err)
			}
		})
	})
}
