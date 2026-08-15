package knowledge_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
)

// 局面で選べるアクション。
//
// 何が選べるかは「局面が何を待っているか」と各席の手牌の両方で決まる。
// ここに挙がらないアクションは局面に適用できないので、選択肢の提示と選択の検証を兼ねる。
func TestLegalActions(t *testing.T) {
	// リーチを宣言し、一巡回って自分のツモ番に戻ってきた局面。
	riichiTurn := func() *kyoku.Kyokumen {
		return mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"},
			Draws: "9s 1z 1z 1z 2p",
			Actions: []kyoku.Action{
				mt.RiichiAction(0, "9s"), mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "1z"), mt.DiscardAction(3, "1z"),
			},
		})).Kyokumen()
	}
	// 席1が席0の切った牌をポンし、一巡回って自分のツモ番に戻ってきた局面。
	// hand はポンの後に残る手の内で、ポンで晒す2枚と巡を回す打牌の 1z を足して配る。
	ponTurn := func(hand, draw, called string) *kyoku.Kyokumen {
		return mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{1: hand + " " + called + " " + called + " 1z"},
			Draws: called + " 1z 1z 1z " + draw,
			Actions: []kyoku.Action{
				mt.DiscardAction(0, called), mt.PonAction(1, called, called+" "+called),
				mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "1z"), mt.DiscardAction(3, "1z"), mt.DiscardAction(0, "1z"),
			},
		}).Kyokumen()
	}
	start := func(hand, draws string) *kyoku.Kyokumen {
		return mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: hand}, Draws: draws}).Kyokumen()
	}

	t.Run("自分の手番", func(t *testing.T) {
		t.Run("打牌", func(t *testing.T) {
			t.Run("手の内の牌とツモ牌はどれでも切れること", func(t *testing.T) {
				got := discardLabels(start("1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s", "5z"), 0)
				if len(got) != 14 || !has(got, "5z") || !has(got, "1m") || !has(got, "9s") {
					t.Fatalf("got %v", got)
				}
			})
		})

		t.Run("ツモ和了", func(t *testing.T) {
			t.Run("手牌にツモ牌を加えて和了形になり、役が付けばツモ和了を選べること", func(t *testing.T) {
				if !hasKind(kindsOf(start("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s"), -1), kyoku.ActionTsumo) {
					t.Fatal("no tsumo")
				}
			})
			t.Run("和了形でも役が付かなければツモ和了を選べないこと", func(t *testing.T) {
				// 鳴いた手は門前清自摸和が付かないので、役が無ければ和了れない。
				if hasKind(kindsOf(ponTurn("3m 4m 5m 6m 7m 8m 1s 2s 3s 9s", "9s", "2p"), -1), kyoku.ActionTsumo) {
					t.Fatal("tsumo offered")
				}
			})
			t.Run("ツモ山最後の1枚を引いた和了は海底摸月になること", func(t *testing.T) {
				// 70枚のツモは親から順に4人で割り切れないので、海底は親の下家に落ちる。
				k := mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
					Hands: map[int]string{1: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"},
					Draws: "1z 5s", Wall: mt.WallOf(2), Actions: []kyoku.Action{mt.DiscardAction(0, "1z")},
				})).Kyokumen()
				if !k.IsHaiteiDraw() || !hasKind(kindsOf(k, -1), kyoku.ActionTsumo) {
					t.Fatalf("haitei %v kinds %v", k.IsHaiteiDraw(), kindsOf(k, -1))
				}
			})
			t.Run("槓の直後に嶺上牌で和了れば嶺上開花になること", func(t *testing.T) {
				k := mt.BuildKyoku(mt.KyokuSpec{
					Hands: map[int]string{0: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p"}, Draws: "5z", Rinshan: "1p",
					Actions: []kyoku.Action{mt.AnkanAction(0, "5z")},
				}).Kyokumen()
				if !k.IsRinshanDraw() || !hasKind(kindsOf(k, -1), kyoku.ActionTsumo) {
					t.Fatalf("rinshan %v kinds %v", k.IsRinshanDraw(), kindsOf(k, -1))
				}
			})
		})

		t.Run("暗槓", func(t *testing.T) {
			t.Run("手の内とツモ牌で同種4枚が揃えば暗槓を選べること", func(t *testing.T) {
				if !hasKind(kindsOf(start("5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s", "5z"), -1), kyoku.ActionAnkan) {
					t.Fatal("no ankan")
				}
			})
			t.Run("海底牌を引いた巡は、王牌へ回す牌が無いので槓を選べないこと", func(t *testing.T) {
				k := mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
					Hands: map[int]string{1: "5z 5z 5z 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s"},
					Draws: "1z 5z", Wall: mt.WallOf(2), Actions: []kyoku.Action{mt.DiscardAction(0, "1z")},
				})).Kyokumen()
				if hasKind(kindsOf(k, -1), kyoku.ActionAnkan) {
					t.Fatal("ankan offered")
				}
			})
		})

		t.Run("加槓", func(t *testing.T) {
			t.Run("既にポンしている牌の4枚目を手の内かツモ牌に持てば加槓を選べること", func(t *testing.T) {
				if !hasKind(kindsOf(ponTurn("3m 4m 5m 6m 7m 8m 1s 2s 3s 9s", "2p", "2p"), -1), kyoku.ActionKakan) {
					t.Fatal("no kakan")
				}
			})
		})

		t.Run("九種九牌", func(t *testing.T) {
			t.Run("鳴きの入らない第一巡で么九牌が9種類以上あれば九種九牌で流せること", func(t *testing.T) {
				if !hasKind(kindsOf(start("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", "1m"), -1), kyoku.ActionKyushukyuhai) {
					t.Fatal("no kyushukyuhai")
				}
			})
			t.Run("么九牌が8種類以下なら九種九牌を選べないこと", func(t *testing.T) {
				if hasKind(kindsOf(start("1m 9m 1p 9p 1s 9s 1z 2z 3m 4m 5m 6m 7m", "8m"), -1), kyoku.ActionKyushukyuhai) {
					t.Fatal("kyushukyuhai offered")
				}
			})
			t.Run("第一巡でも盤面に副露が入っていれば九種九牌を選べないこと", func(t *testing.T) {
				// 席3を親にし、席0の第一ツモが来る前に席1のポンを成立させる。
				k := mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
					Hands: map[int]string{
						0: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z",
						1: "1s 1s 1z 2m 2m 3m 3m 4m 4m 2p 2p 3p 3p",
					},
					Dealer: 3, Draws: "1s 2z 2z 1m",
					Actions: []kyoku.Action{
						mt.DiscardAction(3, "1s"), mt.PonAction(1, "1s", "1s 1s"),
						mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "2z"), mt.DiscardAction(3, "2z"),
					},
				})).Kyokumen()
				if hasKind(kindsOf(k, -1), kyoku.ActionKyushukyuhai) {
					t.Fatal("kyushukyuhai offered")
				}
			})
		})
	})

	// リーチは門前かつテンパイが前提で、宣言には点棒と、あと1巡回るだけのツモ山が要る。
	t.Run("リーチの宣言", func(t *testing.T) {
		// 9s を切れば 5s 待ち、5s を切れば 9s 待ちで、どちらもテンパイを保てる。
		declarable := func(scores map[int]int) []kyoku.Action {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "9s", Scores: scores,
			}).Kyokumen()
			return actionsOfKind(k, kyoku.ActionRiichi)
		}

		t.Run("門前でテンパイを保てる牌を切ってリーチを宣言できること", func(t *testing.T) {
			var got []string
			for _, a := range declarable(nil) {
				got = append(got, a.Tiles()[0].String())
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, []string{"5s", "9s"}) {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("手の内に副露があれば門前でないので宣言できないこと", func(t *testing.T) {
			if hasKind(kindsOf(ponTurn("1m 2m 3m 4m 5m 6m 7m 8m 9m 5s", "9s", "2p"), -1), kyoku.ActionRiichi) {
				t.Fatal("riichi offered")
			}
		})
		t.Run("テンパイを保てる牌が無ければ宣言できないこと", func(t *testing.T) {
			if hasKind(kindsOf(start("1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 2z 3z 4z", "5z"), -1), kyoku.ActionRiichi) {
				t.Fatal("riichi offered")
			}
		})
		t.Run("持ち点が1000点に満たなければ宣言できないこと", func(t *testing.T) {
			if len(declarable(map[int]int{0: 900})) != 0 || len(declarable(map[int]int{0: 1000})) == 0 {
				t.Fatal("score threshold")
			}
		})
		t.Run("ツモ山の残りが1巡（席数）に満たなければ宣言できないこと", func(t *testing.T) {
			tenpai := "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"
			k := mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{2: tenpai, 3: tenpai}, Draws: "1z 2z 3z", Wall: mt.WallOf(6),
				Actions: []kyoku.Action{mt.DiscardAction(0, "1z"), mt.DiscardAction(1, "2z")},
			}))
			if !hasKind(kindsOf(k.Kyokumen(), 2), kyoku.ActionRiichi) {
				t.Fatal("seat 2 cannot riichi")
			}
			after := mt.AfterOthersPass(mustKyoku(t)(k.Take(mt.DiscardAction(2, "3z"))))
			if hasKind(kindsOf(after.Kyokumen(), 3), kyoku.ActionRiichi) {
				t.Fatal("seat 3 can riichi")
			}
		})
		t.Run("既にリーチしていれば再び宣言できないこと", func(t *testing.T) {
			if hasKind(kindsOf(riichiTurn(), -1), kyoku.ActionRiichi) {
				t.Fatal("riichi offered")
			}
		})
	})

	// 鳴いてできた面子と同じ形に持ち替える打牌は、場を進めないので禁じられている。
	t.Run("喰い替えの禁止", func(t *testing.T) {
		// 席1が席0の切った牌を鳴き、その場で打牌する。
		afterCall := func(hand string, call kyoku.Action) *kyoku.Kyokumen {
			return mt.AfterOthersPass(mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 1: hand},
				Draws: "3m", Actions: []kyoku.Action{mt.DiscardAction(0, "3m"), call},
			})).Kyokumen()
		}

		t.Run("鳴いた牌と同じ牌には持ち替えられないこと", func(t *testing.T) {
			if has(discardLabels(afterCall("3m 4m 5m 7m 8m 9m 1p 2p 3p 4p 5p 6p 9p", mt.ChiAction(1, "3m", "4m 5m")), 1), "3m") {
				t.Fatal("3m allowed")
			}
		})
		t.Run("ポンでも鳴いた牌と同じ牌には持ち替えられないこと", func(t *testing.T) {
			if has(discardLabels(afterCall("3m 3m 3m 7m 8m 9m 1p 2p 3p 4p 5p 6p 9p", mt.PonAction(1, "3m", "3m 3m")), 1), "3m") {
				t.Fatal("3m allowed")
			}
		})
		t.Run("両面で鳴いたときは、順子の反対の端にも持ち替えられないこと（筋喰い替え）", func(t *testing.T) {
			got := discardLabels(afterCall("4m 5m 6m 7m 8m 9m 1p 2p 3p 4p 5p 6p 9p", mt.ChiAction(1, "3m", "4m 5m")), 1)
			if has(got, "3m") || has(got, "6m") {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("嵌張で鳴いたときは、鳴いた牌そのものしか禁じられないこと", func(t *testing.T) {
			got := discardLabels(afterCall("2m 4m 5m 6m 7m 8m 1p 2p 3p 4p 5p 6p 9p", mt.ChiAction(1, "3m", "2m 4m")), 1)
			if has(got, "3m") || !has(got, "6m") {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("赤5は通常の5と同じ牌として禁じられること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 1: "0m 6m 7m 3m 4m 5m 1p 2p 3p 4p 5p 6p 9p"},
				Draws: "5m", Actions: []kyoku.Action{mt.DiscardAction(0, "5m"), mt.ChiAction(1, "5m", "6m 7m")},
			}).Kyokumen()
			if has(discardLabels(k, 1), "0m") {
				t.Fatal("0m allowed")
			}
		})
	})

	// リーチ後は手牌が固定され、待ちを変えない操作しか残らない。
	t.Run("リーチ後の制約", func(t *testing.T) {
		t.Run("ツモった牌をそのまま切るしかできないこと", func(t *testing.T) {
			k := riichiTurn()
			discards := actionsOfKind(k, kyoku.ActionDiscard)
			drawn, _ := k.Drawn()
			if len(discards) != 1 || discards[0].Tiles()[0] != drawn {
				t.Fatalf("got %v, drawn %v", discards, drawn)
			}
		})
		t.Run("加槓はできないこと", func(t *testing.T) {
			if hasKind(kindsOf(riichiTurn(), -1), kyoku.ActionKakan) {
				t.Fatal("kakan offered")
			}
		})
		t.Run("リーチしている席は手牌を晒せないので鳴けないこと", func(t *testing.T) {
			// リーチ後、下家が 5s を切っても鳴けない（和了れる牌なのでロンだけが残る）。
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands:   map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", 1: "0s 5s 1z 2z 3z 4z 5z 6z 7z 1p 2p 3p 4p"},
				Draws:   "9s 5s",
				Actions: []kyoku.Action{mt.RiichiAction(0, "9s"), mt.DiscardAction(1, "5s")},
			})
			expectKinds(t, kindsOf(k.Kyokumen(), 0), kyoku.ActionRon, kyoku.ActionPass)
		})
	})

	t.Run("他家の打牌への反応", func(t *testing.T) {
		// 席0が牌を切る。
		afterDiscard := func(hands map[int]string, discarded string) *kyoku.Kyokumen {
			all := map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z"}
			for seat, h := range hands {
				all[seat] = h
			}
			return mt.BuildKyoku(mt.KyokuSpec{Hands: all, Draws: discarded, Actions: []kyoku.Action{mt.DiscardAction(0, discarded)}}).Kyokumen()
		}

		t.Run("手牌にツモ牌相当の1枚を加えて和了形になり役が付けばロンを選べること", func(t *testing.T) {
			k := afterDiscard(map[int]string{2: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, "5s")
			if !hasKind(kindsOf(k, 2), kyoku.ActionRon) {
				t.Fatal("no ron")
			}
		})
		t.Run("フリテンならロンを選べないこと", func(t *testing.T) {
			// 席2は 9s 待ちだが、その 9s を自分で切っている。
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1p 2p 3p 4p 5p 6p 7p 8p 9p 1s 2s 3s 4s", 2: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 9s"},
				Draws: "1z 2z 9s 3z 9s",
				Actions: []kyoku.Action{
					mt.DiscardAction(0, "1z"), mt.DiscardAction(1, "2z"), mt.DiscardAction(2, "9s"), mt.DiscardAction(3, "3z"), mt.DiscardAction(0, "9s"),
				},
			}).Kyokumen()
			if !k.Seat(2).IsFuriten() || hasKind(kindsOf(k, 2), kyoku.ActionRon) {
				t.Fatalf("furiten %v kinds %v", k.Seat(2).IsFuriten(), kindsOf(k, 2))
			}
		})
		t.Run("手の内に同種2枚があればポンを選べること", func(t *testing.T) {
			if !hasKind(kindsOf(afterDiscard(map[int]string{2: "9s 9s 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p"}, "9s"), 2), kyoku.ActionPon) {
				t.Fatal("no pon")
			}
		})
		t.Run("手の内に同種3枚があれば明槓を選べること", func(t *testing.T) {
			if !hasKind(kindsOf(afterDiscard(map[int]string{2: "9s 9s 9s 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p"}, "9s"), 2), kyoku.ActionMinkan) {
				t.Fatal("no minkan")
			}
		})
		t.Run("上家の切った数牌で順子が作れればチーを選べること", func(t *testing.T) {
			if !hasKind(kindsOf(afterDiscard(map[int]string{1: "7s 8s 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p"}, "9s"), 1), kyoku.ActionChi) {
				t.Fatal("no chi")
			}
		})
		t.Run("上家以外の捨て牌はチーできないこと", func(t *testing.T) {
			if hasKind(kindsOf(afterDiscard(map[int]string{2: "7s 8s 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p"}, "9s"), 2), kyoku.ActionChi) {
				t.Fatal("chi offered")
			}
		})
		t.Run("字牌はチーできないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 6z", 1: "6z 6z 1s 2s 3s 4s 5s 6s 7s 8s 9s 1p 2p"},
				Draws: "6z", Actions: []kyoku.Action{mt.DiscardAction(0, "6z")},
			}).Kyokumen()
			if hasKind(kindsOf(k, 1), kyoku.ActionChi) {
				t.Fatal("chi offered")
			}
		})
		t.Run("切った本人は自分の捨て牌に反応できないこと", func(t *testing.T) {
			k := afterDiscard(map[int]string{2: "9s 9s 1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p"}, "9s")
			for _, a := range k.LegalActions() {
				if a.Seat() == 0 {
					t.Fatalf("seat 0 offered %v", a)
				}
			}
		})
	})

	// 赤5と通常5は同種だが、どちらを差し出すかで別の鳴きになる。
	t.Run("赤5の扱い", func(t *testing.T) {
		t.Run("赤5と通常5を持てば、ポンで出す2枚として別の選択肢になること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 6m 7m 8m 9m 1p 2p 3p 4p 5p", 2: "0s 5s 5s 1m 2m 3m 4m 6m 7m 8m 9m 1p 2p"},
				Draws: "5s", Actions: []kyoku.Action{mt.DiscardAction(0, "5s")},
			}).Kyokumen()
			pons := actionsOfKind(k, kyoku.ActionPon)
			var got []string
			for _, a := range pons {
				got = append(got, labels(a.Tiles()))
			}
			if len(pons) != 2 || !has(got, "5s 0s") || !has(got, "5s 5s") {
				t.Fatalf("got %v", got)
			}
		})
	})

	t.Run("選べるアクションが無いとき", func(t *testing.T) {
		t.Run("終局後はどの席も何も選べないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)},
			})
			if len(k.Kyokumen().LegalActions()) != 0 {
				t.Fatal("actions after the end")
			}
		})
		t.Run("流局した後も、鳴き待ちのまま終わっているだけで誰も何も選べないこと", func(t *testing.T) {
			hands := map[int]string{}
			for seat := 0; seat < 4; seat++ {
				hands[seat] = "1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 2z 3z 4z"
			}
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: hands, Wall: mt.WallOf(0)})
			result, _ := k.Result()
			if _, claimed := k.Kyokumen().ClaimedTile(); result.Kind() != kyoku.ResultRyukyoku || !claimed || len(k.Kyokumen().LegalActions()) != 0 {
				t.Fatalf("kind %v claimed %v", result.Kind(), claimed)
			}
		})
	})

	// 提示された選択肢がそのまま適用可能なアクションの全体になっている。
	t.Run("選択の提示と検証", func(t *testing.T) {
		t.Run("提示されたアクションはその局面に適用できること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "9s"})
			for _, a := range k.Kyokumen().LegalActions() {
				if _, err := k.Take(a); err != nil {
					t.Errorf("%v: %v", a, err)
				}
			}
		})
	})
}
