package tile_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/tile"
)

func labelsOf(tiles []tile.Tile) string {
	return strings.Join(tile.Labels(tiles), " ")
}

func TestFullSet(t *testing.T) {
	t.Run("returns the 136 tiles used in a game", func(t *testing.T) {
		if got := len(tile.FullSet(true)); got != 136 {
			t.Fatalf("got %d tiles", got)
		}
	})
	t.Run("keeps four of each kind with red fives", func(t *testing.T) {
		counts := map[tile.Tile]int{}
		for _, x := range tile.FullSet(true) {
			counts[x.Kind()]++
		}
		if len(counts) != 34 {
			t.Fatalf("got %d kinds", len(counts))
		}
		for k, c := range counts {
			if c != 4 {
				t.Errorf("%v: %d copies", k, c)
			}
		}
	})
	t.Run("has one red five per numeric suit", func(t *testing.T) {
		var reds []tile.Tile
		for _, x := range tile.FullSet(true) {
			if x.IsRed() {
				reds = append(reds, x)
			}
		}
		if got := labelsOf(reds); got != "0m 0p 0s" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("has no red fives when red is false", func(t *testing.T) {
		set := tile.FullSet(false)
		for _, x := range set {
			if x.IsRed() {
				t.Fatalf("found red %v", x)
			}
		}
		if len(set) != 136 {
			t.Fatalf("got %d tiles", len(set))
		}
	})
}

func TestKinds(t *testing.T) {
	kinds := tile.Kinds()
	if len(kinds) != 34 {
		t.Fatalf("got %d kinds", len(kinds))
	}
	for _, k := range kinds {
		if k.IsRed() {
			t.Errorf("red five %v listed as a kind", k)
		}
	}
	if got := labelsOf(kinds[:10]); got != "1m 2m 3m 4m 5m 6m 7m 8m 9m 1p" {
		t.Errorf("order: %q", got)
	}
}

func TestParse(t *testing.T) {
	t.Run("returns the tile for a valid label", func(t *testing.T) {
		got, err := tile.Parse("5m")
		if err != nil || got != tile.M5 {
			t.Fatalf("got %v, %v", got, err)
		}
	})
	t.Run("returns ErrInvalidLabel for blank or unknown labels", func(t *testing.T) {
		for _, l := range []string{"", "8z", "xx", "10m", "5M"} {
			if _, err := tile.Parse(l); !errors.Is(err, tile.ErrInvalidLabel) {
				t.Errorf("%q: err = %v", l, err)
			}
		}
	})
	t.Run("MustParse panics on an invalid label", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		tile.MustParse("8z")
	})
	t.Run("ParseAll parses in order and stops at the first bad label", func(t *testing.T) {
		got, err := tile.ParseAll([]string{"1m", "0p", "7z"})
		if err != nil || labelsOf(got) != "1m 0p 7z" {
			t.Fatalf("got %v, %v", got, err)
		}
		if _, err := tile.ParseAll([]string{"1m", "9z"}); !errors.Is(err, tile.ErrInvalidLabel) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestOf(t *testing.T) {
	cases := map[tile.Tile]struct {
		suit tile.Suit
		n    int
	}{
		tile.M1: {tile.Man, 1}, tile.M5: {tile.Man, 5}, tile.P9: {tile.Pin, 9},
		tile.S6: {tile.Sou, 6}, tile.East: {tile.Honor, 1}, tile.Chun: {tile.Honor, 7},
	}
	for want, c := range cases {
		if got := tile.Of(c.suit, c.n); got != want {
			t.Errorf("Of(%v, %d) = %v, want %v", c.suit, c.n, got, want)
		}
	}
	t.Run("panics when no such tile exists", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		tile.Of(tile.Honor, 8)
	})
}

func TestSuitAndNumber(t *testing.T) {
	cases := []struct {
		label  string
		suit   tile.Suit
		number int
		eff    int
	}{
		{"1m", tile.Man, 1, 1}, {"5m", tile.Man, 5, 5}, {"0m", tile.Man, 0, 5}, {"6m", tile.Man, 6, 6}, {"9m", tile.Man, 9, 9},
		{"0p", tile.Pin, 0, 5}, {"9s", tile.Sou, 9, 9},
		{"1z", tile.Honor, 1, 1}, {"4z", tile.Honor, 4, 4}, {"5z", tile.Honor, 5, 5}, {"7z", tile.Honor, 7, 7},
	}
	for _, c := range cases {
		x := tile.MustParse(c.label)
		if x.Suit() != c.suit || x.Number() != c.number || x.EffectiveNumber() != c.eff {
			t.Errorf("%s: suit %v number %d effective %d", c.label, x.Suit(), x.Number(), x.EffectiveNumber())
		}
		if x.String() != c.label {
			t.Errorf("%s: String() = %q", c.label, x.String())
		}
	}
	if tile.Man.String()+tile.Pin.String()+tile.Sou.String()+tile.Honor.String() != "mpsz" {
		t.Error("suit codes")
	}
}

func TestSameKind(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"5m", "5m", true}, {"0m", "5m", true}, {"5m", "0m", true},
		{"5m", "6m", false}, {"5m", "5p", false}, {"0m", "0p", false},
		{"1z", "1z", true}, {"1z", "2z", false},
	}
	for _, c := range cases {
		if got := tile.MustParse(c.a).SameKind(tile.MustParse(c.b)); got != c.want {
			t.Errorf("%s ~ %s = %v", c.a, c.b, got)
		}
	}
	if tile.M5R.Kind() != tile.M5 || tile.M5.Kind() != tile.M5 || tile.East.Kind() != tile.East {
		t.Error("Kind")
	}
	if tile.M5R == tile.M5 {
		t.Error("red five must not equal plain five")
	}
}

func TestPredicates(t *testing.T) {
	type flags struct{ red, numeric, honor, wind, dragon, terminal, yaochuu bool }
	cases := map[string]flags{
		"1m": {terminal: true, yaochuu: true, numeric: true},
		"5m": {numeric: true},
		"0m": {red: true, numeric: true},
		"9s": {terminal: true, yaochuu: true, numeric: true},
		"1z": {honor: true, wind: true, yaochuu: true},
		"4z": {honor: true, wind: true, yaochuu: true},
		"5z": {honor: true, dragon: true, yaochuu: true},
		"7z": {honor: true, dragon: true, yaochuu: true},
	}
	for label, want := range cases {
		x := tile.MustParse(label)
		got := flags{x.IsRed(), x.IsNumeric(), x.IsHonor(), x.IsWind(), x.IsDragon(), x.IsTerminal(), x.IsTerminalOrHonor()}
		if got != want {
			t.Errorf("%s: got %+v want %+v", label, got, want)
		}
	}
	if tile.Tile(0).IsValid() || tile.Tile(38).IsValid() || !tile.Chun.IsValid() {
		t.Error("IsValid")
	}
	if got := tile.Tile(0).String(); got != "Tile(0)" {
		t.Errorf("zero String() = %q", got)
	}
}

func TestOrdering(t *testing.T) {
	t.Run("orders by suit then number", func(t *testing.T) {
		tiles := tile.Sorted([]tile.Tile{tile.P3, tile.M1, tile.Haku, tile.S9, tile.M5R})
		if got := labelsOf(tiles); got != "1m 0m 3p 9s 5z" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("places a red five just after the plain five", func(t *testing.T) {
		tiles := []tile.Tile{tile.M5, tile.M5R, tile.M4, tile.M6}
		tile.Sort(tiles)
		if got := labelsOf(tiles); got != "4m 5m 0m 6m" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("Sorted leaves the input untouched", func(t *testing.T) {
		in := []tile.Tile{tile.P3, tile.M1}
		tile.Sorted(in)
		if !reflect.DeepEqual(in, []tile.Tile{tile.P3, tile.M1}) {
			t.Fatal("input mutated")
		}
	})
}

func TestDora(t *testing.T) {
	cases := map[string]string{
		"1m": "2m", "3p": "4p", "9s": "1s", "9m": "1m", "0m": "6m", "4p": "5p",
		"1z": "2z", "2z": "3z", "3z": "4z", "4z": "1z",
		"5z": "6z", "6z": "7z", "7z": "5z",
	}
	for indicator, want := range cases {
		x := tile.MustParse(indicator)
		if got := x.Dora(); got.String() != want {
			t.Errorf("%s.Dora() = %v, want %s", indicator, got, want)
		}
		if got := tile.MustParse(want).DoraIndicator(); got != x.Kind() {
			t.Errorf("%s.DoraIndicator() = %v, want %v", want, got, x.Kind())
		}
	}
	t.Run("a red five indicates and is indicated as a plain five", func(t *testing.T) {
		if tile.M5R.Dora() != tile.M6 || tile.P5R.DoraIndicator() != tile.P4 || tile.S5R.DoraIndicator() != tile.S4 {
			t.Fatal("red five dora")
		}
	})
	t.Run("every kind round-trips through Dora and DoraIndicator", func(t *testing.T) {
		for _, k := range tile.Kinds() {
			if k.Dora().DoraIndicator() != k {
				t.Errorf("%v", k)
			}
		}
	})
}
