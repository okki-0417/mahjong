package knowledge_test

import (
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/winning"
)

const (
	// 親の一通テンパイ。5s をツモれば和了、5s を切れば他家がロンできる。
	ittsuHand = "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"
	// 字牌の対子だけを抱えた、和了にもテンパイにも遠い手。
	notenHand = "1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 2z 3z 4z"
)

func allSeats(hand string) map[int]string {
	out := map[int]string{}
	for seat := 0; seat < 4; seat++ {
		out[seat] = hand
	}
	return out
}

func expectDeltas(t *testing.T, r *kyoku.Result, want [4]int) {
	t.Helper()
	if r.Deltas() != want {
		t.Errorf("deltas %v, want %v", r.Deltas(), want)
	}
}

// 局の決着。終わり方が、点棒の動き・親の継続・次局の本場・持ち越す供託を決める。
func TestKyokuEnd(t *testing.T) {
	t.Run("和了", func(t *testing.T) {
		t.Run("ツモ和了は、和了者以外の3人が払うこと", func(t *testing.T) {
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsuHand}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			d := r.Deltas()
			if d[0] <= 0 || d[1] >= 0 || d[1] != d[2] || d[2] != d[3] || d[0]+d[1]+d[2]+d[3] != 0 {
				t.Fatalf("deltas %v", d)
			}
		})
		t.Run("ロン和了は、放銃した1人だけが払うこと", func(t *testing.T) {
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 2: ittsuHand},
				Draws: "5s", Actions: []kyoku.Action{mt.DiscardAction(0, "5s"), kyoku.NewRon(2)},
			}))
			w, _ := r.Winner()
			l, _ := r.Loser()
			d := r.Deltas()
			if w != 2 || l != 0 || d[0] != -d[2] || d[1] != 0 || d[3] != 0 {
				t.Fatalf("winner %d loser %d deltas %v", w, l, d)
			}
		})
		t.Run("本場は和了者への上乗せになり、ロンは放銃者が全額を負うこと", func(t *testing.T) {
			settle := func(honba int) *kyoku.Result {
				return resultOf(t, mt.BuildKyoku(mt.KyokuSpec{
					Hands: map[int]string{0: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 2: ittsuHand},
					Draws: "5s", Honba: honba, Actions: []kyoku.Action{mt.DiscardAction(0, "5s"), kyoku.NewRon(2)},
				}))
			}
			if settle(1).Deltas()[2]-settle(0).Deltas()[2] != 300 || settle(1).Deltas()[0]-settle(0).Deltas()[0] != -300 {
				t.Fatal("honba")
			}
		})
		t.Run("供託の点棒は和了者が総取りすること", func(t *testing.T) {
			with := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsuHand}, Draws: "5s", RiichiSticks: 2, Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			without := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsuHand}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			if with.Deltas()[0]-without.Deltas()[0] != 2000 || with.CarriedRiichiSticks() != 0 {
				t.Fatal("sticks")
			}
		})
		t.Run("親が和了れば連荘し、子が和了れば親が流れること", func(t *testing.T) {
			dealerWon := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsuHand}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			childWon := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{1: ittsuHand}, Draws: "5z 5s", Actions: []kyoku.Action{mt.DiscardAction(0, "5z"), kyoku.NewTsumo(1)}}))
			if !dealerWon.DealerContinues() || dealerWon.NextHonba() != 1 || childWon.DealerContinues() || childWon.NextHonba() != 0 {
				t.Fatal("dealer continuation")
			}
		})
		t.Run("何の役で何翻何符の和了だったかが決着から分かること", func(t *testing.T) {
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: ittsuHand}, Draws: "5s", Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			score, ok := r.Score()
			if !ok {
				t.Fatal("no score")
			}
			ittsu := false
			for _, y := range score.Yakus() {
				if y.ID == winning.YakuIttsu {
					ittsu = true
				}
			}
			if !ittsu || score.Han() <= 0 || score.Fu() <= 0 || score.Total() != r.Deltas()[0] {
				t.Fatalf("score %+v deltas %v", score, r.Deltas())
			}
		})
	})

	// ツモ山が尽きて誰も和了れないまま終わる、最も普通の流局。
	t.Run("荒牌平局", func(t *testing.T) {
		// 全員がツモ切りだけで最後まで打ち、誰も和了らずに山を尽きさせる。
		exhaust := func(hands map[int]string, sticks int) *kyoku.Kyoku {
			return mt.BuildKyoku(mt.KyokuSpec{Hands: hands, Wall: mt.WallOf(0), RiichiSticks: sticks})
		}

		t.Run("ツモ山が尽き、最後の打牌に誰も反応しなければ流局すること", func(t *testing.T) {
			k := exhaust(allSeats(notenHand), 0)
			if k.Kyokumen().RemainingDraws() != 0 || resultOf(t, k).Kind() != kyoku.ResultRyukyoku {
				t.Fatal("not ryukyoku")
			}
		})
		t.Run("誰も和了っていないので、役も点数も付かないこと", func(t *testing.T) {
			if _, ok := resultOf(t, exhaust(allSeats(notenHand), 0)).Score(); ok {
				t.Fatal("scored")
			}
		})
		t.Run("河底牌はロンでしか取れず、ポン・チー・カンでは鳴けないこと", func(t *testing.T) {
			for _, a := range exhaust(allSeats(notenHand), 0).Kyokumen().LegalActions() {
				if a.Kind() != kyoku.ActionRon {
					t.Fatalf("offered %v", a)
				}
			}
		})
		t.Run("テンパイしている席とノーテンの席のあいだで、合計3000点をやり取りすること", func(t *testing.T) {
			one := resultOf(t, exhaust(map[int]string{0: ittsuHand, 1: notenHand, 2: notenHand, 3: notenHand}, 0))
			two := resultOf(t, exhaust(map[int]string{0: ittsuHand, 1: "1s 2s 3s 4s 5s 6s 7s 8s 9s 4p 5p 6p 7p", 2: notenHand, 3: notenHand}, 0))
			expectDeltas(t, one, [4]int{3000, -1000, -1000, -1000})
			expectDeltas(t, two, [4]int{1500, 1500, -1500, -1500})
		})
		t.Run("全員ノーテンなら点棒が動かないこと", func(t *testing.T) {
			expectDeltas(t, resultOf(t, exhaust(allSeats(notenHand), 0)), [4]int{})
		})
		t.Run("親がテンパイなら連荘し、ノーテンなら親が流れること", func(t *testing.T) {
			dealerTenpai := resultOf(t, exhaust(map[int]string{0: ittsuHand, 1: notenHand, 2: notenHand, 3: notenHand}, 0))
			dealerNoten := resultOf(t, exhaust(map[int]string{0: notenHand, 1: "1s 2s 3s 4s 5s 6s 7s 8s 9s 4p 5p 6p 7p", 2: notenHand, 3: notenHand}, 0))
			if !dealerTenpai.DealerContinues() || dealerNoten.DealerContinues() {
				t.Fatal("dealer continuation")
			}
		})
		t.Run("本場が1つ増え、供託の点棒は次局へ持ち越されること", func(t *testing.T) {
			r := resultOf(t, exhaust(allSeats(notenHand), 2))
			if r.NextHonba() != 1 || r.CarriedRiichiSticks() != 2 {
				t.Fatal("honba or sticks")
			}
		})
	})

	// 流し満貫。荒牌平局のとき、河が么九牌だけで一度も鳴かれていない席は、
	// テンパイ料ではなく満貫の和了として扱われる。採用しないルールもある。
	t.Run("流し満貫", func(t *testing.T) {
		// 完成面子を作れない中張牌だけのノーテン。么九牌の供給を河に回すために使う。
		const chuuchanNoten = "2m 3m 6m 8m 2p 3p 6p 8p 2s 3s 6s 8s 4m"
		// 指定した席のツモをすべて么九牌にして、全員ツモ切りで流局まで打ち切る卓。
		nagashi := func(seats []int, rs ruleset.RuleSet) *kyoku.Kyoku {
			var yaochuu []string
			for _, l := range []string{"1m", "9m", "1p", "9p", "1s", "9s", "1z", "2z", "3z", "4z", "5z", "6z", "7z"} {
				for i := 0; i < 4; i++ {
					yaochuu = append(yaochuu, l)
				}
			}
			draws := map[int]string{}
			for _, seat := range seats {
				n := 17
				if seat < 2 {
					n = 18
				}
				draws[seat] = strings.Join(yaochuu[:n], " ")
				yaochuu = yaochuu[n:]
			}
			return mt.BuildKyoku(mt.KyokuSpec{Hands: allSeats(chuuchanNoten), SeatDraws: draws, Wall: mt.WallOf(0), RuleSet: rs})
		}
		// 残りの数枚を、ツモ切りと見送りだけで流局まで打ち切る。
		playOut := func(t *testing.T, k *kyoku.Kyoku) *kyoku.Kyoku {
			t.Helper()
			for !k.IsFinished() {
				seat, ok := k.Kyokumen().DiscardingSeat()
				if !ok {
					k = mt.AfterOthersPass(k)
					continue
				}
				drawn, _ := k.Kyokumen().Drawn()
				k = mustKyoku(t)(k.Take(kyoku.NewDiscard(seat, drawn)))
			}
			return k
		}

		t.Run("自分の捨て牌がすべて么九牌で、1枚も鳴かれていなければ成立すること", func(t *testing.T) {
			if resultOf(t, nagashi([]int{1}, ruleset.Default())).Deltas()[1] <= 0 {
				t.Fatal("not paid")
			}
		})
		t.Run("中張牌を1枚でも切っていれば不成立となること", func(t *testing.T) {
			expectDeltas(t, resultOf(t, nagashi(nil, ruleset.Default())), [4]int{})
		})
		t.Run("么九牌だけでも、1枚でも鳴かれていれば不成立となること", func(t *testing.T) {
			// 席0は么九牌しか切っていないが、最後の 9m を席2に鳴かれている。
			// 河底の1枚前なのでポンは実際に選べる。
			k := playOut(t, mt.BuildKyoku(mt.KyokuSpec{
				Hands:     map[int]string{0: chuuchanNoten, 2: "9m 9m 4m 1z 2z 3z 4z 5z 6z 7z 1p 4p 7p"},
				SeatDraws: map[int]string{0: "9p 9p 9p 9p 9s 9s 9s 9s 1s 1s 1s 1s 1m 1m 1m 1m 6z 9m"},
				Wall:      mt.WallOf(2),
				Actions:   []kyoku.Action{mt.DiscardAction(0, "9m"), mt.PonAction(2, "9m", "9m 9m"), mt.DiscardAction(2, "4m")},
			}))
			r := resultOf(t, k)
			if r.Kind() != kyoku.ResultRyukyoku || r.Deltas()[0] > 0 {
				t.Fatalf("kind %v deltas %v", r.Kind(), r.Deltas())
			}
		})
		t.Run("テンパイ料のやり取りに代えて、満貫と同じ点数を受け取ること", func(t *testing.T) {
			expectDeltas(t, resultOf(t, nagashi([]int{1}, ruleset.Default())), [4]int{-4000, 8000, -2000, -2000})
		})
		t.Run("親の流し満貫は親の満貫と同じく 12000点 になること", func(t *testing.T) {
			expectDeltas(t, resultOf(t, nagashi([]int{0}, ruleset.Default())), [4]int{12000, -4000, -4000, -4000})
		})
		t.Run("親が流し満貫なら親の和了と同じく連荘すること", func(t *testing.T) {
			if !resultOf(t, nagashi([]int{0}, ruleset.Default())).DealerContinues() {
				t.Fatal("dealer lost")
			}
		})
		t.Run("複数の席が同時に成立することがあること", func(t *testing.T) {
			d := resultOf(t, nagashi([]int{1, 2}, ruleset.Default())).Deltas()
			if d[1] <= 0 || d[2] <= 0 || d[0]+d[1]+d[2]+d[3] != 0 {
				t.Fatalf("deltas %v", d)
			}
		})
		t.Run("採用しないルールではテンパイ料のやり取りに戻ること", func(t *testing.T) {
			expectDeltas(t, resultOf(t, nagashi([]int{1}, ruleset.Default().WithNagashiMangan(false))), [4]int{})
		})
	})

	// 責任払い（包／パオ）。役満を確定させる牌を鳴かせた席は、和了者への支払いを負う。
	// 確定させた時点で結果が決まっているので、放銃していなくても責任を問われる。
	t.Run("責任払い（パオ）", func(t *testing.T) {
		// 席0が白發中を順に鳴いて大三元を確定させる卓。3つ目の中を出すのは席1なので、
		// 席1が責任を負う。以降 席0は 4m5m6m + 6m6m のテンパイで 3m・6m を待つ（子の役満）。
		// 待ち牌を自分で切るとフリテンでロンできないので、鳴くたびに切るのは 9p 9p 1s。
		daisangen := func(last []kyoku.Action, draws, closed, leave string) *kyoku.Kyoku {
			return mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{
					0: closed,
					1: "6z 7z 1m 2m 8m 9m 1p 2p 3p 4p 5p 6p 7p",
					2: "5z 1s 2s 3s 4s 5s 6s 7s 8s 9s 9p 8p 8p",
				},
				Dealer: 2, Draws: draws,
				Actions: append([]kyoku.Action{
					mt.DiscardAction(2, "5z"),
					mt.PonAction(0, "5z", "5z 5z"), mt.DiscardAction(0, "9p"),
					mt.DiscardAction(1, "6z"),
					mt.PonAction(0, "6z", "6z 6z"), mt.DiscardAction(0, "9p"),
					mt.DiscardAction(1, "7z"),
					mt.PonAction(0, "7z", "7z 7z"), mt.DiscardAction(0, leave),
				}, last...),
			})
		}
		const defaultClosed = "5z 5z 6z 6z 7z 7z 4m 5m 6m 6m 9p 9p 1s"
		toTsumo := []kyoku.Action{mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "2z"), mt.DiscardAction(3, "2z"), kyoku.NewTsumo(0)}

		t.Run("三元牌の3つ目を鳴かせた席は、大三元の責任を負うこと", func(t *testing.T) {
			r := resultOf(t, daisangen(toTsumo, "1z 1z 1z 1z 2z 2z 6m", defaultClosed, "1s"))
			if w, _ := r.Winner(); w != 0 || r.Deltas()[1] != -32000 {
				t.Fatalf("winner %d deltas %v", w, r.Deltas())
			}
		})
		t.Run("ツモ和了なら、責任を負う席が役満の全額を払うこと", func(t *testing.T) {
			expectDeltas(t, resultOf(t, daisangen(toTsumo, "1z 1z 1z 1z 2z 2z 6m", defaultClosed, "1s")), [4]int{32000, -32000, 0, 0})
		})
		t.Run("ロン和了なら、放銃者と責任を負う席が半分ずつ払うこと", func(t *testing.T) {
			r := resultOf(t, daisangen([]kyoku.Action{mt.DiscardAction(1, "1z"), mt.DiscardAction(2, "2z"), mt.DiscardAction(3, "6m"), kyoku.NewRon(0)},
				"1z 1z 1z 1z 2z 6m", defaultClosed, "1s"))
			expectDeltas(t, r, [4]int{32000, -16000, 0, -16000})
		})
		t.Run("責任を負う席が放銃したときは、その席が全額を払うこと", func(t *testing.T) {
			r := resultOf(t, daisangen([]kyoku.Action{mt.DiscardAction(1, "6m"), kyoku.NewRon(0)}, "1z 1z 1z 6m", defaultClosed, "1s"))
			expectDeltas(t, r, [4]int{32000, -32000, 0, 0})
		})
		t.Run("確定させた鳴きが暗刻や暗槓なら、誰も責任を負わないこと", func(t *testing.T) {
			// 白發中がすべて手の内で揃うので、誰にも鳴かせていない。親の役満をツモ和了。
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "5z 5z 5z 6z 6z 6z 7z 7z 7z 4m 5m 6m 6m"}, Draws: "6m", Actions: []kyoku.Action{kyoku.NewTsumo(0)}}))
			d := r.Deltas()
			if d[1] != -16000 || d[2] != -16000 || d[3] != -16000 {
				t.Fatalf("deltas %v", d)
			}
		})
		t.Run("四風牌の4つ目を鳴かせた席は、大四喜の責任を負うこと", func(t *testing.T) {
			r := resultOf(t, mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{
					0: "1z 1z 2z 2z 3z 3z 4z 4z 9m 9m 9p 9p 9p",
					1: "3z 4z 1m 2m 8m 1p 2p 3p 4p 5p 6p 7p 8p",
					2: "1z 2z 1s 2s 3s 4s 5s 6s 7s 8s 9s 6z 6z",
				},
				Dealer: 2, Draws: "5z 5z 5z 5z 7z 7z 7z 7z 9m",
				Actions: []kyoku.Action{
					mt.DiscardAction(2, "1z"),
					mt.PonAction(0, "1z", "1z 1z"), mt.DiscardAction(0, "9p"),
					mt.DiscardAction(1, "3z"),
					mt.PonAction(0, "3z", "3z 3z"), mt.DiscardAction(0, "9p"),
					mt.DiscardAction(1, "4z"),
					mt.PonAction(0, "4z", "4z 4z"), mt.DiscardAction(0, "9p"),
					mt.DiscardAction(1, "5z"), mt.DiscardAction(2, "2z"),
					mt.PonAction(0, "2z", "2z 2z"), mt.DiscardAction(0, "9m"),
					mt.DiscardAction(1, "7z"), mt.DiscardAction(2, "7z"), mt.DiscardAction(3, "7z"), kyoku.NewTsumo(0),
				},
			}))
			// 4つ目の四風牌を出したのは席2。
			expectDeltas(t, r, [4]int{32000, 0, -32000, 0})
		})
		t.Run("責任の対象でない役満が複合しても、責任を負う分はその役満に限られること", func(t *testing.T) {
			// 大三元 + 字一色 のダブル役満（子64000）。責任を負うのは大三元の1つ分だけで、
			// 残りは通常のツモ払いになる。
			r := resultOf(t, daisangen(
				[]kyoku.Action{mt.DiscardAction(1, "3z"), mt.DiscardAction(2, "2z"), mt.DiscardAction(3, "2z"), kyoku.NewTsumo(0)},
				"3z 3z 3z 3z 2z 2z 4z", "5z 5z 6z 6z 7z 7z 1z 1z 4z 4z 9p 9p 9p", "9p"))
			// 大三元ぶん32000は責任者が全額、字一色ぶん32000は親16000・子8000ずつ。
			expectDeltas(t, r, [4]int{64000, -40000, -16000, -8000})
		})
	})
}
