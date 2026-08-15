package knowledge_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

// 半荘。局の決着が次の局を生み、その連鎖は規定局数の打ち切りか飛びで途切れる。
// 局の中の出来事は局の知識であり、ここにあるのは局と局の境界の知識だけ。
func TestHanchan(t *testing.T) {
	// 次に打つ卓の場況。前局があればその決着から導き、無ければ開局の場況。
	// 手積みの卓を組むための下敷きとして、局に配らせて読む。
	atAfter := func(t *testing.T, k *kyoku.Kyoku) *kyoku.Kyokumen {
		t.Helper()
		if k == nil {
			return mustKyoku(t)(kyoku.Deal(kyoku.Setup{})).Kyokumen()
		}
		return mustKyoku(t)(k.DealNext()).Kyokumen()
	}
	// 場況どおりの卓を手積みで組む。
	kyokuAt := func(at *kyoku.Kyokumen, spec mt.KyokuSpec) *kyoku.Kyoku {
		spec.Dealer, spec.RoundWind, spec.KyokuNumber = at.DealerSeat(), at.RoundWind(), at.KyokuNumber()
		spec.Honba, spec.RiichiSticks = at.Honba(), at.RiichiSticks()
		if spec.Scores == nil {
			spec.Scores = map[int]int{}
			for seat := 0; seat < 4; seat++ {
				spec.Scores[seat] = at.Score(seat)
			}
		}
		return mt.BuildKyoku(spec)
	}
	// 場況の卓を、親の下家のツモ和了で打ち終える。
	childWins := func(at *kyoku.Kyokumen) *kyoku.Kyoku {
		winner := (at.DealerSeat() + 1) % 4
		return kyokuAt(at, mt.KyokuSpec{
			Hands: map[int]string{winner: ittsuHand}, Draws: "5z 5s",
			Actions: []kyoku.Action{mt.DiscardAction(at.DealerSeat(), "5z"), kyoku.NewTsumo(winner)},
		})
	}
	// 場況の卓を、親だけテンパイの荒牌平局で打ち終える(連荘)。
	dealerHolds := func(at *kyoku.Kyokumen) *kyoku.Kyoku {
		hands := allSeats(notenHand)
		hands[at.DealerSeat()] = ittsuHand
		return kyokuAt(at, mt.KyokuSpec{Hands: hands, Wall: mt.WallOf(0)})
	}
	// 場況の卓を、全員ノーテンの荒牌平局で打ち終える(親流れ・本場は増える)。
	allFold := func(at *kyoku.Kyokumen) *kyoku.Kyoku {
		return kyokuAt(at, mt.KyokuSpec{Hands: allSeats(notenHand), Wall: mt.WallOf(0)})
	}
	// 子の和了で親を流しながら東場を打ち終えた、東4局の決着済みの卓。
	throughEast := func(t *testing.T) *kyoku.Kyoku {
		t.Helper()
		k := childWins(atAfter(t, nil))
		for i := 0; i < 3; i++ {
			k = childWins(atAfter(t, k))
		}
		return k
	}
	// さらに南場も打ち終えた、南4局の決着済みの卓。
	throughSouth := func(t *testing.T) *kyoku.Kyoku {
		t.Helper()
		k := throughEast(t)
		for i := 0; i < 4; i++ {
			k = childWins(atAfter(t, k))
		}
		return k
	}

	t.Run("対局の始まり", func(t *testing.T) {
		t.Run("取り決めだけで開局でき、最初の局は東1局0本場、供託なし、配給原点の持ち点であること", func(t *testing.T) {
			opening := mustKyoku(t)(kyoku.Deal(kyoku.Setup{RuleSet: ruleset.Default()})).Kyokumen()
			if opening.RoundWind() != tile.EastWind || opening.KyokuNumber() != 1 || opening.Honba() != 0 || opening.RiichiSticks() != 0 {
				t.Fatal("opening")
			}
			for seat := 0; seat < 4; seat++ {
				if opening.Score(seat) != 25000 {
					t.Fatalf("seat %d: %d", seat, opening.Score(seat))
				}
			}
		})
		t.Run("起家が最初の親であること", func(t *testing.T) {
			if mustKyoku(t)(kyoku.Deal(kyoku.Setup{})).Kyokumen().DealerSeat() != 0 {
				t.Fatal("dealer")
			}
		})
		t.Run("山を渡さなければ、局が自分で山を積んで開局すること", func(t *testing.T) {
			if len(mustKyoku(t)(kyoku.Deal(kyoku.Setup{})).Wall().Tiles()) != 136 {
				t.Fatal("wall")
			}
		})
	})

	t.Run("局の進行", func(t *testing.T) {
		t.Run("子が和了ると親が流れ、次の局へ進み、親は下家へ移ること", func(t *testing.T) {
			after := atAfter(t, childWins(atAfter(t, nil)))
			if after.KyokuNumber() != 2 || after.DealerSeat() != 1 || after.Honba() != 0 {
				t.Fatalf("kyoku %d dealer %d honba %d", after.KyokuNumber(), after.DealerSeat(), after.Honba())
			}
		})
		t.Run("親がテンパイのまま流局すると連荘し、局は据え置きで本場が重なること", func(t *testing.T) {
			after := atAfter(t, dealerHolds(atAfter(t, nil)))
			if after.KyokuNumber() != 1 || after.DealerSeat() != 0 || after.Honba() != 1 {
				t.Fatalf("kyoku %d dealer %d honba %d", after.KyokuNumber(), after.DealerSeat(), after.Honba())
			}
		})
		t.Run("流局で親が流れたときは、局が進んでも本場は増えること", func(t *testing.T) {
			after := atAfter(t, allFold(atAfter(t, nil)))
			if after.KyokuNumber() != 2 || after.Honba() != 1 {
				t.Fatalf("kyoku %d honba %d", after.KyokuNumber(), after.Honba())
			}
		})
		t.Run("動いた持ち点が次の局へそのまま持ち越されること", func(t *testing.T) {
			k := dealerHolds(atAfter(t, nil))
			after := atAfter(t, k)
			scores := resultOf(t, k).Scores()
			for seat := 0; seat < 4; seat++ {
				if after.Score(seat) != scores[seat] {
					t.Fatalf("seat %d: %d vs %d", seat, after.Score(seat), scores[seat])
				}
			}
		})
		t.Run("流局で残ったリーチ棒は、供託として次の局へ持ち越されること", func(t *testing.T) {
			at := atAfter(t, nil)
			dealer := at.DealerSeat()
			hands := allSeats(notenHand)
			hands[dealer] = ittsuHand
			// 残りツモが4枚を切るとリーチできないので、山を6枚残して宣言後に尽きさせる。
			riichiDraw := kyokuAt(at, mt.KyokuSpec{
				Hands: hands, Draws: "8p 5z 5z 5z 6z 6z", Wall: mt.WallOf(6),
				Actions: []kyoku.Action{
					mt.RiichiAction(dealer, "8p"),
					mt.DiscardAction((dealer+1)%4, "5z"), mt.DiscardAction((dealer+2)%4, "5z"), mt.DiscardAction((dealer+3)%4, "5z"),
					mt.DiscardAction(dealer, "6z"), mt.DiscardAction((dealer+1)%4, "6z"),
				},
			})
			if atAfter(t, riichiDraw).RiichiSticks() != 1 {
				t.Fatal("stick not carried")
			}
		})
		t.Run("東4局で親が流れると南入すること", func(t *testing.T) {
			after := atAfter(t, throughEast(t))
			if after.RoundWind() != tile.SouthWind || after.KyokuNumber() != 1 || after.DealerSeat() != 0 {
				t.Fatalf("wind %v kyoku %d dealer %d", after.RoundWind(), after.KyokuNumber(), after.DealerSeat())
			}
		})
	})

	t.Run("対局の終わり", func(t *testing.T) {
		t.Run("南4局で親が流れると、規定の局数を打ち切って半荘が終わり、次の局は無いこと", func(t *testing.T) {
			k := throughSouth(t)
			if !k.IsLast() {
				t.Fatal("not last")
			}
			if _, err := k.DealNext(); !errors.Is(err, kyoku.ErrHanchanOver) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("オーラスでも親が続くかぎり半荘は終わらないこと", func(t *testing.T) {
			south3 := throughEast(t)
			for i := 0; i < 3; i++ {
				south3 = childWins(atAfter(t, south3))
			}
			k := dealerHolds(atAfter(t, south3))
			if !k.IsAllLast() || k.IsLast() || atAfter(t, k).KyokuNumber() != 4 {
				t.Fatal("all last")
			}
		})
		t.Run("誰かの持ち点が負になると、オーラスでなくても半荘が終わること（飛び）", func(t *testing.T) {
			busted := kyokuAt(atAfter(t, nil), mt.KyokuSpec{
				Hands: map[int]string{1: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"}, Draws: "9m",
				Actions: []kyoku.Action{mt.DiscardAction(0, "9m"), kyoku.NewRon(1)},
			})
			if resultOf(t, busted).Scores()[0] >= 0 || busted.IsAllLast() || !busted.IsLast() {
				t.Fatal("bust")
			}
		})
		t.Run("持ち点0はまだ飛んでいないこと", func(t *testing.T) {
			// 席1は 2600点 持ちで、一通の 2600点 に放銃してちょうど0点になる。
			// ドラが手に乗ると点が動くので、表示牌は誰の手にも効かない字牌に固定する。
			toZero := kyokuAt(atAfter(t, nil), mt.KyokuSpec{
				Hands: map[int]string{2: ittsuHand}, Draws: "5z 5s", Dora: "1z",
				Scores:  map[int]int{0: 25000, 1: 2600, 2: 25000, 3: 47400},
				Actions: []kyoku.Action{mt.DiscardAction(0, "5z"), mt.DiscardAction(1, "5s"), kyoku.NewRon(2)},
			})
			if resultOf(t, toZero).Scores()[1] != 0 || toZero.IsLast() {
				t.Fatal("zero")
			}
		})
		t.Run("オーラスでトップ目の親は、和了して対局を終わらせられること（アガリやめ）", func(t *testing.T) { t.Skip("未実装のドメイン知識") })
		t.Run("南4局を終えて誰も返し点に届いていなければ、西入して続くこと", func(t *testing.T) { t.Skip("未実装のドメイン知識") })
	})

	t.Run("対局の決着", func(t *testing.T) {
		t.Run("順位は持ち点で決まり、同点は起家に近い席が上であること", func(t *testing.T) { t.Skip("未実装のドメイン知識") })
		t.Run("終局時に残った供託はトップが総取りすること", func(t *testing.T) { t.Skip("未実装のドメイン知識") })
		t.Run("順位に応じたウマとオカで、素点が最終ポイントへ精算されること", func(t *testing.T) { t.Skip("未実装のドメイン知識") })
	})
}
