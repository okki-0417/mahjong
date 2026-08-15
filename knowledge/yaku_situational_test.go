package knowledge_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// 状況役だけを見たいので、手役は一気通貫だけに固定した門前のテンパイを使う。
const ittsuTenpai = "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"

// 和了の状況（宣言・巡目・ツモ/ロン・槓）から生まれる役。手牌の構成だけでは判定できない。
func TestSituationalYaku(t *testing.T) {
	type situationEdit func(*winning.Situation)
	winWith := func(kind winning.WinKind, edits ...situationEdit) (*winning.Winning, error) {
		s := sit(kind, tile.EastWind, tile.SouthWind)
		for _, e := range edits {
			e(&s)
		}
		return winOf(ittsuTenpai, "5s", nil, s, ruleset.Default())
	}
	riichi := func(s *winning.Situation) { s.Riichi = true }
	doubleRiichi := func(s *winning.Situation) { s.DoubleRiichi = true }
	ippatsu := func(s *winning.Situation) { s.Ippatsu = true }

	t.Run("立直（リーチ）", func(t *testing.T) {
		t.Run("門前でテンパイし立直を宣言したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("立直")(winWith(winning.Ron, riichi)); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("副露しているとき宣言できないこと", func(t *testing.T) { t.Skip(pendingKyoku) })
	})

	t.Run("ダブル立直（ダブルリーチ）", func(t *testing.T) {
		t.Run("第一巡（自分の最初の打牌）で立直したとき成立し 2翻 となること", func(t *testing.T) {
			if got := hanOf("ダブル立直")(winWith(winning.Ron, doubleRiichi)); got != 2 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("立直と重複して数えないこと", func(t *testing.T) {
			if has(yakuNames(winWith(winning.Ron, doubleRiichi)), "立直") {
				t.Fatal("riichi counted")
			}
		})
	})

	t.Run("一発（イッパツ）", func(t *testing.T) {
		t.Run("立直後1巡以内に、鳴きを挟まず和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("一発")(winWith(winning.Ron, riichi, ippatsu)); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("立直していないとき成立しないこと", func(t *testing.T) {
			if _, err := winWith(winning.Ron, ippatsu); !errors.Is(err, winning.ErrInvalidSituation) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("立直後に他家の鳴きが入ると消えること", func(t *testing.T) { t.Skip(pendingKyoku) })
	})

	t.Run("門前清自摸和（メンゼンツモ）", func(t *testing.T) {
		t.Run("門前でツモ和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("門前清自摸和")(winWith(winning.Tsumo)); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("副露しているとき不成立となること", func(t *testing.T) {
			names := yakuNames(winOf("4m 5m 6m 7m 8m 9m 1p 2p 3p 5s", "5s", []hand.Meld{mt.Chi("1m 2m 3m")},
				sit(winning.Tsumo, tile.EastWind, tile.SouthWind), ruleset.Default()))
			if has(names, "門前清自摸和") {
				t.Fatalf("got %v", names)
			}
		})
	})

	t.Run("海底摸月（ハイテイ）", func(t *testing.T) {
		t.Run("最後の自摸牌でツモ和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("海底摸月")(winWith(winning.Tsumo, func(s *winning.Situation) { s.Haitei = true })); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("河底撈魚（ホウテイ）", func(t *testing.T) {
		t.Run("最後の打牌をロン和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("河底撈魚")(winWith(winning.Ron, func(s *winning.Situation) { s.Houtei = true })); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("嶺上開花（リンシャン）", func(t *testing.T) {
		t.Run("槓の後の嶺上牌でツモ和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("嶺上開花")(winWith(winning.Tsumo, func(s *winning.Situation) { s.Rinshan = true })); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("槍槓（チャンカン）", func(t *testing.T) {
		t.Run("他家の加槓牌をロン和了したとき成立し 1翻 となること", func(t *testing.T) {
			if got := hanOf("槍槓")(winWith(winning.Ron, func(s *winning.Situation) { s.Chankan = true })); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("加槓を横取りした和了に、局面から槍槓が付くこと", func(t *testing.T) { t.Skip(pendingKyoku) })
	})
}
