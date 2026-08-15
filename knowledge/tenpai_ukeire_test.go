package knowledge_test

import (
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
	"github.com/okki-0417/mahjong/winning"
)

// 受け入れ（有効牌）= 引くと向聴数が1つ進む牌と、その残り枚数。
func TestUkeire(t *testing.T) {
	ukeireOf := func(closed, seen string) ukeire.Ukeire {
		h := mt.Hand(closed)
		if seen == "" {
			return ukeire.OfHand(h)
		}
		return ukeire.Of(h, tile.MustSupply(append(mt.Tiles(closed), mt.Tiles(seen)...)))
	}
	entryTiles := func(u ukeire.Ukeire) []tile.Tile {
		var out []tile.Tile
		for _, e := range u.Entries() {
			out = append(out, e.Tile)
		}
		return out
	}

	t.Run("有効牌の定義", func(t *testing.T) {
		t.Run("引くことで向聴数が1つ進む牌がすべて有効牌となること", func(t *testing.T) {
			before := mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s")
			for _, x := range before.ImprovingTiles() {
				after := hand.ShantenOf(append(before.ClosedTiles(), x), nil)
				if after >= before.Shanten() {
					t.Errorf("%v does not improve", x)
				}
			}
		})
		t.Run("テンパイのとき、和了できる牌が有効牌となること", func(t *testing.T) {
			u := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "")
			if labels(entryTiles(u)) != "5s" {
				t.Fatalf("got %v", entryTiles(u))
			}
			if !reflect.DeepEqual(entryTiles(u), mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s").Waits()) {
				t.Fatal("entries differ from waits")
			}
		})
	})

	t.Run("残り枚数", func(t *testing.T) {
		t.Run("自分の手牌に使われている分を除いた残り枚数を数えること", func(t *testing.T) {
			if got := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "").Entries()[0].Remaining; got != tile.CopiesPerKind-1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("他家の副露・河・ドラ表示牌に出ている分も除くこと", func(t *testing.T) {
			if got := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s").Entries()[0].Remaining; got != tile.CopiesPerKind-2 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("有効牌が場に出尽くしているとき、その牌の残り枚数は0枚となること", func(t *testing.T) {
			u := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s", "9s 9s 9s")
			if labels(entryTiles(u)) != "9s" || u.Entries()[0].Remaining != 0 || u.RemainingTotal() != 0 {
				t.Fatalf("got %+v total %d", u.Entries(), u.RemainingTotal())
			}
		})
		t.Run("どの牌が有効牌かは手牌の形だけで決まり、場に何が出ていても変わらないこと", func(t *testing.T) {
			empty := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "")
			crowded := ukeireOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s 0s")
			if !reflect.DeepEqual(entryTiles(empty), entryTiles(crowded)) {
				t.Fatal("tiles differ")
			}
		})
	})

	t.Run("待ちごとの有効牌の数", func(t *testing.T) {
		t.Run("両面待ちの有効牌は2種となること", func(t *testing.T) {
			if got := len(ukeireOf("3m 4m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s", "").Entries()); got != 2 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("嵌張・辺張・単騎待ちの有効牌は1種となること", func(t *testing.T) {
			for _, closed := range []string{
				"3m 5m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s",
				"1m 2m 1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 1s",
				"1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s",
			} {
				if got := len(ukeireOf(closed, "").Entries()); got != 1 {
					t.Errorf("%s: %d", closed, got)
				}
			}
		})
		t.Run("シャンポン待ちの有効牌は2種となること", func(t *testing.T) {
			if got := len(ukeireOf("1m 1m 5s 5s 1p 2p 3p 4p 5p 6p 7p 8p 9p", "").Entries()); got != 2 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("多面張", func(t *testing.T) {
		t.Run("1つのテンパイが複数の待ちを含むとき、その和集合が有効牌となること", func(t *testing.T) {
			u := ukeireOf("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m", "")
			kinds := map[winning.WaitKind]bool{}
			for _, e := range u.Entries() {
				for _, k := range u.WaitKinds(e.Tile) {
					kinds[k] = true
				}
			}
			if !kinds[winning.Ryanmen] || !kinds[winning.Penchan] || !kinds[winning.Shanpon] {
				t.Fatalf("got %v", kinds)
			}
		})
		t.Run("九蓮宝燈形は同一色の9種すべてが有効牌となること", func(t *testing.T) {
			if got := labels(entryTiles(ukeireOf("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m 9m", ""))); got != "1m 2m 3m 4m 5m 6m 7m 8m 9m" {
				t.Fatalf("got %q", got)
			}
		})
	})
}
