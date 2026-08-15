package knowledge_test

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// 山。配牌を終えたあとに卓へ残る牌。ツモの供給源であり、尽きると局が終わる。
// ツモ山と王牌からなる。嶺上牌を取ると海底から王牌へ補充されるので、この2つは分けられない。
func TestWall(t *testing.T) {
	// 親が 5z の暗槓を打てる開始局面。
	kanReady := func() *kyoku.Kyokumen {
		return mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"}, Draws: "5z"}).Kyokumen()
	}
	afterKan := func(extra ...kyoku.Action) *kyoku.Kyokumen {
		return mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"}, Draws: "5z",
			Actions: append([]kyoku.Action{mt.AnkanAction(0, "5z")}, extra...),
		}).Kyokumen()
	}
	seeded := func() kyoku.Wall { return kyoku.ShuffledWall(rand.New(rand.NewPCG(1, 1)), true) }
	countKinds := func(tiles []tile.Tile, byLabel bool) map[string]int {
		counts := map[string]int{}
		for _, x := range tiles {
			if byLabel {
				counts[x.String()]++
			} else {
				counts[x.Kind().String()]++
			}
		}
		return counts
	}

	// 打ち始める前に、牌を混ぜて山に積む。並びが決まれば配牌もツモ順も決まるので、
	// 同じ並びを取っておけば同じ局を何度でも再現できる。
	t.Run("洗牌と配牌", func(t *testing.T) {
		wall := seeded()

		t.Run("山が牌136枚すべてでできていること", func(t *testing.T) {
			if len(wall.Tiles()) != 136 {
				t.Fatal("size")
			}
		})
		t.Run("同種が4枚ずつそろっていること", func(t *testing.T) {
			for k, c := range countKinds(wall.Tiles(), false) {
				if c != 4 {
					t.Errorf("%s: %d", k, c)
				}
			}
		})
		t.Run("赤5を採用すると各色の5が1枚だけ赤に置き換わること", func(t *testing.T) {
			counts := countKinds(wall.Tiles(), true)
			if counts["0m"] != 1 || counts["0p"] != 1 || counts["0s"] != 1 || counts["5m"] != 3 || counts["5p"] != 3 || counts["5s"] != 3 {
				t.Fatalf("counts %v", counts)
			}
		})
		t.Run("赤5を採用しなければ通常の5が4枚ずつになること", func(t *testing.T) {
			counts := countKinds(kyoku.ShuffledWall(rand.New(rand.NewPCG(1, 1)), false).Tiles(), true)
			if counts["5m"] != 4 || counts["5p"] != 4 || counts["5s"] != 4 {
				t.Fatalf("counts %v", counts)
			}
		})
		t.Run("配牌は親から順に4枚ずつ3周し、最後に1枚ずつ取ること", func(t *testing.T) {
			tiles := wall.Tiles()
			want := append(append(append(append([]tile.Tile(nil), tiles[0:4]...), tiles[16:20]...), tiles[32:36]...), tiles[48])
			if !reflect.DeepEqual(wall.Hands()[0], want) {
				t.Fatalf("got %v", wall.Hands()[0])
			}
		})
		t.Run("4人が13枚ずつ取ること", func(t *testing.T) {
			for _, h := range wall.Hands() {
				if len(h) != 13 {
					t.Fatal("hand size")
				}
			}
		})
		t.Run("配ったあとの山が、ツモ山70枚と王牌14枚に分かれること", func(t *testing.T) {
			if len(wall.DrawTiles()) != 70 || len(wall.DeadTiles()) != 14 {
				t.Fatal("split")
			}
		})
		t.Run("同じ並びの山からは同じ局が始まること", func(t *testing.T) {
			restored := kyoku.MustWall(wall.Tiles())
			a := mustKyoku(t)(kyoku.Deal(kyoku.Setup{Wall: &restored}))
			b := mustKyoku(t)(kyoku.Deal(kyoku.Setup{Wall: &wall}))
			if !reflect.DeepEqual(a.Kyokumen(), b.Kyokumen()) {
				t.Fatal("differs")
			}
		})
	})

	t.Run("山の構成", func(t *testing.T) {
		t.Run("配牌を終えたあと、70枚がツモ牌になること", func(t *testing.T) {
			if mt.BuildKyoku(mt.KyokuSpec{}).Kyokumen().RemainingDraws() != 70 {
				t.Fatal("draws")
			}
		})
		t.Run("残りの14枚が王牌になること", func(t *testing.T) {
			wall := seeded()
			if len(wall.Tiles())-4*13-len(wall.DrawTiles()) != 14 {
				t.Fatal("dead")
			}
		})
	})

	t.Run("ツモ", func(t *testing.T) {
		t.Run("山の先頭から1枚ずつツモること", func(t *testing.T) {
			start := mt.BuildKyoku(mt.KyokuSpec{Draws: "5z 6z"}).Kyokumen()
			if d, _ := start.Drawn(); d != tile.Haku {
				t.Fatalf("first draw %v", d)
			}
			next := mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{Draws: "5z 6z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")}})).Kyokumen()
			if d, _ := next.Drawn(); d != tile.Hatsu {
				t.Fatalf("second draw %v", d)
			}
		})
		t.Run("ツモるたびに残り枚数が1枚減ること", func(t *testing.T) {
			start := mt.BuildKyoku(mt.KyokuSpec{Draws: "5z"}).Kyokumen()
			after := mt.BuildKyoku(mt.KyokuSpec{Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")}}).Kyokumen()
			if after.RemainingDraws() != start.RemainingDraws()-1 {
				t.Fatal("remaining")
			}
		})
		t.Run("残り0枚になると、もう誰もツモれず局が流れること", func(t *testing.T) {
			exhausted := mt.BuildKyoku(mt.KyokuSpec{Wall: mt.WallOf(0)})
			_, drawn := exhausted.Kyokumen().Drawn()
			if exhausted.Kyokumen().RemainingDraws() != 0 || drawn || resultOf(t, exhausted).Kind() != kyoku.ResultRyukyoku {
				t.Fatal("exhaust")
			}
		})
	})

	// 嶺上牌を取ると王牌が1枚欠けるので、海底から1枚補充される。
	// 結果として王牌は常に14枚のまま、ツモ山が1枚減る。
	t.Run("嶺上牌", func(t *testing.T) {
		t.Run("槓のときに王牌から1枚ツモれること", func(t *testing.T) {
			k := afterKan()
			if _, ok := k.Drawn(); !k.IsRinshanDraw() || !ok {
				t.Fatal("rinshan")
			}
		})
		t.Run("嶺上牌をツモるとツモ山が1枚減ること", func(t *testing.T) {
			// 槓の巡のツモで1枚、嶺上への補充で1枚。
			if afterKan(mt.DiscardAction(0, "1s")).RemainingDraws() != kanReady().RemainingDraws()-2 {
				t.Fatal("remaining")
			}
		})
		t.Run("嶺上牌をツモっても王牌は14枚のままであること", func(t *testing.T) {
			k := afterKan(mt.DiscardAction(0, "1s"))
			if len(k.DoraIndicators())+len(k.UradoraIndicators()) != 4 {
				t.Fatal("dead wall")
			}
		})
		t.Run("一度取った嶺上牌は二度と出てこないこと", func(t *testing.T) {
			hands := map[int]string{0: "5z 5z 5z 6z 6z 6z 1m 2m 3m 4m 5m 6m 1p"}
			one := mt.BuildKyoku(mt.KyokuSpec{Hands: hands, Draws: "5z", Rinshan: "6z 9s", Actions: []kyoku.Action{mt.AnkanAction(0, "5z")}}).Kyokumen()
			two := mt.BuildKyoku(mt.KyokuSpec{Hands: hands, Draws: "5z", Rinshan: "6z 9s", Actions: []kyoku.Action{mt.AnkanAction(0, "5z"), mt.AnkanAction(0, "6z")}}).Kyokumen()
			d1, _ := one.Drawn()
			d2, _ := two.Drawn()
			if d1 != tile.Hatsu || d2 != tile.S9 {
				t.Fatalf("%v %v", d1, d2)
			}
		})
		t.Run("補充された牌はドラ表示牌の位置をずらさないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"},
				Draws: "5z", Dora: "4z", Uradora: "3z", Rinshan: "1p",
				Actions: []kyoku.Action{mt.AnkanAction(0, "5z"), mt.DiscardAction(0, "1p")},
			}).Kyokumen()
			if k.DoraIndicators()[0] != tile.North || k.UradoraIndicators()[0] != tile.West {
				t.Fatal("indicators moved")
			}
		})
	})

	t.Run("ドラ表示牌", func(t *testing.T) {
		t.Run("局の開始時に1枚だけ公開されていること", func(t *testing.T) {
			if len(mt.BuildKyoku(mt.KyokuSpec{}).Kyokumen().DoraIndicators()) != 1 {
				t.Fatal("dora")
			}
		})
		t.Run("槓のたびに1枚ずつ公開されること", func(t *testing.T) {
			if len(afterKan().DoraIndicators()) != 2 {
				t.Fatal("dora")
			}
		})
		t.Run("槓は4つまでなので、公開は5枚で打ち止めであること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1p 1m 1m 1m 2m 2m 2m 3m 3m 3m 4m 4m 4m"},
				Draws: "1m", Rinshan: "2m 3m 4m 9p",
				Actions: []kyoku.Action{mt.AnkanAction(0, "1m"), mt.AnkanAction(0, "2m"), mt.AnkanAction(0, "3m"), mt.AnkanAction(0, "4m")},
			}).Kyokumen()
			if len(k.DoraIndicators()) != 5 {
				t.Fatal("dora")
			}
		})
		t.Run("裏ドラ表示牌は公開されている表ドラと同じ枚数あること", func(t *testing.T) {
			k := afterKan()
			if len(k.UradoraIndicators()) != len(k.DoraIndicators()) {
				t.Fatal("uradora")
			}
		})
	})

	// 牌が増えたり消えたりしていないことは、山と手牌を合わせて数えれば分かる。
	t.Run("牌の総数が変わらないこと", func(t *testing.T) {
		counted := func(k *kyoku.Kyokumen) int {
			total := k.RemainingDraws() + 14
			for seat := 0; seat < 4; seat++ {
				s := k.Seat(seat)
				total += len(s.Hand().ClosedTiles()) + len(s.Discards())
				for _, m := range s.Hand().Melds() {
					total += len(m.Tiles())
				}
			}
			return total
		}

		t.Run("ツモっても、山と手牌を合わせた総数が変わらないこと", func(t *testing.T) {
			if counted(mt.BuildKyoku(mt.KyokuSpec{Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")}}).Kyokumen()) != 136 {
				t.Fatal("count")
			}
		})
		t.Run("嶺上牌をツモっても、山と手牌を合わせた総数が変わらないこと", func(t *testing.T) {
			if counted(afterKan(mt.DiscardAction(0, "1s"))) != 136 {
				t.Fatal("count")
			}
		})
	})
}
