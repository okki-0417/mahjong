package knowledge_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// 席。1人分の状態。手牌に、局の進行でしか生まれない事柄（河・リーチ・持ち点）が加わる。
func TestSeat(t *testing.T) {
	// 席0が 2m を切り、席2がポンして 9p を切ったところ。
	afterPon := mt.BuildKyoku(mt.KyokuSpec{
		Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
		Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "2m"), mt.PonAction(2, "2m", "2m 2m"), mt.DiscardAction(2, "9p")},
	}).Kyokumen()
	// 席2は 5s で和了れる一通のテンパイ。席0が 5s、続けて席1が 7z を切る。
	seenOwnWinningTile := func(actions ...kyoku.Action) *kyoku.Kyoku {
		return mt.BuildKyoku(mt.KyokuSpec{
			Hands: map[int]string{
				1: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 9s",
				2: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s",
				3: "7z 7z 1z 2z 3z 4z 5z 6z 1m 4m 7m 1p 4p",
			},
			Draws: "5s 7z", Actions: actions,
		})
	}

	t.Run("見え方", func(t *testing.T) {
		t.Run("自分の手の内は自分にしか見えないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 1m 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s"}}).Kyokumen()
			if k.TileSupply(0).Remaining(tile.M1) != 1 || k.TileSupply(1).Remaining(tile.M1) <= 1 {
				t.Fatal("supply")
			}
		})
		t.Run("自分の副露は誰からも見えること", func(t *testing.T) {
			// 晒された3枚は、鳴いた本人にも他家にも同じだけ引かれて見える。
			if afterPon.TileSupply(1).Remaining(tile.M2) != 1 || afterPon.TileSupply(2).Remaining(tile.M2) != 1 {
				t.Fatal("supply")
			}
		})
		t.Run("自分の河は誰からも見えること", func(t *testing.T) {
			if afterPon.TileSupply(1).Remaining(tile.P9) >= tile.CopiesPerKind {
				t.Fatal("supply")
			}
		})
		// ポン・チーは打牌と同時に確定するが、卓の上では宣言した時点で牌が倒れている。
		t.Run("晒すと約束した牌は、確定を待たずに手の内から出ていること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
				Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "2m"), kyoku.NewPass(1), mt.PonAction(2, "2m", "2m 2m")},
			}).Kyokumen()
			if k.SeenBy(0).ConcealedCount(2) != 11 {
				t.Fatalf("got %d", k.SeenBy(0).ConcealedCount(2))
			}
		})
		t.Run("鳴かれた捨て牌は、鳴いた席の副露として数えられ、河には残らないこと", func(t *testing.T) {
			if contains(inRiver(afterPon.Seat(0).Discards()), tile.M2) {
				t.Fatal("still in river")
			}
			supply := afterPon.TileSupply(1)
			if supply.Remaining(tile.M2) != 1 {
				t.Fatal("counted twice")
			}
		})
		// 卓の外にいる打ち手が受け取れるのは、その席に座って見えるぶんだけ。
		t.Run("自分のツモ牌は自分にしか見えないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "5s"}).Kyokumen()
			drawn, ok := k.SeenBy(0).Drawn()
			_, other := k.SeenBy(1).Drawn()
			if !ok || drawn != tile.S5 || other {
				t.Fatal("drawn")
			}
		})
		t.Run("他家について見えるのは、河・副露・鳴きの約束・持ち点・リーチと手の内の枚数だけであること", func(t *testing.T) {
			raw, err := json.Marshal(mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 1m 1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s"}}).SeenBy(1))
			if err != nil {
				t.Fatal(err)
			}
			var seen struct {
				Seats []map[string]any `json:"seats"`
			}
			if err := json.Unmarshal(raw, &seen); err != nil {
				t.Fatal(err)
			}
			keys := map[string]bool{}
			for _, s := range seen.Seats {
				if s["seat"] == float64(1) {
					continue
				}
				for k := range s {
					keys[k] = true
				}
			}
			var got []string
			for k := range keys {
				got = append(got, k)
			}
			sort.Strings(got)
			want := []string{"concealed_count", "discards", "melds", "pending_call", "riichi", "score", "seat", "seat_wind"}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("選べる選択は自分のぶんしか見えないこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "2m 1m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 4s", 2: "2m 2m 3p 4p 5p 6p 7p 8p 9p 1s 5s 6s 7s"},
				Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, "2m")},
			}).Kyokumen()
			expectKinds(t, actionKinds(k.SeenBy(2).LegalActions()), kyoku.ActionPon, kyoku.ActionPass)
			if len(k.SeenBy(0).LegalActions()) != 0 {
				t.Fatal("seat 0 sees actions")
			}
		})
	})

	t.Run("捨て牌", func(t *testing.T) {
		t.Run("切った牌が河に並ぶこと", func(t *testing.T) {
			if labels(inRiver(afterPon.Seat(2).Discards())) != "9p" {
				t.Fatal("river")
			}
		})
		t.Run("ツモ切りか手出しかが分かること", func(t *testing.T) {
			spec := func(discard string) *kyoku.Kyokumen {
				return mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 3s 9s"}, Draws: "5z", Actions: []kyoku.Action{mt.DiscardAction(0, discard)}}).Kyokumen()
			}
			last := func(k *kyoku.Kyokumen) kyoku.Discard {
				ds := k.Seat(0).Discards()
				return ds[len(ds)-1]
			}
			if !last(spec("5z")).IsTsumogiri() || last(spec("1m")).IsTsumogiri() {
				t.Fatal("tsumogiri")
			}
		})
		t.Run("リーチ宣言牌が分かること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "9s", Actions: []kyoku.Action{mt.RiichiAction(0, "9s")}}).Kyokumen()
			ds := k.Seat(0).Discards()
			if !ds[len(ds)-1].IsRiichiDeclaration() {
				t.Fatal("declaration")
			}
		})
		t.Run("鳴かれた牌は誰に鳴かれたかが分かること", func(t *testing.T) {
			ds := afterPon.Seat(0).Discards()
			if by, ok := ds[len(ds)-1].CalledBy(); !ok || by != 2 {
				t.Fatal("called by")
			}
		})
	})

	// フリテンには3つの原因があるが、いずれも「自分の待ち牌で和了れない」という一つの状態。
	t.Run("フリテン", func(t *testing.T) {
		// 5s 単騎のテンパイ。ツモった牌をそのまま切った席0を返す。
		discardedOwn := func(draw string) kyoku.SeatState {
			return mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: draw, Actions: []kyoku.Action{mt.DiscardAction(0, draw)}}).Kyokumen().Seat(0)
		}

		t.Run("自分の待ち牌を自分が切っていればフリテンであること", func(t *testing.T) {
			if !discardedOwn("5s").IsFuriten() {
				t.Fatal("not furiten")
			}
		})
		t.Run("赤5を切っていれば通常の5の待ちもフリテンであること", func(t *testing.T) {
			if !discardedOwn("0s").IsFuriten() {
				t.Fatal("not furiten")
			}
		})
		t.Run("鳴かれて河から消えた牌でも、自分が切った以上フリテンであること", func(t *testing.T) {
			s := mt.BuildKyoku(mt.KyokuSpec{
				Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", 2: "5s 0s 3p 4p 5p 6p 7p 8p 9p 1s 6s 7s 1z"},
				Draws: "5s", Actions: []kyoku.Action{mt.DiscardAction(0, "5s"), mt.PonAction(2, "5s", "5s 0s"), mt.DiscardAction(2, "1z")},
			}).Kyokumen().Seat(0)
			if len(inRiver(s.Discards())) != 0 || !s.IsFuriten() {
				t.Fatal("furiten")
			}
		})
		t.Run("待ち牌を1枚も切っていなければフリテンでないこと", func(t *testing.T) {
			if discardedOwn("1z").IsFuriten() {
				t.Fatal("furiten")
			}
		})
		t.Run("テンパイしていなければフリテンでないこと", func(t *testing.T) {
			s := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 2z 3z 4z"}, Draws: "1m", Actions: []kyoku.Action{mt.DiscardAction(0, "1m")}}).Kyokumen().Seat(0)
			if s.IsFuriten() {
				t.Fatal("furiten")
			}
		})
	})

	t.Run("見逃しによるフリテン", func(t *testing.T) {
		t.Run("和了牌を見逃すと、次に自分がツモるまでフリテンであること", func(t *testing.T) {
			// 席2 は 5s で和了れるが、席0 の 5s に答えないまま席1 の手番へ移る。
			// 席1 の捨てた 7z は席3 がポンできるので、席2 がツモる前の局面が残る。
			missed := seenOwnWinningTile(mt.DiscardAction(0, "5s"), mt.DiscardAction(1, "7z"))
			drewAgain := seenOwnWinningTile(mt.DiscardAction(0, "5s"), mt.DiscardAction(1, "7z"), kyoku.NewPass(3))
			seat, _ := drewAgain.Kyokumen().DiscardingSeat()
			if !missed.Kyokumen().Seat(2).IsFuriten() || seat != 2 || drewAgain.Kyokumen().Seat(2).IsFuriten() {
				t.Fatal("furiten")
			}
		})
		t.Run("リーチ後に和了牌を見逃すと、その局のあいだフリテンが続くこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{
				Hands:   map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"},
				Draws:   "9s 5s 1z 2z 3z",
				Actions: []kyoku.Action{mt.RiichiAction(0, "9s"), mt.DiscardAction(1, "5s"), mt.DiscardAction(2, "1z"), mt.DiscardAction(3, "2z")},
			})
			if !k.Kyokumen().Seat(0).IsFuriten() {
				t.Fatal("not furiten")
			}
		})
	})

	t.Run("リーチ", func(t *testing.T) {
		t.Run("リーチしていない席は一発もダブルリーチも成立しないこと", func(t *testing.T) {
			s := mt.BuildKyoku(mt.KyokuSpec{}).Kyokumen().Seat(0)
			if s.IsRiichi() || s.IsIppatsu() || s.IsDoubleRiichi() {
				t.Fatal("riichi flags")
			}
		})
		t.Run("第一巡以外のリーチはダブルリーチにならないこと", func(t *testing.T) {
			s := mt.BuildKyoku(mt.KyokuSpec{
				Hands:   map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"},
				Draws:   "1z 2z 3z 4z 9s",
				Actions: append(discards(0, "1z", "2z", "3z", "4z"), mt.RiichiAction(0, "9s")),
			}).Kyokumen().Seat(0)
			if !s.IsRiichi() || s.IsDoubleRiichi() {
				t.Fatal("double riichi")
			}
		})
		t.Run("第一巡のリーチはダブルリーチになること", func(t *testing.T) {
			s := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "9s", Actions: []kyoku.Action{mt.RiichiAction(0, "9s")}}).Kyokumen().Seat(0)
			if !s.IsDoubleRiichi() {
				t.Fatal("not double")
			}
		})
	})

	t.Run("持ち点", func(t *testing.T) {
		t.Run("席ごとに持ち点を持つこと", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Scores: map[int]int{0: 32000, 1: 18000}}).Kyokumen()
			if k.Score(0) != 32000 || k.Score(1) != 18000 || k.Score(2) != 25000 || k.Score(3) != 25000 {
				t.Fatal("scores")
			}
		})
		t.Run("リーチを宣言すると1000点が供託に出ること", func(t *testing.T) {
			k := mt.BuildKyoku(mt.KyokuSpec{Hands: map[int]string{0: "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"}, Draws: "9s", Actions: []kyoku.Action{mt.RiichiAction(0, "9s")}}).Kyokumen()
			if k.Score(0) != 24000 || k.RiichiSticks() != 1 {
				t.Fatal("stick")
			}
		})
	})
}
