package knowledge_test

import (
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// 役満。翻数ではなく固定の役満点で計算し、通常役・ドラは数えない。
func TestYakuman(t *testing.T) {
	tsumoNames := func(closed, win string, melds ...hand.Meld) []string {
		return yakuNames(structuralWin(closed, win, winning.Tsumo, melds...))
	}

	t.Run("国士無双（コクシムソウ）", func(t *testing.T) {
		t.Run("13種の么九牌を各1枚以上、いずれか1種を2枚（雀頭）そろえたとき成立すること", func(t *testing.T) {
			expectNames(t, ronNames("1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z", "7z"), "国士無双")
		})
		t.Run("門前限定であること", func(t *testing.T) {
			if got := ronNames("1m 9m 1p 9p 1s 9s 2z 3z 4z 5z", "6z", mt.Pon("1z 1z 1z")); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
	})

	t.Run("四暗刻（スーアンコウ）", func(t *testing.T) {
		t.Run("暗刻4つ（暗槓を含む）+ 雀頭 のとき成立すること", func(t *testing.T) {
			expectNames(t, tsumoNames("2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 2p 9s 9s", "2p"), "四暗刻")
			expectNames(t, tsumoNames("5m 5m 5m 8m 8m 8m 2p 2p 9s 9s", "2p", mt.Ankan("2m 2m 2m 2m")), "四暗刻")
		})
		t.Run("シャンポン待ちをロン和了すると和了牌の刻子が明刻となり、四暗刻は不成立（対々和・三暗刻扱い）となること", func(t *testing.T) {
			names := ronNames("2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 2p 9s 9s", "2p")
			expectLacks(t, names, "四暗刻")
			expectHas(t, names, "対々和")
			expectHas(t, names, "三暗刻")
		})
	})

	t.Run("大三元（ダイサンゲン）", func(t *testing.T) {
		t.Run("白・發・中すべてを刻子（槓子）でそろえたとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("5z 5z 5z 6z 6z 6z 7z 7z 7z 2m 3m 4m 9s", "9s"), "大三元")
		})
	})

	t.Run("字一色（ツーイーソー）", func(t *testing.T) {
		t.Run("字牌のみで構成されるとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1z 1z 1z 2z 2z 2z 5z 5z 5z 6z 6z 6z 7z", "7z"), "字一色")
		})
	})

	t.Run("緑一色（リューイーソー）", func(t *testing.T) {
		t.Run("緑の牌（索子の 2・3・4・6・8 と 發）のみで構成されるとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("2s 2s 2s 3s 3s 3s 4s 4s 4s 6s 6s 6s 8s", "8s"), "緑一色")
		})
		t.Run("索子でも 1・5・7・9 を含むとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("2s 2s 2s 3s 3s 3s 4s 4s 4s 5s 5s 5s 8s", "8s"), "緑一色")
		})
	})

	t.Run("清老頭（チンロウトウ）", func(t *testing.T) {
		t.Run("老頭牌（1・9）の刻子のみで構成されるとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 1m 1m 9m 9m 9m 1p 1p 1p 9p 9p 9p 1s", "1s"), "清老頭")
		})
		t.Run("字牌を含むとき不成立（混老頭）となること", func(t *testing.T) {
			names := ronNames("1m 1m 1m 9m 9m 9m 1p 1p 1p 1s", "1s", mt.Pon("1z 1z 1z"))
			expectLacks(t, names, "清老頭")
			expectHas(t, names, "混老頭")
		})
	})

	t.Run("四槓子（スーカンツ）", func(t *testing.T) {
		t.Run("槓子を4つそろえたとき成立すること", func(t *testing.T) {
			names := ronNames("9s", "9s", mt.Minkan("1m 1m 1m 1m"), mt.Minkan("2m 2m 2m 2m"), mt.Ankan("3m 3m 3m 3m"), mt.Ankan("4m 4m 4m 4m"))
			expectHas(t, names, "四槓子")
		})
	})

	t.Run("小四喜（ショウスーシー）", func(t *testing.T) {
		t.Run("四風牌のうち3種を刻子、残り1種を雀頭にしたとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1z 1z 1z 2z 2z 2z 3z 3z 3z 4z 2m 3m 4m", "4z"), "小四喜")
		})
	})

	t.Run("大四喜（ダイスーシー）", func(t *testing.T) {
		t.Run("四風牌すべてを刻子（槓子）でそろえたとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1z 1z 1z 2z 2z 2z 3z 3z 3z 4z 4z 5z 5z", "4z"), "大四喜")
		})
	})

	t.Run("九蓮宝燈（チューレンポウトウ）", func(t *testing.T) {
		t.Run("同一色で 1112345678999 + 同色1枚 の門前形のとき成立すること", func(t *testing.T) {
			expectHas(t, ronNames("1m 1m 1m 2m 3m 4m 5m 5m 6m 7m 8m 9m 9m", "9m"), "九蓮宝燈")
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			expectLacks(t, ronNames("1m 1m 1m 2m 3m 4m 5m 5m 9m 9m", "9m", mt.Chi("6m 7m 8m")), "九蓮宝燈")
		})
	})

	t.Run("天和（テンホウ）", func(t *testing.T) {
		t.Run("親が配牌の時点で和了しているとき成立すること", func(t *testing.T) {
			s := sit(winning.Tsumo, tile.EastWind, tile.EastWind)
			s.Tenhou = true
			expectHas(t, yakuNames(winOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s", nil, s, ruleset.Default())), "天和")
		})
	})

	t.Run("地和（チーホウ）", func(t *testing.T) {
		t.Run("子が第一自摸で和了したとき成立すること", func(t *testing.T) {
			s := sit(winning.Tsumo, tile.EastWind, tile.SouthWind)
			s.Chiihou = true
			expectHas(t, yakuNames(winOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s", nil, s, ruleset.Default())), "地和")
		})
	})

	t.Run("役満の扱い", func(t *testing.T) {
		// 通常なら 門前清自摸和・対々和・三暗刻 が付く形を、四暗刻の成立で塗り替えて見る。
		suuankou := func(t *testing.T, doraCount int) winning.Score {
			t.Helper()
			w, err := structuralWin("2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 2p 9s 9s", "2p", winning.Tsumo)
			if err != nil {
				t.Fatal(err)
			}
			return w.Score(doraCount)
		}

		t.Run("役満成立時は通常役・ドラを数えず、役満点で計算すること", func(t *testing.T) {
			score := suuankou(t, 0)
			expectNames(t, yakuNamesOf(score), "四暗刻")
			if score.Han() != 0 || score.Fu() != 0 || score.RankLabel() != "役満" {
				t.Fatalf("han %d fu %d rank %q", score.Han(), score.Fu(), score.RankLabel())
			}
			if suuankou(t, 3).Total() != score.Total() {
				t.Fatal("dora changed a yakuman")
			}
		})
		t.Run("複数の役満はダブル役満・トリプル役満として加算すること", func(t *testing.T) {
			w, err := structuralWin("1z 1z 1z 2z 2z 2z 3z 3z 3z 4z 4z 5z 5z", "4z", winning.Ron)
			if err != nil {
				t.Fatal(err)
			}
			score := w.Score(0)
			names := yakuNamesOf(score)
			if len(names) != 2 || !has(names, "大四喜") || !has(names, "字一色") {
				t.Fatalf("got %v", names)
			}
			if score.YakumanCount() != 2 || score.RankLabel() != "ダブル役満" {
				t.Fatalf("yakuman %d rank %q", score.YakumanCount(), score.RankLabel())
			}
		})
	})
}

func yakuNamesOf(score winning.Score) []string {
	var names []string
	for _, y := range score.Yakus() {
		names = append(names, y.Name)
	}
	return names
}

// 特定の待ち・純粋形で通常の役満を上書きするダブル役満。
func TestCompositeYakuman(t *testing.T) {
	scoreOf := func(t *testing.T, closed, win string, kind winning.WinKind) winning.Score {
		t.Helper()
		w, err := structuralWin(closed, win, kind)
		if err != nil {
			t.Fatal(err)
		}
		return w.Score(0)
	}

	t.Run("国士無双十三面（コクシジュウサンメン）", func(t *testing.T) {
		// 13種が1枚ずつそろった形では、どの么九牌を和了っても雀頭になる。
		const juusanmen = "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"

		t.Run("13種すべてを1枚ずつそろえ、13面待ちのいずれかで和了したときダブル役満となること", func(t *testing.T) {
			if scoreOf(t, juusanmen, "1m", winning.Ron).YakumanCount() != 2 || scoreOf(t, juusanmen, "7z", winning.Ron).YakumanCount() != 2 {
				t.Fatal("not a double yakuman")
			}
		})
		t.Run("通常の国士無双を上書きすること（両方は数えない）", func(t *testing.T) {
			expectNames(t, yakuNamesOf(scoreOf(t, juusanmen, "1m", winning.Ron)), "国士無双13面待ち")
		})
	})

	t.Run("四暗刻単騎（スーアンコウタンキ）", func(t *testing.T) {
		const tenpai = "2m 2m 2m 5m 5m 5m 8m 8m 8m 2p 2p 2p 9s"

		t.Run("暗刻4つが完成済みで、単騎待ちを和了したときダブル役満となること", func(t *testing.T) {
			if scoreOf(t, tenpai, "9s", winning.Tsumo).YakumanCount() != 2 {
				t.Fatal("not a double yakuman")
			}
		})
		t.Run("単騎ロンでも和了牌は雀頭側なので、四暗刻が崩れず成立すること", func(t *testing.T) {
			expectHas(t, yakuNamesOf(scoreOf(t, tenpai, "9s", winning.Ron)), "四暗刻単騎")
		})
		t.Run("通常の四暗刻を上書きすること", func(t *testing.T) {
			expectLacks(t, yakuNamesOf(scoreOf(t, tenpai, "9s", winning.Ron)), "四暗刻")
		})
	})

	t.Run("純正九蓮宝燈（ジュンセイチューレンポウトウ）", func(t *testing.T) {
		const tenpai = "1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m"

		t.Run("1112345678999 の純正形で、9面待ちのいずれかを和了したときダブル役満となること", func(t *testing.T) {
			if scoreOf(t, tenpai, "5m", winning.Ron).YakumanCount() != 2 || scoreOf(t, tenpai, "1m", winning.Ron).YakumanCount() != 2 {
				t.Fatal("not a double yakuman")
			}
		})
		t.Run("通常の九蓮宝燈を上書きすること", func(t *testing.T) {
			expectLacks(t, yakuNamesOf(scoreOf(t, tenpai, "5m", winning.Ron)), "九蓮宝燈")
		})
	})
}
