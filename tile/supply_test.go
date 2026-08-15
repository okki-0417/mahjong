package tile_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/tile"
)

func TestNewSupply(t *testing.T) {
	t.Run("rejects an invalid tile", func(t *testing.T) {
		if _, err := tile.NewSupply([]tile.Tile{tile.S3, 0}); !errors.Is(err, tile.ErrInvalidLabel) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects five of a kind", func(t *testing.T) {
		if _, err := tile.NewSupply([]tile.Tile{tile.S3, tile.S3, tile.S3, tile.S3, tile.S3}); !errors.Is(err, tile.ErrOversupplied) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("counts a red five with the plain fives", func(t *testing.T) {
		if _, err := tile.NewSupply([]tile.Tile{tile.M5, tile.M5, tile.M5, tile.M5, tile.M5R}); !errors.Is(err, tile.ErrOversupplied) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("accepts exactly four of a kind", func(t *testing.T) {
		if _, err := tile.NewSupply([]tile.Tile{tile.S3, tile.S3, tile.S3, tile.S3}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("MustSupply panics on error", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		tile.MustSupply([]tile.Tile{0})
	})
}

func TestSupplyRemaining(t *testing.T) {
	cases := []struct {
		name string
		seen []tile.Tile
		of   tile.Tile
		want int
	}{
		{"an unseen kind has four", nil, tile.S3, 4},
		{"the zero value has seen nothing", nil, tile.East, 4},
		{"seen copies are subtracted", []tile.Tile{tile.S3, tile.S3}, tile.S3, 2},
		{"a red five counts as a plain five", []tile.Tile{tile.M5R, tile.M5}, tile.M5, 2},
		{"a red five's remaining is the plain five's", []tile.Tile{tile.M5R, tile.M5}, tile.M5R, 2},
		{"a kind seen four times has none", []tile.Tile{tile.S3, tile.S3, tile.S3, tile.S3}, tile.S3, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s tile.Supply
			if c.seen != nil {
				s = tile.MustSupply(c.seen)
			}
			if got := s.Remaining(c.of); got != c.want {
				t.Fatalf("got %d", got)
			}
		})
	}
}

func TestDoraCount(t *testing.T) {
	cases := []struct {
		name       string
		tiles      []tile.Tile
		indicators []tile.Tile
		want       int
	}{
		{"counts each tile that is the indicator's dora", []tile.Tile{tile.M2, tile.M2, tile.M3, tile.M4}, []tile.Tile{tile.M1}, 2},
		{"counts every indicator", []tile.Tile{tile.M2, tile.M3, tile.M4}, []tile.Tile{tile.M1, tile.M2}, 2},
		{"is zero without dora", []tile.Tile{tile.P5, tile.P6, tile.P7}, []tile.Tile{tile.M1}, 0},
		{"a red five is always one", []tile.Tile{tile.M5R, tile.P1, tile.P2}, []tile.Tile{tile.S9}, 1},
		{"a red five that is also indicated counts twice", []tile.Tile{tile.M5R, tile.M5}, []tile.Tile{tile.M4}, 3},
		{"no indicators counts only reds", []tile.Tile{tile.M5R, tile.P5R}, nil, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := tile.DoraCount(c.tiles, c.indicators); got != c.want {
				t.Fatalf("got %d", got)
			}
		})
	}
}

func TestWind(t *testing.T) {
	winds := []tile.Wind{tile.EastWind, tile.SouthWind, tile.WestWind, tile.NorthWind}
	names := []string{"east", "south", "west", "north"}
	tiles := []tile.Tile{tile.East, tile.South, tile.West, tile.North}
	for i, w := range winds {
		if w.String() != names[i] || w.Tile() != tiles[i] || !w.IsValid() {
			t.Errorf("%v", w)
		}
		if parsed, err := tile.ParseWind(names[i]); err != nil || parsed != w {
			t.Errorf("ParseWind(%q) = %v, %v", names[i], parsed, err)
		}
		if w.Next() != winds[(i+1)%4] {
			t.Errorf("%v.Next() = %v", w, w.Next())
		}
	}
	if tile.Wind(0).IsValid() || tile.Wind(5).IsValid() || tile.Wind(0).String() != "Wind(0)" {
		t.Error("invalid winds")
	}
	if _, err := tile.ParseWind("up"); err == nil {
		t.Error("ParseWind accepted garbage")
	}
}
