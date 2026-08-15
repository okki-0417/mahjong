package knowledge_test

import (
	"reflect"
	"testing"

	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// 牌。萬子・筒子・索子の1..9と字牌7種で34種類あり、1局にはそれぞれ4枚ずつ入る。
func TestTile(t *testing.T) {
	t.Run("牌の種類", func(t *testing.T) {
		t.Run("数牌は萬子・筒子・索子の3色に1..9があること", func(t *testing.T) {
			for _, suit := range []tile.Suit{tile.Man, tile.Pin, tile.Sou} {
				var numbers []int
				for _, k := range tile.Kinds() {
					if k.Suit() == suit {
						numbers = append(numbers, k.Number())
					}
				}
				if !reflect.DeepEqual(numbers, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
					t.Errorf("%v: %v", suit, numbers)
				}
			}
		})
		t.Run("字牌は東南西北白發中の7種であること", func(t *testing.T) {
			var honors []tile.Tile
			winds, dragons := 0, 0
			for _, k := range tile.Kinds() {
				if !k.IsHonor() {
					continue
				}
				honors = append(honors, k)
				if k.IsWind() {
					winds++
				}
				if k.IsDragon() {
					dragons++
				}
			}
			if !reflect.DeepEqual(tile.Labels(honors), []string{"1z", "2z", "3z", "4z", "5z", "6z", "7z"}) {
				t.Errorf("honors %v", honors)
			}
			if winds != 4 || dragons != 3 {
				t.Errorf("winds %d dragons %d", winds, dragons)
			}
		})
		t.Run("牌の種類は全部で34種であること", func(t *testing.T) {
			if got := len(tile.Kinds()); got != 34 {
				t.Fatalf("got %d", got)
			}
		})
	})

	t.Run("1局に使う牌", func(t *testing.T) {
		fullSet := tile.FullSet(true)

		t.Run("1局に使う牌は136枚であること", func(t *testing.T) {
			if len(fullSet) != 136 {
				t.Fatalf("got %d", len(fullSet))
			}
		})
		t.Run("同じ種類の牌は4枚ずつあること", func(t *testing.T) {
			counts := map[tile.Tile]int{}
			for _, x := range fullSet {
				counts[x.Kind()]++
			}
			for k, c := range counts {
				if c != tile.CopiesPerKind {
					t.Errorf("%v: %d", k, c)
				}
			}
		})
		t.Run("赤5は通常の5と同じ種類として数えること", func(t *testing.T) {
			if mt.T("0m").Kind() != mt.T("5m") {
				t.Error("kind")
			}
			if !mt.T("0m").SameKind(mt.T("5m")) {
				t.Error("same kind")
			}
			if mt.T("0m") == mt.T("5m") {
				t.Error("must not be the same tile")
			}
		})
		t.Run("赤5を入れても牌の種類が増えないこと", func(t *testing.T) {
			kinds := map[tile.Tile]bool{}
			for _, x := range fullSet {
				kinds[x.Kind()] = true
			}
			if len(kinds) != 34 {
				t.Fatalf("got %d kinds", len(kinds))
			}
		})
		t.Run("赤5は各色に1枚ずつであること", func(t *testing.T) {
			var reds []string
			for _, x := range fullSet {
				if x.IsRed() {
					reds = append(reds, x.String())
				}
			}
			if !reflect.DeepEqual(reds, []string{"0m", "0p", "0s"}) {
				t.Fatalf("got %v", reds)
			}
		})
	})
}
