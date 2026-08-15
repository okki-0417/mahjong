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

// 役牌（ヤクハイ）。特定の字牌の刻子・槓子が1つにつき1翻を生む。
func TestYakuhai(t *testing.T) {
	names := func(closed, win string, seat tile.Wind, melds ...hand.Meld) []string {
		return yakuNames(winOf(closed, win, melds, sit(winning.Ron, tile.EastWind, seat), ruleset.Default()))
	}
	// 役牌そのものだけを見たいので、他の役が付かない形に役牌の刻子を1つ置く。
	withKoutsu := func(labels string) []string {
		return names(labels+" 1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", tile.SouthWind)
	}
	// 役牌にならない風牌の刻子だけでは和了れないので、一気通貫を並べて和了形にする。
	besideIttsu := func(labels string) []string {
		return names("1m 2m 3m 4m 5m 6m 7m 8m 9m "+labels+" 1p", "1p", tile.SouthWind)
	}
	expectNames := func(t *testing.T, got []string, want ...string) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	t.Run("三元牌", func(t *testing.T) {
		t.Run("白の刻子・槓子で 1翻 となること", func(t *testing.T) {
			expectNames(t, withKoutsu("5z 5z 5z"), "役牌（白）")
			expectNames(t, names("1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", tile.SouthWind, mt.Minkan("5z 5z 5z 5z")), "役牌（白）")
		})
		t.Run("發の刻子・槓子で 1翻 となること", func(t *testing.T) {
			expectNames(t, withKoutsu("6z 6z 6z"), "役牌（發）")
			expectNames(t, names("1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", tile.SouthWind, mt.Minkan("6z 6z 6z 6z")), "役牌（發）")
		})
		t.Run("中の刻子・槓子で 1翻 となること", func(t *testing.T) {
			expectNames(t, withKoutsu("7z 7z 7z"), "役牌（中）")
			expectNames(t, names("1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", tile.SouthWind, mt.Minkan("7z 7z 7z 7z")), "役牌（中）")
		})
		t.Run("三元牌が雀頭のときは役牌として成立しないこと", func(t *testing.T) {
			expectNames(t, names("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5z", "5z", tile.SouthWind), "一気通貫")
		})
	})

	t.Run("場風", func(t *testing.T) {
		t.Run("場風牌の刻子・槓子で 1翻 となること", func(t *testing.T) {
			expectNames(t, withKoutsu("1z 1z 1z"), "場風")
		})
		t.Run("場風でない風牌の刻子では成立しないこと", func(t *testing.T) {
			expectNames(t, besideIttsu("3z 3z 3z"), "一気通貫")
		})
	})

	t.Run("自風", func(t *testing.T) {
		t.Run("自風牌の刻子・槓子で 1翻 となること", func(t *testing.T) {
			expectNames(t, withKoutsu("2z 2z 2z"), "自風")
		})
		t.Run("自風でない風牌の刻子では成立しないこと", func(t *testing.T) {
			expectNames(t, besideIttsu("4z 4z 4z"), "一気通貫")
		})
	})

	t.Run("連風牌（ダブ東・ダブ南など）", func(t *testing.T) {
		t.Run("場風かつ自風の風牌の刻子は 場風1翻 + 自風1翻 = 2翻 となること", func(t *testing.T) {
			w, err := winOf("1z 1z 1z 1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", nil, sit(winning.Ron, tile.EastWind, tile.EastWind), ruleset.Default())
			if err != nil {
				t.Fatal(err)
			}
			yakus := w.Score(0).Yakus()
			got := yakuNames(w, nil)
			if len(got) != 2 || !has(got, "場風") || !has(got, "自風") {
				t.Fatalf("got %v", got)
			}
			if yakus[0].Han+yakus[1].Han != 2 {
				t.Fatalf("han %d", yakus[0].Han+yakus[1].Han)
			}
		})
	})

	t.Run("役牌は副露していても成立すること（食い下がりなし）", func(t *testing.T) {
		w, err := winOf("1m 2m 3m 4p 5p 6p 7s 8s 9s 2p", "2p", []hand.Meld{mt.Pon("5z 5z 5z")}, sit(winning.Ron, tile.EastWind, tile.SouthWind), ruleset.Default())
		if err != nil {
			t.Fatal(err)
		}
		expectNames(t, yakuNames(w, nil), "役牌（白）")
		if hanOf("役牌（白）")(w, nil) != 1 {
			t.Fatal("han")
		}
	})
}
