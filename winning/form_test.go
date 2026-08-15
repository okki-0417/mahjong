package winning_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

func forms(closed, win string, melds ...hand.Meld) []winning.Form {
	return winning.Forms(mt.Hand(closed, melds...), mt.T(win), winning.Situation{})
}

func waitKinds(closed, win string, melds ...hand.Meld) []winning.WaitKind {
	seen := map[winning.WaitKind]bool{}
	var out []winning.WaitKind
	for _, f := range forms(closed, win, melds...) {
		if !seen[f.WaitKind()] {
			seen[f.WaitKind()] = true
			out = append(out, f.WaitKind())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func mentsuKinds(f *winning.Form) []winning.MentsuKind {
	var out []winning.MentsuKind
	for _, m := range f.Mentsu() {
		out = append(out, m.Kind())
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func TestForms(t *testing.T) {
	t.Run("decomposition", func(t *testing.T) {
		t.Run("reads an unambiguous hand as one form", func(t *testing.T) {
			got := forms("1m 2m 3m 5m 5m 4p 5p 6p 7s 8s 9s 1z 1z", "5m")
			pair, _ := got[0].PairTile()
			if len(got) != 1 || pair != tile.East {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("reads 555m 666m 777m as three triplets and as three sequences", func(t *testing.T) {
			got := forms("5m 5m 5m 6m 6m 6m 7m 7m 7m 1p 1p 1z 1z", "1p")
			koutsuOnly, withShuntsu := 0, 0
			for i := range got {
				kinds := mentsuKinds(&got[i])
				shuntsu := 0
				for _, k := range kinds {
					if k == winning.Shuntsu {
						shuntsu++
					}
				}
				if shuntsu == 0 {
					koutsuOnly++
				}
				if shuntsu == 3 {
					withShuntsu++
				}
			}
			if koutsuOnly == 0 || withShuntsu == 0 {
				t.Fatalf("koutsu-only %d with-shuntsu %d", koutsuOnly, withShuntsu)
			}
		})
		t.Run("counts melds toward the four sets", func(t *testing.T) {
			got := forms("1m 2m 3m 5m 5m 5m 4p 5p 1z 1z", "6p", mt.Chi("7s 8s 9s"))
			if len(got) != 1 || len(got[0].Mentsu()) != 4 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("reads only the pair with four melds", func(t *testing.T) {
			got := forms("5z", "5z", mt.Chi("1m 2m 3m"), mt.Pon("5p 5p 5p"), mt.Pon("7s 7s 7s"), mt.Pon("1z 1z 1z"))
			pair, _ := got[0].PairTile()
			if len(got) != 1 || pair != tile.Haku {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("makes no honor sequences", func(t *testing.T) {
			if got := forms("1z 2z 3z 4z 5z 6z 7z 1m 2m 3m 4p 5p 6p", "7z"); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("splits a fourth tile between a triplet and a sequence", func(t *testing.T) {
			got := forms("5m 5m 5m 5m 6m 7m 1p 1p 1p 9s 9s 9s 1z", "1z")
			if len(got) != 1 {
				t.Fatalf("got %d forms", len(got))
			}
			if kinds := mentsuKinds(&got[0]); !reflect.DeepEqual(kinds, []winning.MentsuKind{winning.Shuntsu, winning.Koutsu, winning.Koutsu, winning.Koutsu}) {
				t.Fatalf("got %v", kinds)
			}
		})
		t.Run("is empty without a pair", func(t *testing.T) {
			if got := forms("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z", "5z"); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("is empty when the rest does not split into sets", func(t *testing.T) {
			if got := forms("1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 1z 2z 3z", "4z"); len(got) != 0 {
				t.Fatalf("got %v", got)
			}
		})
		t.Run("reads a red five as a plain five", func(t *testing.T) {
			if got := forms("1m 2m 3m 0m 5m 4p 5p 6p 7s 8s 9s 1z 1z", "5m"); len(got) != 1 {
				t.Fatalf("got %v", got)
			}
		})
	})

	t.Run("waits", func(t *testing.T) {
		cases := []struct {
			name   string
			closed string
			win    string
			want   []winning.WaitKind
		}{
			{"the pair is tanki", "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 1z 5m", "5m", []winning.WaitKind{winning.Tanki}},
			{"the middle of a sequence is kanchan", "1m 3m 4p 5p 6p 7s 8s 9s 1z 1z 1z 5z 5z", "2m", []winning.WaitKind{winning.Kanchan}},
			{"the end of a sequence is ryanmen", "3m 4m 4p 5p 6p 7s 8s 9s 1z 1z 1z 5z 5z", "2m", []winning.WaitKind{winning.Ryanmen}},
			{"the 3 of 123 is penchan", "1m 2m 4p 5p 6p 7s 8s 9s 1z 1z 1z 5z 5z", "3m", []winning.WaitKind{winning.Penchan}},
			{"the 7 of 789 is penchan", "8m 9m 4p 5p 6p 7s 8s 9s 1z 1z 1z 5z 5z", "7m", []winning.WaitKind{winning.Penchan}},
			{"a triplet is shanpon", "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 5z 5z", "5z", []winning.WaitKind{winning.Shanpon}},
			{"a tile that fits the pair or a sequence is tanki and ryanmen", "2m 2m 3m 4m 4p 5p 6p 7s 8s 9s 1z 1z 1z", "2m", []winning.WaitKind{winning.Ryanmen, winning.Tanki}},
			{"a tile that fits two sequences gives each reading", "1m 2m 3m 3m 4m 4p 5p 6p 7s 8s 9s 1z 1z", "2m", []winning.WaitKind{winning.Ryanmen, winning.Kanchan}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				if got := waitKinds(c.closed, c.win); !reflect.DeepEqual(got, c.want) {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			})
		}
	})

	t.Run("ron marks the completed set open", func(t *testing.T) {
		ron := winning.Forms(mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 5z 5z"), tile.Haku, southRon())
		tsumo := winning.Forms(mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 5z 5z"), tile.Haku, winning.Situation{WinKind: winning.Tsumo})
		openIn := func(fs []winning.Form) int {
			n := 0
			for _, m := range fs[0].Mentsu() {
				if m.IsOpen() {
					n++
				}
			}
			return n
		}
		if openIn(ron) != 1 || openIn(tsumo) != 0 {
			t.Fatalf("ron %d tsumo %d", openIn(ron), openIn(tsumo))
		}
		if !ron[0].IsMenzen() {
			t.Fatal("a ron-completed set must not break menzen")
		}
	})
}

func TestFormStandard(t *testing.T) {
	t.Run("keeps its parts", func(t *testing.T) {
		f := forms("1m 2m 3m 5m 5m 4p 5p 6p 7s 8s 9s 1z 1z", "5m")[0]
		pair, ok := f.PairTile()
		if f.Kind() != winning.Standard || !f.IsMenzen() || !ok || pair != tile.East || f.WinningTile() != tile.M5 || len(f.PairTiles()) != 0 {
			t.Fatalf("got %+v", f)
		}
	})
	t.Run("is not menzen with a call, which appears as an open set", func(t *testing.T) {
		f := forms("5m 5m 5m 4p 5p 6p 7s 8s 9s 1z", "1z", mt.Chi("1m 2m 3m"))[0]
		open := 0
		for _, m := range f.Mentsu() {
			if m.IsOpen() {
				open++
			}
		}
		if f.IsMenzen() || open != 1 {
			t.Fatalf("menzen %v open %d", f.IsMenzen(), open)
		}
	})
	t.Run("stays menzen with an ankan", func(t *testing.T) {
		if !forms("5m 5m 5m 4p 5p 6p 7s 8s 9s 1z", "1z", mt.Ankan("1m 1m 1m 1m"))[0].IsMenzen() {
			t.Fatal("not menzen")
		}
	})
	t.Run("Mentsu returns a copy", func(t *testing.T) {
		f := forms("1m 2m 3m 5m 5m 4p 5p 6p 7s 8s 9s 1z 1z", "5m")[0]
		f.Mentsu()[0] = winning.Mentsu{}
		if f.Mentsu()[0].Kind() == 0 {
			t.Fatal("shared")
		}
	})
}

func TestFormChiitoitsu(t *testing.T) {
	f := forms("1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z", "7z")[0]
	if f.Kind() != winning.Chiitoitsu || !f.IsMenzen() || f.WaitKind() != winning.Tanki || len(f.PairTiles()) != 7 || len(f.Mentsu()) != 0 {
		t.Fatalf("got %+v", f)
	}
	if _, ok := f.PairTile(); ok {
		t.Fatal("chiitoitsu has no single pair")
	}
}

func TestFormKokushi(t *testing.T) {
	f := forms("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", "1m")[0]
	pair, ok := f.PairTile()
	if f.Kind() != winning.Kokushi || !f.IsMenzen() || f.WaitKind() != winning.Tanki || !ok || pair != tile.M1 {
		t.Fatalf("got %+v", f)
	}
}

func TestMentsu(t *testing.T) {
	f := forms("5m 5m 5m 4p 5p 6p 7s 8s 9s 1z", "1z", mt.Minkan("1m 1m 1m 1m"))[0]
	var kan winning.Mentsu
	for _, m := range f.Mentsu() {
		if m.Kind() == winning.Kantsu {
			kan = m
		}
	}
	if !kan.IsTriplet() || !kan.IsOpen() || len(kan.Tiles()) != 4 || kan.String() != "kantsu(open)[1m 1m 1m 1m]" {
		t.Fatalf("got %v", kan)
	}
	kan.Tiles()[0] = tile.M9
	if kan.Tiles()[0] != tile.M1 {
		t.Fatal("Tiles() must return a copy")
	}
}
