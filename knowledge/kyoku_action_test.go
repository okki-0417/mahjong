package knowledge_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/okki-0417/mahjong/kyoku"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

// アクション。プレイヤーが局面で下す選択そのもの。
// 局面はアクションを順に適用して復元されるので、アクションの列が対局の真実になる。
func TestAction(t *testing.T) {
	// 鳴かれる牌・ツモ牌・和了牌は、直前に何が起きたかで既に定まっている。
	// プレイヤーが選ぶのは「手の内から何を出すか」だけ。
	t.Run("アクションが持つ牌", func(t *testing.T) {
		t.Run("打牌・リーチは、切る牌1枚を持つこと", func(t *testing.T) {
			if labels(mt.DiscardAction(0, "1m").Tiles()) != "1m" || labels(mt.RiichiAction(0, "1m").Tiles()) != "1m" {
				t.Fatal("tiles")
			}
		})
		t.Run("暗槓・加槓は、対象の牌種1枚を持つこと", func(t *testing.T) {
			if labels(mt.AnkanAction(0, "5z").Tiles()) != "5z" || labels(mt.KakanAction(0, "5z").Tiles()) != "5z" {
				t.Fatal("tiles")
			}
		})
		t.Run("チー・ポンは、手の内から出す2枚を持つこと", func(t *testing.T) {
			if labels(mt.ChiAction(0, "3m", "1m 2m").Tiles()) != "1m 2m" || labels(mt.PonAction(0, "3m", "3m 3m").Tiles()) != "3m 3m" {
				t.Fatal("tiles")
			}
		})
		t.Run("明槓は手の内の3枚すべてを使うので、出す牌を選ばないこと", func(t *testing.T) {
			if len(mt.MinkanAction(0, "3m").Tiles()) != 0 {
				t.Fatal("tiles")
			}
		})
		t.Run("チー・ポン・明槓は、鳴いた牌を面子の同定として併せて持つこと", func(t *testing.T) {
			for _, a := range []kyoku.Action{mt.ChiAction(0, "3m", "1m 2m"), mt.PonAction(0, "3m", "3m 3m"), mt.MinkanAction(0, "3m")} {
				if called, ok := a.CalledTile(); !ok || called != tile.M3 {
					t.Errorf("%v: called %v %v", a, called, ok)
				}
			}
		})
		t.Run("ツモ和了・ロン・九種九牌は牌を持たないこと", func(t *testing.T) {
			for _, a := range []kyoku.Action{kyoku.NewTsumo(0), kyoku.NewRon(0), kyoku.NewKyushukyuhai(0)} {
				if _, ok := a.CalledTile(); ok || len(a.Tiles()) != 0 {
					t.Errorf("%v holds tiles", a)
				}
			}
		})
		t.Run("赤5を使うかどうかは、チー・ポンで出す牌として区別されること", func(t *testing.T) {
			if mt.PonAction(0, "5m", "5m 5m") == mt.PonAction(0, "5m", "0m 5m") {
				t.Fatal("red five ignored")
			}
		})
	})

	t.Run("アクションの同一性", func(t *testing.T) {
		t.Run("席・種別・出す牌が同じアクションは等しいこと", func(t *testing.T) {
			a, b := mt.DiscardAction(0, "1m"), mt.DiscardAction(0, "1m")
			if a != b {
				t.Fatal("not equal")
			}
		})
		t.Run("チーで出す2枚は、並び順が違っても同じアクションとして扱われること", func(t *testing.T) {
			if mt.ChiAction(0, "3m", "1m 2m") != mt.ChiAction(0, "3m", "2m 1m") {
				t.Fatal("order matters")
			}
		})
		t.Run("席が違えば別のアクションであること", func(t *testing.T) {
			if mt.DiscardAction(0, "1m") == mt.DiscardAction(1, "1m") {
				t.Fatal("seat ignored")
			}
		})
	})

	// 局面はアクションの列から復元されるので、アクションは保存して読み戻せる必要がある。
	t.Run("アクションの永続化", func(t *testing.T) {
		roundTrip := func(t *testing.T, original kyoku.Action) {
			t.Helper()
			raw, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var restored kyoku.Action
			if err := json.Unmarshal(raw, &restored); err != nil {
				t.Fatal(err)
			}
			if restored != original {
				t.Fatalf("%v -> %s -> %v", original, raw, restored)
			}
		}
		t.Run("牌のラベルを使って保存できる形に変換でき、そこから元へ戻せること", func(t *testing.T) {
			roundTrip(t, mt.PonAction(2, "3m", "0m 5m"))
		})
		t.Run("選択ではない進行のマーカーも同じように往復できること", func(t *testing.T) {
			roundTrip(t, kyoku.NewPass(1))
			roundTrip(t, kyoku.NewRinshan(1))
		})
	})

	t.Run("成立しないアクション", func(t *testing.T) {
		expectInvalid := func(t *testing.T, kind kyoku.ActionKind, seat int, tiles []tile.Tile, called tile.Tile) {
			t.Helper()
			if _, err := kyoku.NewAction(kind, seat, tiles, called); !errors.Is(err, kyoku.ErrInvalidAction) {
				t.Fatalf("err = %v", err)
			}
		}
		t.Run("種別に対して牌の枚数が合わないアクションは作れないこと", func(t *testing.T) {
			expectInvalid(t, kyoku.ActionDiscard, 0, mt.Tiles("1m 2m"), 0)
		})
		t.Run("存在しない種別のアクションは作れないこと", func(t *testing.T) {
			expectInvalid(t, kyoku.ActionKind(99), 0, nil, 0)
		})
		t.Run("席が0..3の外にあるアクションは作れないこと", func(t *testing.T) {
			expectInvalid(t, kyoku.ActionDiscard, 4, mt.Tiles("1m"), 0)
		})
		t.Run("鳴きでないのに鳴いた牌を持つアクションは作れないこと", func(t *testing.T) {
			expectInvalid(t, kyoku.ActionTsumo, 0, nil, tile.M1)
		})
	})
}

func actionKinds(actions []kyoku.Action) []kyoku.ActionKind {
	var out []kyoku.ActionKind
	seen := map[kyoku.ActionKind]bool{}
	for _, a := range actions {
		if !seen[a.Kind()] {
			seen[a.Kind()] = true
			out = append(out, a.Kind())
		}
	}
	return out
}

func kindsOf(k *kyoku.Kyokumen, seat int) []kyoku.ActionKind {
	var actions []kyoku.Action
	for _, a := range k.LegalActions() {
		if seat < 0 || a.Seat() == seat {
			actions = append(actions, a)
		}
	}
	return actionKinds(actions)
}

func hasKind(kinds []kyoku.ActionKind, kind kyoku.ActionKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func discardLabels(k *kyoku.Kyokumen, seat int) []string {
	var out []string
	for _, a := range k.LegalActions() {
		if a.Kind() == kyoku.ActionDiscard && a.Seat() == seat {
			out = append(out, a.Tiles()[0].String())
		}
	}
	return out
}

func actionsOfKind(k *kyoku.Kyokumen, kind kyoku.ActionKind) []kyoku.Action {
	var out []kyoku.Action
	for _, a := range k.LegalActions() {
		if a.Kind() == kind {
			out = append(out, a)
		}
	}
	return out
}

func mustKyoku(t *testing.T) func(*kyoku.Kyoku, error) *kyoku.Kyoku {
	return func(k *kyoku.Kyoku, err error) *kyoku.Kyoku {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
}

func expectKinds(t *testing.T, got []kyoku.ActionKind, want ...kyoku.ActionKind) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
