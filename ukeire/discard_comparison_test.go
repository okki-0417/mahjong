package ukeire_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
)

func TestCompareDiscards(t *testing.T) {
	t.Run("lists one candidate per kind, red five folded into the plain five", func(t *testing.T) {
		c, err := ukeire.CompareDiscards(mt.Hand("1m 2m 3m 4m 0m 5m 7p 8p 9p 1s 2s 3s 9s"), tile.M5)
		if err != nil {
			t.Fatal(err)
		}
		var tiles []tile.Tile
		for _, cand := range c.Candidates() {
			tiles = append(tiles, cand.Tile)
		}
		if len(tiles) != 12 {
			t.Fatalf("got %v", tiles)
		}
		for _, x := range tiles {
			if x.IsRed() {
				t.Fatalf("red five listed: %v", tiles)
			}
		}
	})
	t.Run("orders by shanten then by ukeire", func(t *testing.T) {
		c, _ := ukeire.CompareDiscards(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s"), tile.P3)
		cands := c.Candidates()
		if cands[0].Tile != tile.S5 || !cands[0].Shanten.IsTenpai() {
			t.Fatalf("best %+v", cands[0])
		}
		for i := 1; i < len(cands); i++ {
			a, b := cands[i-1], cands[i]
			if a.Shanten > b.Shanten || (a.Shanten == b.Shanten && a.Ukeire.RemainingTotal() < b.Ukeire.RemainingTotal()) {
				t.Fatalf("out of order at %d: %+v then %+v", i, a, b)
			}
		}
	})
	t.Run("Shanten is agari when the draw completes the hand", func(t *testing.T) {
		c, _ := ukeire.CompareDiscards(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s"), tile.S5)
		if c.Shanten() != hand.Agari || !c.IsWinning() {
			t.Fatal("not agari")
		}
	})
	t.Run("Shanten is the best candidate's otherwise", func(t *testing.T) {
		c, _ := ukeire.CompareDiscards(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s"), tile.Chun)
		if c.Shanten() != 1 || c.IsWinning() {
			t.Fatalf("shanten %d", c.Shanten())
		}
	})
	t.Run("rejects an invalid draw and a fifth copy", func(t *testing.T) {
		if _, err := ukeire.CompareDiscards(mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s"), 0); !errors.Is(err, hand.ErrInvalidTile) {
			t.Fatalf("err = %v", err)
		}
		if _, err := ukeire.CompareDiscards(mt.Hand("1m 1m 1m 1m 5m 6m 7m 8m 9m 1p 2p 4p 5s"), tile.M1); !errors.Is(err, hand.ErrTooManyCopies) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("exposes the hand and the draw", func(t *testing.T) {
		h := mt.Hand("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s")
		c, _ := ukeire.CompareDiscards(h, tile.Chun)
		if !c.Hand().Equal(h) || c.Tsumo() != tile.Chun {
			t.Fatal("accessors")
		}
	})
}

func TestCompareDrawnTiles(t *testing.T) {
	t.Run("takes the last tile as the draw", func(t *testing.T) {
		c, err := ukeire.CompareDrawnTiles(mt.Tiles("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s 3p"), nil)
		if err != nil || c.Tsumo() != tile.P3 {
			t.Fatalf("got %v, %v", c.Tsumo(), err)
		}
	})
	t.Run("counts melds", func(t *testing.T) {
		if _, err := ukeire.CompareDrawnTiles(mt.Tiles("4m 5m 6m 7m 8m 9m 1p 2p 4p 5s 3p"), []hand.Meld{mt.Chi("1m 2m 3m")}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rejects the wrong count", func(t *testing.T) {
		if _, err := ukeire.CompareDrawnTiles(mt.Tiles("1m 2m 3m"), nil); !errors.Is(err, ukeire.ErrDrawnCount) {
			t.Fatalf("err = %v", err)
		}
	})
}
