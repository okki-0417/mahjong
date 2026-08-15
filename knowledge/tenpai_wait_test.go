package knowledge_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/winning"
)

func waitsOf(closed string, melds ...hand.Meld) string {
	return labels(mt.Hand(closed, melds...).Waits())
}

func formsOf(closed, win string, melds ...hand.Meld) []winning.Form {
	return winning.Forms(mt.Hand(closed, melds...), mt.T(win), winning.Situation{})
}

func waitKindsOf(closed, win string, melds ...hand.Meld) []winning.WaitKind {
	seen := map[winning.WaitKind]bool{}
	var out []winning.WaitKind
	for _, f := range formsOf(closed, win, melds...) {
		if k := f.WaitKind(); !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

func expectWaitKinds(t *testing.T, got []winning.WaitKind, want ...winning.WaitKind) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// テンパイの手牌が「どの牌で和了できるか」と「待ちの種類」の知識。
// 待ちの種類は符計算（両面0符 / 嵌張・辺張・単騎2符）にも効く。
func TestWait(t *testing.T) {
	t.Run("両面待ち（リャンメン）", func(t *testing.T) {
		t.Run("連続する2牌（例 3m4m）で、両端の2種（2m・5m）を待つ形となること", func(t *testing.T) {
			if got := waitsOf("3m 4m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s"); got != "2m 5m" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("有効牌は2種となること", func(t *testing.T) {
			if got := len(mt.Hand("3m 4m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s").Waits()); got != 2 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("待ちの種別は両面であること", func(t *testing.T) {
			expectWaitKinds(t, waitKindsOf("3m 4m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s", "2m"), winning.Ryanmen)
		})
	})

	t.Run("嵌張待ち（カンチャン）", func(t *testing.T) {
		t.Run("1つ飛ばしの2牌（例 3m5m）で、間の1種（4m）を待つ形となること", func(t *testing.T) {
			if got := waitsOf("3m 5m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s"); got != "4m" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("待ちの種別は嵌張であること", func(t *testing.T) {
			expectWaitKinds(t, waitKindsOf("3m 5m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s", "4m"), winning.Kanchan)
		})
	})

	t.Run("辺張待ち（ペンチャン）", func(t *testing.T) {
		t.Run("12 で 3、または 89 で 7 のみを待つ形となること", func(t *testing.T) {
			if got := waitsOf("1m 2m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s"); got != "3m" {
				t.Errorf("12: %q", got)
			}
			if got := waitsOf("8m 9m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s"); got != "7m" {
				t.Errorf("89: %q", got)
			}
		})
		t.Run("待ちの種別は辺張であること", func(t *testing.T) {
			expectWaitKinds(t, waitKindsOf("1m 2m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s", "3m"), winning.Penchan)
		})
	})

	t.Run("単騎待ち（タンキ）", func(t *testing.T) {
		t.Run("4面子が完成し、雀頭が1牌のときその1種を待つ形となること", func(t *testing.T) {
			if got := waitsOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"); got != "5s" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("待ちの種別は単騎であること", func(t *testing.T) {
			expectWaitKinds(t, waitKindsOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s"), winning.Tanki)
		})
	})

	t.Run("双碰待ち（シャンポン）", func(t *testing.T) {
		t.Run("対子2組で、どちらかが刻子になる2種を待つ形となること", func(t *testing.T) {
			if got := waitsOf("1m 1m 5s 5s 1p 2p 3p 4p 5p 6p 7p 8p 9p"); got != "1m 5s" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("待ちの種別は双碰であること", func(t *testing.T) {
			expectWaitKinds(t, waitKindsOf("1m 1m 5s 5s 1p 2p 3p 4p 5p 6p 7p 8p 9p", "1m"), winning.Shanpon)
		})
		t.Run("和了した側は刻子、残った側は雀頭となること", func(t *testing.T) {
			form := formsOf("1m 1m 5s 5s 1p 2p 3p 4p 5p 6p 7p 8p 9p", "1m")[0]
			var koutsu []string
			for _, m := range form.Mentsu() {
				if m.Kind() == winning.Koutsu {
					koutsu = append(koutsu, m.Tiles()[0].String())
				}
			}
			pair, _ := form.PairTile()
			if !reflect.DeepEqual(koutsu, []string{"1m"}) || pair.String() != "5s" {
				t.Fatalf("koutsu %v pair %v", koutsu, pair)
			}
		})
	})

	t.Run("多面待ち", func(t *testing.T) {
		t.Run("1つの手牌が複数の待ちに解釈できるとき、すべての待ちを列挙すること", func(t *testing.T) {
			// 234567m の伸びで、両端と内側の3種を待つ。
			if got := waitsOf("2m 3m 4m 5m 6m 7m 1p 2p 3p 4p 5p 9p 9p"); got != "3p 6p" {
				t.Errorf("got %q", got)
			}
			if got := waitsOf("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m"); got != "1m 2m 3m 4m 5m 6m 7m 8m 9m" {
				t.Errorf("chuuren: %q", got)
			}
		})
		t.Run("同じ和了牌でも解釈により待ちの種類が変わりうること", func(t *testing.T) {
			// 九蓮宝燈形の 3m は、12m の辺張とも 45m の両面とも読める。
			kinds := waitKindsOf("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m", "3m")
			sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
			expectWaitKinds(t, kinds, winning.Ryanmen, winning.Penchan)
		})
	})
}
