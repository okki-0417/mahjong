package knowledge_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// 手牌。closed(手の内の牌) と melds(副露) を所有する。
// 副露は手牌の枠を3枚消費するので、副露が増えるほど手の内は減る。
func TestHand(t *testing.T) {
	t.Run("手牌の枚数", func(t *testing.T) {
		t.Run("副露がないとき手の内は13枚であること", func(t *testing.T) {
			if got := len(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s").ClosedTiles()); got != 13 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("副露が1つ増えるごとに手の内が3枚減ること", func(t *testing.T) {
			if got := len(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", mt.Pon("1z 1z 1z")).ClosedTiles()); got != 10 {
				t.Errorf("one meld: %d", got)
			}
			if got := len(mt.Hand("1m 2m 3m 4m 5m 6m 5s", mt.Pon("1z 1z 1z"), mt.Pon("2z 2z 2z")).ClosedTiles()); got != 7 {
				t.Errorf("two melds: %d", got)
			}
		})
		t.Run("枚数が 13-3N に合わない手牌は作れないこと", func(t *testing.T) {
			_, err := hand.New(mt.Tiles("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s 6s"), []hand.Meld{mt.Pon("1z 1z 1z")})
			if !errors.Is(err, hand.ErrClosedTileCount) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("副露は4つまでであること", func(t *testing.T) {
			melds := []hand.Meld{mt.Pon("1z 1z 1z"), mt.Pon("2z 2z 2z"), mt.Pon("3z 3z 3z"), mt.Pon("4z 4z 4z"), mt.Pon("5z 5z 5z")}
			if _, err := hand.New(nil, melds); !errors.Is(err, hand.ErrTooManyMelds) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("同種を5枚持つ手牌は存在しないこと", func(t *testing.T) {
			_, err := hand.New(mt.Tiles("1m 1m 1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m"), nil)
			if !errors.Is(err, hand.ErrTooManyCopies) {
				t.Fatalf("err = %v", err)
			}
		})
	})

	// 手牌は常に 13-3N 枚なので、切るだけでは枚数が足りなくなる。
	// 打牌は「1枚足されて1枚出る」で初めて釣り合う。
	t.Run("打牌", func(t *testing.T) {
		before := mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s")

		t.Run("手の内の牌を1枚切ると同時に1枚足され、手牌の枚数が変わらないこと", func(t *testing.T) {
			after := must(t)(before.Discard(mt.T("5s"), mt.T("9s")))
			if len(after.ClosedTiles()) != len(before.ClosedTiles()) {
				t.Fatal("count changed")
			}
		})
		t.Run("切った牌が手の内から無くなり、足された牌が手の内に残ること", func(t *testing.T) {
			after := must(t)(before.Discard(mt.T("5s"), mt.T("9s")))
			if contains(after.ClosedTiles(), mt.T("5s")) || !contains(after.ClosedTiles(), mt.T("9s")) {
				t.Fatalf("got %v", after)
			}
		})
		t.Run("足された牌をそのまま切る(ツモ切り)ことができること", func(t *testing.T) {
			after := must(t)(before.Discard(mt.T("9s"), mt.T("9s")))
			if !after.Equal(before) {
				t.Fatalf("got %v", after)
			}
		})
		t.Run("手の内にも足された牌にも無い牌は切れないこと", func(t *testing.T) {
			if _, err := before.Discard(mt.T("7s"), mt.T("9s")); !errors.Is(err, hand.ErrTileNotInHand) {
				t.Fatalf("err = %v", err)
			}
		})
	})

	t.Run("鳴き", func(t *testing.T) {
		t.Run("チーは他家の捨て牌1枚と、自分の手の内の連続する2枚で成立すること", func(t *testing.T) {
			after := must(t)(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s").
				Chi(mt.T("4p"), mt.Tiles("2p 3p"), mt.T("5s")))
			if !reflect.DeepEqual(meldKinds(after), []hand.MeldKind{hand.Chi}) {
				t.Fatalf("melds %v", after.Melds())
			}
			if got := labels(tile.Sorted(after.Melds()[0].Tiles())); got != "2p 3p 4p" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("ポンは他家の捨て牌1枚と、自分の手の内の同じ牌2枚で成立すること", func(t *testing.T) {
			after := must(t)(mt.Hand("1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 5s").
				Pon(mt.T("1m"), mt.Tiles("1m 1m"), mt.T("5s")))
			if !reflect.DeepEqual(meldKinds(after), []hand.MeldKind{hand.Pon}) {
				t.Fatalf("melds %v", after.Melds())
			}
		})
		t.Run("明槓は他家の捨て牌1枚と、自分の手の内の同じ牌3枚で成立すること", func(t *testing.T) {
			after := must(t)(mt.Hand("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p").
				Minkan(mt.T("1m"), mt.Tiles("1m 1m 1m")))
			if !reflect.DeepEqual(meldKinds(after), []hand.MeldKind{hand.Minkan}) || len(after.Melds()[0].Tiles()) != 4 {
				t.Fatalf("melds %v", after.Melds())
			}
		})
		t.Run("暗槓は自分の手の内の同じ牌4枚だけで成立すること", func(t *testing.T) {
			after := must(t)(mt.Hand("1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p").
				Ankan(mt.Tiles("1m 1m 1m 1m"), mt.T("1m")))
			if !reflect.DeepEqual(meldKinds(after), []hand.MeldKind{hand.Ankan}) {
				t.Fatalf("melds %v", after.Melds())
			}
		})
		t.Run("加槓は既にポンしている牌に、自分の手の内の同じ牌1枚を加えて成立すること", func(t *testing.T) {
			after := must(t)(mt.Hand("2m 3m 4m 5m 6m 7m 8m 9m 1p 2p", mt.Pon("1m 1m 1m")).
				Kakan(mt.T("1m"), mt.T("1m")))
			if !reflect.DeepEqual(meldKinds(after), []hand.MeldKind{hand.Minkan}) || len(after.Melds()[0].Tiles()) != 4 {
				t.Fatalf("melds %v", after.Melds())
			}
		})
		t.Run("鳴きに使う牌が手の内に無いときは鳴けないこと", func(t *testing.T) {
			_, err := mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s").
				Pon(mt.T("9s"), mt.Tiles("9s 9s"), mt.T("5s"))
			if err == nil {
				t.Fatal("no error")
			}
		})
		t.Run("鳴いた牌は手の内から副露へ移ること", func(t *testing.T) {
			after := must(t)(mt.Hand("1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 5s").
				Pon(mt.T("1m"), mt.Tiles("1m 1m"), mt.T("5s")))
			if contains(after.ClosedTiles(), mt.T("1m")) {
				t.Fatal("1m still in hand")
			}
			count := 0
			for _, x := range after.Melds()[0].Tiles() {
				if x == mt.T("1m") {
					count++
				}
			}
			if count != 3 {
				t.Fatalf("got %d", count)
			}
		})
	})

	// 副露は手牌の枠を3枚消費する。手の内から出す枚数がそれと食い違う鳴きは、
	// 差を埋める牌（切る牌・自摸牌）が揃って初めて1つの鳴きとして完結する。
	t.Run("鳴きと手牌の枚数", func(t *testing.T) {
		t.Run("チー・ポンは手の内から2枚しか出さないので、1枚切るまでが1つの鳴きであること", func(t *testing.T) {
			if !lastParamIsTile(hand.Hand{}.Chi, 3) || !lastParamIsTile(hand.Hand{}.Pon, 3) {
				t.Fatal("chi/pon must take (called, consumed, discard)")
			}
		})
		t.Run("明槓は手の内から3枚出すので、切らずに枚数が揃うこと", func(t *testing.T) {
			if reflect.TypeOf(hand.Hand{}.Minkan).NumIn() != 2 {
				t.Fatal("minkan must take (called, consumed) only")
			}
		})
		t.Run("暗槓・加槓は手の内から出す枚数が枠と食い違うので、足される牌が揃って初めて成立すること", func(t *testing.T) {
			if !lastParamIsTile(hand.Hand{}.Ankan, 2) || !lastParamIsTile(hand.Hand{}.Kakan, 2) {
				t.Fatal("ankan/kakan must take the added tile")
			}
		})
		t.Run("鳴きの前後で、手の内と副露を合わせた牌の総数は変わらないこと", func(t *testing.T) {
			before := mt.Hand("1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 5s")
			after := must(t)(before.Pon(mt.T("1m"), mt.Tiles("1m 1m"), mt.T("5s")))
			if len(after.AllTiles()) != len(before.AllTiles()) {
				t.Fatalf("%d -> %d", len(before.AllTiles()), len(after.AllTiles()))
			}
		})
	})

	t.Run("鳴きと門前", func(t *testing.T) {
		t.Run("チー・ポン・明槓をすると門前でなくなること", func(t *testing.T) {
			for _, m := range []hand.Meld{mt.Chi("1p 2p 3p"), mt.Pon("1z 1z 1z"), mt.Minkan("1z 1z 1z 1z")} {
				if mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", m).IsMenzen() {
					t.Errorf("%v keeps menzen", m)
				}
			}
		})
		t.Run("暗槓は自分の牌だけで成立するので門前を保つこと", func(t *testing.T) {
			if !mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", mt.Ankan("1z 1z 1z 1z")).IsMenzen() {
				t.Fatal("not menzen")
			}
		})
		t.Run("ポンから加槓した牌は、元のポンが鳴きなので門前を崩したままであること", func(t *testing.T) {
			after := must(t)(mt.Hand("2m 3m 4m 5m 6m 7m 8m 9m 1p 2p", mt.Pon("1m 1m 1m")).Kakan(mt.T("1m"), mt.T("1m")))
			if after.IsMenzen() {
				t.Fatal("menzen")
			}
		})
	})

	t.Run("向聴・受け入れ", func(t *testing.T) {
		t.Run("向聴数は打牌を終えた手牌に対して定まること", func(t *testing.T) {
			before := mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s")
			after := must(t)(before.Discard(mt.T("9s"), mt.T("3p")))
			if !before.IsTenpai() || !after.IsTenpai() {
				t.Fatal("tenpai")
			}
			if got := labels(after.Waits()); got != "3p" {
				t.Fatalf("waits %q", got)
			}
		})
	})
}
