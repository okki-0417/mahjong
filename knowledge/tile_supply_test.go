package knowledge_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// 牌の供給。同じ牌は1種につき4枚しかない。
// 「この牌はまだ使えるか」は牌単体でも手牌単体でも答えられず、
// 場に見えている牌が揃って初めて答えられる。
func TestTileSupply(t *testing.T) {
	supply := func(seen string) tile.Supply {
		return tile.MustSupply(mt.Tiles(seen))
	}

	// 席0が 1m を切り、席2が 1m をポンして 9p を切ったところ。
	board := mt.BuildKyoku(mt.KyokuSpec{
		Hands: map[int]string{
			0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "1m 1m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s",
			1: "2s 3s 4s 6s 7s 8s 2p 3p 4p 5p 6p 7p 8p",
		},
		Draws: "5z", Dora: "6z",
		Actions: []kyoku.Action{mt.DiscardAction(0, "1m"), mt.PonAction(2, "1m", "1m 1m"), mt.DiscardAction(2, "9p")},
	}).Kyokumen()

	t.Run("見えている牌", func(t *testing.T) {
		t.Run("自分の手牌が見えている牌に数えられること", func(t *testing.T) {
			if board.TileSupply(1).Remaining(board.Seat(1).Hand().ClosedTiles()[0]) >= tile.CopiesPerKind {
				t.Fatal("own hand not counted")
			}
		})
		t.Run("他家の副露が見えている牌に数えられること", func(t *testing.T) {
			if board.TileSupply(1).Remaining(mt.T("1m")) != 1 {
				t.Fatal("meld not counted")
			}
		})
		t.Run("河に捨てられた牌が見えている牌に数えられること", func(t *testing.T) {
			if board.TileSupply(1).Remaining(mt.T("9p")) >= tile.CopiesPerKind {
				t.Fatal("river not counted")
			}
		})
		t.Run("公開されているドラ表示牌が見えている牌に数えられること", func(t *testing.T) {
			if board.TileSupply(1).Remaining(mt.T("6z")) != tile.CopiesPerKind-1 {
				t.Fatal("indicator not counted")
			}
		})
		t.Run("誰の手にも河にも現れていない牌は見えていないこと", func(t *testing.T) {
			if board.TileSupply(1).Remaining(mt.T("7z")) != tile.CopiesPerKind {
				t.Fatal("unseen tile counted")
			}
		})
	})

	t.Run("牌の残り枚数", func(t *testing.T) {
		t.Run("どの牌も1種につき4枚から数え始めること", func(t *testing.T) {
			if supply("").Remaining(mt.T("1m")) != 4 {
				t.Fatal("not 4")
			}
		})
		t.Run("見えている枚数を4枚から引いた数が残り枚数となること", func(t *testing.T) {
			if supply("1m 1m").Remaining(mt.T("1m")) != 2 {
				t.Fatal("not 2")
			}
		})
		t.Run("赤5は通常の5と同じ種類として数えること", func(t *testing.T) {
			if supply("0m").Remaining(mt.T("5m")) != 3 || supply("5m").Remaining(mt.T("0m")) != 3 {
				t.Fatal("red five counted apart")
			}
		})
		t.Run("4枚すべてが見えている牌の残り枚数は0枚となること", func(t *testing.T) {
			if supply("1m 1m 1m 1m").Remaining(mt.T("1m")) != 0 {
				t.Fatal("not 0")
			}
		})
	})

	t.Run("牌が使えるかどうか", func(t *testing.T) {
		t.Run("残り枚数が1枚以上ある牌は、まだ使えること", func(t *testing.T) {
			if supply("1m 1m 1m").Remaining(mt.T("1m")) <= 0 {
				t.Fatal("not positive")
			}
		})
		t.Run("4枚すべてが見えている牌は、もう使えないこと", func(t *testing.T) {
			if supply("1m 1m 1m 1m").Remaining(mt.T("1m")) != 0 {
				t.Fatal("not 0")
			}
		})
		t.Run("自分で4枚持っている牌は、もう使えないこと", func(t *testing.T) {
			h := mt.Hand("1m 1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9p")
			if h.CanHold(mt.T("1m")) || !h.CanHold(mt.T("2m")) {
				t.Fatal("CanHold")
			}
		})
		t.Run("赤5と通常の5を合わせて4枚見えていれば、もう使えないこと", func(t *testing.T) {
			if supply("0m 5m 5m 5m").Remaining(mt.T("5m")) != 0 {
				t.Fatal("not 0")
			}
		})
	})

	t.Run("1種4枚を超える供給は存在しないこと", func(t *testing.T) {
		t.Run("同じ牌が5枚見えている供給は作れないこと", func(t *testing.T) {
			if _, err := tile.NewSupply(mt.Tiles("1m 1m 1m 1m 1m")); !errors.Is(err, tile.ErrOversupplied) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("同じ牌が5枚見えている局面は作れないこと", func(t *testing.T) {
			if _, err := mt.TryBuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 1m 1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m"}}); err == nil {
				t.Fatal("built")
			}
		})
	})
}

// ドラ。ドラ表示牌の「次の牌」がドラになる。翻を加算するが単独では役にならない。
func TestDora(t *testing.T) {
	t.Run("ドラ表示牌から次のドラを求める", func(t *testing.T) {
		t.Run("数牌は次の数へ進み、9の次は1へ戻ること", func(t *testing.T) {
			if mt.T("1m").Dora() != mt.T("2m") || mt.T("8s").Dora() != mt.T("9s") || mt.T("9p").Dora() != mt.T("1p") {
				t.Fatal("dora")
			}
		})
		t.Run("風牌は 東→南→西→北→東 と循環すること", func(t *testing.T) {
			var got []tile.Tile
			for _, l := range []string{"1z", "2z", "3z", "4z"} {
				got = append(got, mt.T(l).Dora())
			}
			if labels(got) != "2z 3z 4z 1z" {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("三元牌は 白→發→中→白 と循環すること", func(t *testing.T) {
			var got []tile.Tile
			for _, l := range []string{"5z", "6z", "7z"} {
				got = append(got, mt.T(l).Dora())
			}
			if labels(got) != "6z 7z 5z" {
				t.Fatalf("got %v", got)
			}
		})
	})

	t.Run("ドラの数え方", func(t *testing.T) {
		count := func(held, indicators string) int {
			return tile.DoraCount(mt.Tiles(held), mt.Tiles(indicators))
		}

		t.Run("表示牌の次の牌を持っている枚数だけ数えること", func(t *testing.T) {
			if count("2m 2m 3m 4m", "1m") != 2 {
				t.Fatal("count")
			}
		})
		t.Run("表示牌が複数あればそれぞれ数えること", func(t *testing.T) {
			if count("2m 3m 4m", "1m 2m") != 2 {
				t.Fatal("count")
			}
		})
		t.Run("ドラを1枚も持っていなければ0であること", func(t *testing.T) {
			if count("5p 6p 7p", "1m") != 0 {
				t.Fatal("count")
			}
		})
	})

	t.Run("赤ドラ", func(t *testing.T) {
		t.Run("赤5は表示牌に関わらず常にドラ1枚として数えること", func(t *testing.T) {
			if tile.DoraCount(mt.Tiles("0m 1p 2p"), mt.Tiles("9s")) != 1 {
				t.Fatal("count")
			}
		})
		t.Run("赤5がドラ表示牌のときは通常の5として扱うこと", func(t *testing.T) {
			if mt.T("0m").Dora() != mt.T("6m") {
				t.Fatal("dora")
			}
		})
		t.Run("赤5が表ドラでもあれば重ねて数えること", func(t *testing.T) {
			if tile.DoraCount(mt.Tiles("0m 5m"), mt.Tiles("4m")) != 3 {
				t.Fatal("count")
			}
		})
	})

	t.Run("加点", func(t *testing.T) {
		// 表示牌 4s のドラは 5s。親の一通は和了牌を含めて 5s を2枚使う。
		settle := func(t *testing.T, dora, uradora string, riichiFirst bool) *kyoku.Result {
			t.Helper()
			spec := mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "5s", Dora: dora, Uradora: uradora}
			if riichiFirst {
				spec.Draws = "9s 1z 1z 1z 5s"
				spec.Actions = []kyoku.Action{mt.RiichiAction(0, "9s"), mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "1z"), mt.DiscardAction(3, "1z")}
			}
			spec.Actions = append(spec.Actions, kyoku.NewTsumo(0))
			return resultOf(t, mt.BuildKyoku(spec))
		}

		t.Run("ドラは翻数に加算すること", func(t *testing.T) {
			if settle(t, "4s", "", false).Deltas()[0] <= settle(t, "7z", "", false).Deltas()[0] {
				t.Fatal("dora not counted")
			}
		})
		t.Run("リーチしていない和了では裏ドラを数えないこと", func(t *testing.T) {
			if settle(t, "7z", "4s", false).Deltas()[0] != settle(t, "7z", "6z", false).Deltas()[0] {
				t.Fatal("uradora counted")
			}
		})
		t.Run("リーチした和了では裏ドラも数えること", func(t *testing.T) {
			if settle(t, "7z", "4s", true).Deltas()[0] <= settle(t, "7z", "6z", true).Deltas()[0] {
				t.Fatal("uradora not counted")
			}
		})
	})
}
