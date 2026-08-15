package knowledge_test

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

func discards(seat int, labels ...string) []kyoku.Action {
	out := make([]kyoku.Action, 0, len(labels))
	for i, l := range labels {
		out = append(out, mt.DiscardAction((seat+i)%4, l))
	}
	return out
}

func inRiver(ds []kyoku.Discard) []tile.Tile {
	var out []tile.Tile
	for _, d := range ds {
		if d.IsInRiver() {
			out = append(out, d.Tile())
		}
	}
	return out
}

func discardTiles(ds []kyoku.Discard) []tile.Tile {
	var out []tile.Tile
	for _, d := range ds {
		out = append(out, d.Tile())
	}
	return out
}

func resultOf(t *testing.T, k *kyoku.Kyoku) *kyoku.Result {
	t.Helper()
	r, ok := k.Result()
	if !ok {
		t.Fatal("not finished")
	}
	return r
}

func winnerOf(t *testing.T, k *kyoku.Kyoku) int {
	t.Helper()
	w, ok := resultOf(t, k).Winner()
	if !ok {
		t.Fatal("no winner")
	}
	return w
}

// 局の進行。配牌から決着まで、1局を打ち切るあいだに卓で起きることの順序と規則。
//
// 局の真実は「開始局面」と「プレイヤーが下した選択の列」だけで、ある時点の場況は
// そこから畳んで導かれる。ここに書くのは、その畳み方が従う麻雀のルール。
func TestKyokuFlow(t *testing.T) {
	// 親が 2m を切った直後の卓。下家はチー、対面はポン、上家はロンができる。
	mixedClaims := func(claims ...kyoku.Action) *kyoku.Kyoku {
		return mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{
				0: "2m 9m 9m 1z 1z 2z 2z 3z 3z 4z 4z 5z 5z",
				1: "3m 4m 6m 7m 9m 2p 3p 5p 6p 8p 1s 3s 7s",
				2: "2m 2m 6m 7m 9m 2s 3s 5s 6s 8s 1p 3p 5p",
				3: "3m 4m 6m 7m 8m 2p 3p 4p 6p 7p 8p 5s 5s",
			},
			Draws: "7z", Actions: append([]kyoku.Action{mt.DiscardAction(0, "2m")}, claims...),
		})
	}
	// 親が切った 2m で、他の3人がそろってロンできる卓。
	tripleRon := func(claims ...kyoku.Action) *kyoku.Kyoku {
		return mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{
				0: "2m 1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z",
				1: "3m 4m 6m 7m 8m 2p 3p 4p 6p 7p 8p 5s 5s",
				2: "3m 4m 6m 7m 8m 2s 3s 4s 6s 7s 8s 5p 5p",
				3: "1m 3m 4m 5m 6m 7m 8m 9m 2s 3s 4s 9s 9s",
			},
			Draws: "7z", Actions: append([]kyoku.Action{mt.DiscardAction(0, "2m")}, claims...),
		})
	}
	ponSpec := func(actions ...kyoku.Action) mt.KyokuSpec {
		return mt.KyokuSpec{
			Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
			Draws: "5z", Actions: actions,
		}
	}

	t.Run("局の始まり", func(t *testing.T) {
		start := mt.BuildKyoku(mt.KyokuSpec{}).Kyokumen()

		t.Run("積んだ山を配れば局が始まること", func(t *testing.T) {
			w := kyoku.ShuffledWall(rand.New(rand.NewPCG(1, 1)), true)
			k, err := kyoku.Deal(kyoku.Setup{Wall: &w})
			if err != nil {
				t.Fatal(err)
			}
			if !k.Kyokumen().IsOpening() || k.IsFinished() {
				t.Fatal("not an opening")
			}
		})
		t.Run("4人に13枚ずつ配られること", func(t *testing.T) {
			for seat := 0; seat < 4; seat++ {
				if got := len(start.Seat(seat).Hand().ClosedTiles()); got != 13 {
					t.Errorf("seat %d: %d", seat, got)
				}
			}
		})
		t.Run("親が最初にツモること", func(t *testing.T) {
			seat, ok := start.DiscardingSeat()
			_, drawn := start.Drawn()
			if !ok || seat != 0 || !drawn {
				t.Fatalf("seat %d ok %v drawn %v", seat, ok, drawn)
			}
		})
		t.Run("ドラ表示牌が1枚だけ公開されていること", func(t *testing.T) {
			if len(start.DoraIndicators()) != 1 {
				t.Fatal("dora")
			}
		})
		t.Run("巡目が1であること", func(t *testing.T) {
			if start.Junme() != 1 {
				t.Fatal("junme")
			}
		})
		t.Run("河も副露も無いこと", func(t *testing.T) {
			if !start.IsOpening() {
				t.Fatal("not opening")
			}
		})
	})

	// 1人の手番は「ツモる → 何をするか選ぶ」で1組。打牌で手番が閉じ、他家の反応を待つ。
	t.Run("手番", func(t *testing.T) {
		t.Run("手番の席はツモってから打牌すること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s"}, Draws: "5z"}).Kyokumen()
			drawn, _ := k.Drawn()
			if drawn != tile.Haku || len(actionsOfKind(k, kyoku.ActionDiscard)) != 14 {
				t.Fatalf("drawn %v discards %d", drawn, len(actionsOfKind(k, kyoku.ActionDiscard)))
			}
		})
		t.Run("鳴ける席があれば、打牌したあと反応を待つこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 5z", 2: "5z 5z 1m 2m 3m 4m 6m 7m 8m 9m 1p 2p 3p"},
				Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")},
			}).Kyokumen()
			claimed, ok := k.ClaimedTile()
			from, _ := k.ClaimedFrom()
			if !ok || claimed != tile.Haku || from != 0 {
				t.Fatalf("claimed %v %v from %d", claimed, ok, from)
			}
		})
		// 反応できる席が1つも無い捨て牌は、誰の答えも要らない。
		t.Run("誰も鳴けない捨て牌なら、待たずに下家のツモへ進むこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{
					0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s",
					1: "1z 2z 3z 4z 5z 6z 7z 1m 4m 7m 1p 4p 7p",
					2: "1z 2z 3z 4z 5z 6z 7z 1m 4m 7m 1p 4p 7p",
					3: "1z 2z 3z 4z 5z 6z 7z 1m 4m 7m 1p 4p 7p",
				},
				Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")},
			}).Kyokumen()
			_, claimed := k.ClaimedTile()
			seat, _ := k.DiscardingSeat()
			if claimed || seat != 1 {
				t.Fatalf("claimed %v seat %d", claimed, seat)
			}
		})
		t.Run("手番の席は自分の捨て牌に反応できないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s"}, Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")}}).Kyokumen()
			for _, a := range k.LegalActions() {
				if a.Seat() == 0 {
					t.Fatalf("seat 0 offered %v", a)
				}
			}
		})
		t.Run("誰も反応しなければ下家がツモること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s"},
				Draws: "5z 1z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z"), mt.DiscardAction(1, "1z")},
			}).Kyokumen()
			ds := k.Seat(1).Discards()
			if ds[len(ds)-1].Tile() != tile.East {
				t.Fatal("seat 1 did not discard")
			}
		})
		t.Run("鳴くと手番が鳴いた席へ移り、飛ばされた席はその巡のツモを失うこと", func(t *testing.T) {
			k := mt.BuildKyoku(ponSpec(mt.DiscardAction(0, "2m"), mt.PonAction(2, "2m", "2m 2m"), kyoku.NewPass(1))).Kyokumen()
			seat, _ := k.DiscardingSeat()
			if seat != 2 || len(k.Seat(1).Discards()) != 0 {
				t.Fatalf("seat %d discards %v", seat, k.Seat(1).Discards())
			}
		})
		t.Run("鳴いた席はツモらずに打牌すること", func(t *testing.T) {
			k := mt.BuildKyoku(ponSpec(mt.DiscardAction(0, "2m"), mt.PonAction(2, "2m", "2m 2m"))).Kyokumen()
			if _, ok := k.Drawn(); ok {
				t.Fatal("drawn")
			}
		})
	})

	t.Run("巡目", func(t *testing.T) {
		t.Run("親に手番が戻るたびに巡目が1つ進むこと", func(t *testing.T) {
			oneAround := discards(0, "1z", "2z", "3z")
			if mt.BuildKyoku(mt.KyokuSpec{Draws: "1z 2z 3z 4z", Actions: oneAround}).Kyokumen().Junme() != 1 {
				t.Error("junme after three")
			}
			if mt.BuildKyoku(mt.KyokuSpec{Draws: "1z 2z 3z 4z", Actions: discards(0, "1z", "2z", "3z", "4z")}).Kyokumen().Junme() != 2 {
				t.Error("junme after four")
			}
		})
		t.Run("鳴きで親を飛び越したときも巡目が進むこと", func(t *testing.T) {
			// 席3の捨て牌を席1がポンすると、親(席0)のツモ番が飛ばされる。
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{3: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 1: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
				Draws: "1z 2z 3z",
				Actions: append(discards(0, "1z", "2z", "3z"),
					mt.DiscardAction(3, "2m"), mt.PonAction(1, "2m", "2m 2m"), kyoku.NewPass(0)),
			}).Kyokumen()
			if k.Junme() != 2 {
				t.Fatalf("junme %d", k.Junme())
			}
		})
	})

	// 鳴くかどうかは打ち手が決める。鳴かないことも1つの選択なので、鳴ける席は答えるまで
	// 待たれる。答えようのない席には聞かないので、卓は止まらない。
	t.Run("見送り", func(t *testing.T) {
		t.Run("鳴ける席には、鳴かない選択も出ること", func(t *testing.T) {
			expectKinds(t, kindsOf(mixedClaims().Kyokumen(), 1), kyoku.ActionChi, kyoku.ActionPass)
		})
		t.Run("鳴きようが無い席には何も聞かないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 5z", 2: "5z 5z 1m 2m 3m 4m 6m 7m 8m 9m 1p 2p 3p"},
				Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "5z")},
			})
			if !reflect.DeepEqual(k.AwaitingSeats(), []int{2}) {
				t.Fatalf("awaiting %v", k.AwaitingSeats())
			}
		})
		t.Run("鳴ける席が全員見送ると、鳴き待ちが閉じて下家のツモへ進むこと", func(t *testing.T) {
			k := mixedClaims(kyoku.NewPass(1), kyoku.NewPass(2), kyoku.NewPass(3)).Kyokumen()
			_, claimed := k.ClaimedTile()
			seat, _ := k.DiscardingSeat()
			if claimed || seat != 1 {
				t.Fatalf("claimed %v seat %d", claimed, seat)
			}
		})
		t.Run("一部の席が見送っただけでは、まだ他の席の答えを待つこと", func(t *testing.T) {
			k := mixedClaims(kyoku.NewPass(1))
			if _, claimed := k.Kyokumen().ClaimedTile(); !claimed || !reflect.DeepEqual(k.AwaitingSeats(), []int{2, 3}) {
				t.Fatalf("claimed %v awaiting %v", claimed, k.AwaitingSeats())
			}
		})
		t.Run("見送った席は同じ捨て牌に二度答えられないこと", func(t *testing.T) {
			if _, err := mixedClaims(kyoku.NewPass(1)).Take(kyoku.NewPass(1)); !errors.Is(err, kyoku.ErrIllegalAction) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("和了牌を見送ると、その捨て牌は取れないままフリテンになること", func(t *testing.T) {
			if !mixedClaims(kyoku.NewPass(1), kyoku.NewPass(2), kyoku.NewPass(3)).Kyokumen().Seat(3).IsFuriten() {
				t.Fatal("not furiten")
			}
		})
	})

	// 1つの捨て牌に重なった宣言は1つしか成立しない。和了は鳴きより強く、
	// 同じ牌を3枚使うポンは順子のチーより強い。
	t.Run("宣言の競合", func(t *testing.T) {
		t.Run("和了の宣言は鳴きより強いこと", func(t *testing.T) {
			k := mixedClaims(mt.ChiAction(1, "2m", "3m 4m"), mt.PonAction(2, "2m", "2m 2m"), kyoku.NewRon(3))
			if winnerOf(t, k) != 3 {
				t.Fatal("winner")
			}
		})
		t.Run("ポンはチーより強いこと", func(t *testing.T) {
			k := mixedClaims(mt.ChiAction(1, "2m", "3m 4m"), mt.PonAction(2, "2m", "2m 2m"), kyoku.NewPass(3)).Kyokumen()
			seat, _ := k.DiscardingSeat()
			if _, pending := k.PendingCall(1); seat != 2 || pending {
				t.Fatalf("seat %d pending %v", seat, pending)
			}
		})
		t.Run("同じ強さの宣言が重なると、切った席に近い席が通ること（頭ハネ）", func(t *testing.T) {
			if winnerOf(t, tripleRon(kyoku.NewRon(2), kyoku.NewRon(3), kyoku.NewPass(1))) != 2 {
				t.Fatal("winner")
			}
		})
	})

	// 同じ捨て牌に複数の席が反応できるとき、成立するのは1つだけ。
	t.Run("鳴きの優先順位", func(t *testing.T) {
		t.Run("ロンはポン・カン・チーより優先されること", func(t *testing.T) {
			if winnerOf(t, mixedClaims(mt.PonAction(2, "2m", "2m 2m"), kyoku.NewRon(3), kyoku.NewPass(1))) != 3 {
				t.Fatal("winner")
			}
		})
		t.Run("ポン・明槓はチーより優先されること", func(t *testing.T) {
			seat, _ := mixedClaims(mt.ChiAction(1, "2m", "3m 4m"), mt.PonAction(2, "2m", "2m 2m"), kyoku.NewPass(3)).Kyokumen().DiscardingSeat()
			if seat != 2 {
				t.Fatalf("seat %d", seat)
			}
		})
		t.Run("同じ牌に複数のロンが重なったとき、切った席から見て上家の1人だけが和了ること", func(t *testing.T) {
			if winnerOf(t, tripleRon(kyoku.NewRon(3), kyoku.NewRon(1), kyoku.NewPass(2))) != 1 {
				t.Fatal("winner")
			}
		})
		t.Run("優先されない宣言は成立せず、その席の手牌は動かないこと", func(t *testing.T) {
			k := mixedClaims(mt.ChiAction(1, "2m", "3m 4m"), mt.PonAction(2, "2m", "2m 2m"), kyoku.NewPass(3)).Kyokumen()
			closed := labels(k.Seat(1).Hand().ClosedTiles())
			if len(k.Seat(1).Hand().Melds()) != 0 || !containsLabel(closed, "3m") || !containsLabel(closed, "4m") {
				t.Fatalf("hand %v", k.Seat(1).Hand())
			}
		})
		// 先に出た宣言で確定してしまうなら、後から来る宣言で結果が変わることはない。
		t.Run("宣言が出そろうまでは、先に出た宣言でも確定しないこと", func(t *testing.T) {
			chi := mt.ChiAction(1, "2m", "3m 4m")
			if _, ok := mixedClaims(chi).Kyokumen().DiscardingSeat(); ok {
				t.Fatal("decided early")
			}
			seat, _ := mixedClaims(chi, mt.PonAction(2, "2m", "2m 2m"), kyoku.NewPass(3)).Kyokumen().DiscardingSeat()
			if seat != 2 {
				t.Fatalf("seat %d", seat)
			}
		})
		t.Run("先に出た弱い宣言は、後から来た強い宣言に覆されること", func(t *testing.T) {
			declared := mixedClaims(mt.ChiAction(1, "2m", "3m 4m"))
			k := mustKyoku(t)(mustKyoku(t)(declared.Take(kyoku.NewRon(3))).Take(kyoku.NewPass(2)))
			if winnerOf(t, k) != 3 {
				t.Fatal("winner")
			}
		})
	})

	// 鳴きは手の内を晒す。晒した瞬間から門前ではなくなり、リーチもできない。
	t.Run("鳴きの効果", func(t *testing.T) {
		afterPon := mt.BuildKyoku(ponSpec(mt.DiscardAction(0, "2m"), mt.PonAction(2, "2m", "2m 2m"), mt.DiscardAction(2, "9p"))).Kyokumen()

		t.Run("鳴いた牌は切った席の河から消え、鳴いた席の副露になること", func(t *testing.T) {
			if contains(inRiver(afterPon.Seat(0).Discards()), tile.M2) || !contains(discardTiles(afterPon.Seat(0).Discards()), tile.M2) {
				t.Fatal("river")
			}
			if len(afterPon.Seat(2).Hand().Melds()[0].Tiles()) != 3 {
				t.Fatal("meld")
			}
		})
		t.Run("ポン・チーは、鳴いた牌と手の内の2枚で面子が確定すること", func(t *testing.T) {
			m := afterPon.Seat(2).Hand().Melds()[0]
			if m.Kind().String() != "pon" || labels(m.Tiles()) != "2m 2m 2m" {
				t.Fatalf("meld %v", m)
			}
		})
		t.Run("鳴くと門前でなくなり、以後リーチを宣言できないこと", func(t *testing.T) {
			if afterPon.Seat(2).IsMenzen() {
				t.Fatal("menzen")
			}
		})
		t.Run("明槓は手の内の3枚と鳴いた牌で即座に確定し、嶺上牌をツモること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 2m 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
				Draws: "5z", Dora: "1z",
				Actions: []kyoku.Action{mt.DiscardAction(0, "2m"), mt.MinkanAction(2, "2m"), kyoku.NewPass(1)},
			}).Kyokumen()
			seat, _ := k.DiscardingSeat()
			if k.Seat(2).Hand().Melds()[0].Kind().String() != "minkan" || seat != 2 || !k.IsRinshanDraw() {
				t.Fatalf("meld %v seat %d rinshan %v", k.Seat(2).Hand().Melds(), seat, k.IsRinshanDraw())
			}
		})
		t.Run("鳴きが入ると、リーチしている席の一発が消えること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", 1: "9s 9s 1z 2z 3z 4p 5p 6p 7p 8p 1s 2s 3s"},
				Draws: "9s", Actions: []kyoku.Action{mt.RiichiAction(0, "9s"), mt.PonAction(1, "9s", "9s 9s")},
			}).Kyokumen()
			if !k.Seat(0).IsRiichi() || k.Seat(0).IsIppatsu() {
				t.Fatal("ippatsu")
			}
		})
		t.Run("暗槓でも一発が消えること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", 1: "5z 5z 5z 1z 2z 3z 4p 5p 6p 7p 8p 9p 1s"},
				Draws: "9s 5z", Actions: []kyoku.Action{mt.RiichiAction(0, "9s"), mt.AnkanAction(1, "5z")},
			}).Kyokumen()
			if k.Seat(0).IsIppatsu() {
				t.Fatal("ippatsu")
			}
		})
	})

	t.Run("槓", func(t *testing.T) {
		kanSpec := mt.KyokuSpec{Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"}, Draws: "5z"}
		afterAnkan := mt.BuildKyoku(mt.KyokuSpec{Hands: kanSpec.Hands, Draws: kanSpec.Draws, Actions: []kyoku.Action{mt.AnkanAction(0, "5z")}}).Kyokumen()

		t.Run("槓のたびにカンドラが1枚めくれること", func(t *testing.T) {
			if len(afterAnkan.DoraIndicators()) != 2 {
				t.Fatal("dora")
			}
		})
		t.Run("槓のあとは嶺上牌をツモり、そのまま打牌すること", func(t *testing.T) {
			seat, _ := afterAnkan.DiscardingSeat()
			if !afterAnkan.IsRinshanDraw() || seat != 0 {
				t.Fatal("rinshan")
			}
		})
		t.Run("嶺上牌を取ると海底からツモ山が1枚減ること", func(t *testing.T) {
			before := mt.BuildKyoku(kanSpec).Kyokumen()
			after := mt.BuildKyoku(mt.KyokuSpec{Hands: kanSpec.Hands, Draws: kanSpec.Draws, Actions: []kyoku.Action{mt.AnkanAction(0, "5z"), mt.DiscardAction(0, "1s")}}).Kyokumen()
			// 槓の巡のツモで1枚、嶺上への補充で1枚。
			if after.RemainingDraws() != before.RemainingDraws()-2 {
				t.Fatalf("%d -> %d", before.RemainingDraws(), after.RemainingDraws())
			}
		})
	})

	// 加槓は、既に晒してある刻子に4枚目を足す。晒す前の一瞬だけ、その牌を横取りできる。
	t.Run("槍槓", func(t *testing.T) {
		// 席0が 3m をポンし、一巡してから4枚目を引いて加槓する。席2は 12m の辺張で 3m を待つ
		// 一通のテンパイ。字牌は4枚とも加槓する側に集まってしまうので、槍槓は数牌でしか起きない。
		kakanKyoku := func(extra ...kyoku.Action) *kyoku.Kyoku {
			return mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "3m 3m 1p 2p 3p 4p 5p 6p 7p 8p 9p 5s 5s", 2: "1m 2m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s 9s"},
				Draws: "9s 3m 1z 2z 3z 3m",
				Actions: append([]kyoku.Action{
					mt.DiscardAction(0, "9s"), mt.DiscardAction(1, "3m"),
					mt.PonAction(0, "3m", "3m 3m"), mt.DiscardAction(0, "5s"),
					mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "2z"), mt.DiscardAction(3, "3z"),
					mt.KakanAction(0, "3m"),
				}, extra...),
			})
		}

		t.Run("加槓する牌でロンできる席があれば、加槓に反応できること", func(t *testing.T) {
			k := kakanKyoku().Kyokumen()
			claimed, ok := k.ClaimedTile()
			if !ok || claimed != tile.M3 {
				t.Fatalf("claimed %v %v", claimed, ok)
			}
			expectKinds(t, kindsOf(k, -1), kyoku.ActionRon, kyoku.ActionPass)
		})
		t.Run("槍槓が成立すると加槓は取り消され、加槓した席が放銃者になること", func(t *testing.T) {
			r := resultOf(t, kakanKyoku(kyoku.NewRon(2)))
			loser, _ := r.Loser()
			if w, _ := r.Winner(); w != 2 || loser != 0 || r.Deltas()[0] >= 0 {
				t.Fatalf("winner %d loser %d deltas %v", w, loser, r.Deltas())
			}
		})
		t.Run("誰も反応しなければ加槓が確定し、嶺上牌をツモること", func(t *testing.T) {
			k := kakanKyoku(mt.DiscardAction(0, "9p")).Kyokumen()
			m := k.Seat(0).Hand().Melds()[0]
			if m.Kind().String() != "minkan" || len(m.Tiles()) != 4 {
				t.Fatalf("meld %v", m)
			}
		})
		t.Run("暗槓には反応できないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"}, Draws: "5z", Actions: []kyoku.Action{mt.AnkanAction(0, "5z")}}).Kyokumen()
			if _, ok := k.ClaimedTile(); ok {
				t.Fatal("claimable")
			}
		})
	})

	// 局の途中で成立が確定し、その場で打ち切りになる流局。誰の手も完成していないので罰符は無い。
	t.Run("途中流局", func(t *testing.T) {
		t.Run("鳴きの入らない第一巡で、么九牌9種類以上を持つ席は九種九牌を宣言できること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"}, Draws: "1m"}).Kyokumen()
			if !hasKind(kindsOf(k, -1), kyoku.ActionKyushukyuhai) {
				t.Fatal("no kyushukyuhai")
			}
		})
		t.Run("么九牌が8種類以下なら九種九牌を宣言できないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 9m 1p 9p 1s 9s 1z 2z 3m 4m 5m 6m 7m"}, Draws: "8m"}).Kyokumen()
			if hasKind(kindsOf(k, -1), kyoku.ActionKyushukyuhai) {
				t.Fatal("kyushukyuhai offered")
			}
		})
		t.Run("第一巡で4人が同じ風牌を切ると四風連打で流れること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Draws: "1z 1z 1z 1z", Actions: discards(0, "1z", "1z", "1z", "1z")})
			if resultOf(t, k).Kind() != kyoku.ResultSuufonRenda {
				t.Fatal("kind")
			}
		})
		t.Run("同じ風牌でも4人揃わなければ流れないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Draws: "1z 1z 1z 2z", Actions: discards(0, "1z", "1z", "1z", "2z")})
			if k.IsFinished() {
				t.Fatal("finished")
			}
		})
		t.Run("4人ともリーチを宣言すると四家立直で流れること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{
					0: "1m 1m 2m 2m 3m 3m 4m 4m 6m 6m 7m 7m 8m",
					1: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z",
					2: "1p 1p 2p 2p 3p 3p 4p 4p 5p 5p 6p 6p 7p",
					3: "1s 1s 2s 2s 3s 3s 4s 4s 6s 6s 7s 7s 8s",
				},
				Draws:   "5m 9m 8p 9p",
				Actions: []kyoku.Action{mt.RiichiAction(0, "5m"), mt.RiichiAction(1, "9m"), mt.RiichiAction(2, "8p"), mt.RiichiAction(3, "9p")},
			})
			if resultOf(t, k).Kind() != kyoku.ResultSuuchaRiichi || k.Kyokumen().RiichiSticks() != 4 {
				t.Fatal("kind or sticks")
			}
		})
		t.Run("1つの捨て牌に3人がロンを宣言すると三家和になること", func(t *testing.T) {
			r := resultOf(t, tripleRon(kyoku.NewRon(1), kyoku.NewRon(2), kyoku.NewRon(3)))
			if r.Kind() != kyoku.ResultSanchaho || r.Deltas() != [4]int{} {
				t.Fatalf("kind %v deltas %v", r.Kind(), r.Deltas())
			}
		})
		t.Run("2人以上で合計4つの槓が入ると四槓散了になること", func(t *testing.T) {
			// 槓の直後は嶺上牌が手番に残るので、4枚目を嶺上に積めば1手番で槓を連ねられる。
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1p 2p 3p 4p 5p 6p 7p 1m 1m 1m 2m 2m 2m", 1: "1s 2s 3s 4s 6s 7s 8s 3m 3m 3m 4m 4m 4m"},
				Draws: "1m 3m", Rinshan: "2m 9p 4m 9s",
				Actions: []kyoku.Action{
					mt.AnkanAction(0, "1m"), mt.AnkanAction(0, "2m"), mt.DiscardAction(0, "9p"),
					mt.AnkanAction(1, "3m"), mt.AnkanAction(1, "4m"), mt.DiscardAction(1, "9s"),
				},
			}).Kyokumen()
			if d, ok := k.DrawKind(); !ok || d != kyoku.DrawSuukaikan {
				t.Fatalf("draw %v %v", d, ok)
			}
		})
		t.Run("1人で4つ槓している場合は四槓散了にならないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1p 1m 1m 1m 2m 2m 2m 3m 3m 3m 4m 4m 4m"},
				Draws: "1m", Rinshan: "2m 3m 4m 9p",
				Actions: []kyoku.Action{
					mt.AnkanAction(0, "1m"), mt.AnkanAction(0, "2m"), mt.AnkanAction(0, "3m"), mt.AnkanAction(0, "4m"), mt.DiscardAction(0, "9p"),
				},
			}).Kyokumen()
			if _, ok := k.DrawKind(); ok {
				t.Fatal("drawn")
			}
		})
		t.Run("途中流局では点棒が動かず、親が続いて本場が増えること", func(t *testing.T) {
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z"}, Draws: "1m", Actions: []kyoku.Action{kyoku.NewKyushukyuhai(0)}}))
			if r.Deltas() != [4]int{} || !r.DealerContinues() || r.NextHonba() != 1 {
				t.Fatalf("deltas %v continues %v honba %d", r.Deltas(), r.DealerContinues(), r.NextHonba())
			}
		})
	})

	t.Run("局の決着", func(t *testing.T) {
		t.Run("終わった局には、それ以上どの席も何も選べないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}})
			if len(k.Kyokumen().LegalActions()) != 0 || !k.IsFinished() {
				t.Fatal("not over")
			}
		})
	})

	// 配牌して、その局面で選べるものを選び続ければ、局は必ず決着に至る。
	// 途中で「誰も何も選べないのに終わってもいない」局面が現れたら、卓が止まっている。
	t.Run("1局を打ち切る", func(t *testing.T) {
		seededWall := func() kyoku.Wall { return kyoku.ShuffledWall(rand.New(rand.NewPCG(1, 1)), true) }
		playedOut := func(t *testing.T) *kyoku.Kyoku {
			t.Helper()
			w := seededWall()
			k, err := kyoku.Deal(kyoku.Setup{Wall: &w})
			if err != nil {
				t.Fatal(err)
			}
			for !k.IsFinished() {
				k = mustKyoku(t)(k.Take(k.Kyokumen().LegalActions()[0]))
			}
			return k
		}

		t.Run("配牌から選び続ければ決着に至ること", func(t *testing.T) {
			if _, ok := playedOut(t).Result(); !ok {
				t.Fatal("no result")
			}
		})
		t.Run("終局するまで、どの局面にも選べるアクションがあること", func(t *testing.T) {
			w := seededWall()
			replayed, _ := kyoku.Deal(kyoku.Setup{Wall: &w})
			for _, a := range playedOut(t).Actions() {
				found := false
				for _, legal := range replayed.Kyokumen().LegalActions() {
					if legal == a {
						found = true
					}
				}
				if !found {
					t.Fatalf("%v not legal", a)
				}
				replayed = mustKyoku(t)(replayed.Take(a))
			}
		})
		t.Run("積んだ山と打った選択の列だけで、同じ局が復元できること", func(t *testing.T) {
			played := playedOut(t)
			w := seededWall()
			restored, err := kyoku.Deal(kyoku.Setup{Wall: &w, Actions: played.Actions()})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(restored.Kyokumen(), played.Kyokumen()) || resultOf(t, restored).Kind() != resultOf(t, played).Kind() {
				t.Fatal("differs")
			}
		})
		t.Run("積んだ山の並びだけで、同じ配牌から打ち直せること", func(t *testing.T) {
			w := seededWall()
			copied := kyoku.MustWall(w.Tiles())
			a, _ := kyoku.Deal(kyoku.Setup{Wall: &copied})
			b, _ := kyoku.Deal(kyoku.Setup{Wall: &w})
			if !reflect.DeepEqual(a.Kyokumen(), b.Kyokumen()) {
				t.Fatal("differs")
			}
		})
	})
}

func containsLabel(joined, label string) bool {
	return has(strings.Fields(joined), label)
}
