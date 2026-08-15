package hand_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

func labelsOf(tiles []tile.Tile) string {
	return strings.Join(tile.Labels(tiles), " ")
}

func kindsOf(melds []hand.Meld) []hand.MeldKind {
	out := make([]hand.MeldKind, len(melds))
	for i, m := range melds {
		out[i] = m.Kind()
	}
	return out
}

func TestNew(t *testing.T) {
	t.Run("accepts 13 closed tiles with no melds", func(t *testing.T) {
		if _, err := hand.New(mt.Tiles("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z"), nil); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rejects 14 closed tiles: a hand is what is left after discarding", func(t *testing.T) {
		_, err := hand.New(mt.Tiles("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z 5z"), nil)
		if !errors.Is(err, hand.ErrClosedTileCount) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("accepts 10 closed tiles with one meld", func(t *testing.T) {
		if _, err := hand.New(mt.Tiles("4p 5p 6p 7s 8s 9s 1z 2z 3z 4z"), []hand.Meld{mt.Chi("1m 2m 3m")}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rejects a closed count that does not match the melds", func(t *testing.T) {
		if _, err := hand.New(mt.Tiles("1m 2m 3m"), nil); !errors.Is(err, hand.ErrClosedTileCount) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects five melds", func(t *testing.T) {
		m := mt.Chi("1m 2m 3m")
		_, err := hand.New(mt.Tiles("1z"), []hand.Meld{m, m, m, m, m})
		if !errors.Is(err, hand.ErrTooManyMelds) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects five of a kind", func(t *testing.T) {
		_, err := hand.New(mt.Tiles("1m 1m 1m 1m 1m 2m 3m 4p 5p 6p 7s 8s 9s"), nil)
		if !errors.Is(err, hand.ErrTooManyCopies) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("counts meld tiles toward the four-copy limit", func(t *testing.T) {
		_, err := hand.New(mt.Tiles("1m 1m 4p 5p 6p 7s 8s 9s 1z 2z"), []hand.Meld{mt.Pon("1m 1m 1m")})
		if !errors.Is(err, hand.ErrTooManyCopies) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("counts a red five as a plain five toward the limit", func(t *testing.T) {
		_, err := hand.New(mt.Tiles("0m 5m 5m 5m 5m 2m 3m 4p 5p 6p 7s 8s 9s"), nil)
		if !errors.Is(err, hand.ErrTooManyCopies) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects an invalid tile", func(t *testing.T) {
		tiles := mt.Tiles("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z")
		_, err := hand.New(append(tiles, 0), nil)
		if !errors.Is(err, hand.ErrInvalidTile) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("Must panics on an invalid hand", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		hand.Must(mt.Tiles("1m"), nil)
	})
	t.Run("copies its inputs", func(t *testing.T) {
		closed := mt.Tiles("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")
		h := hand.Must(closed, nil)
		closed[0] = tile.Chun
		if h.ClosedTiles()[0] != tile.M1 {
			t.Fatal("hand shares the caller's slice")
		}
		h.ClosedTiles()[0] = tile.Chun
		if h.ClosedTiles()[0] != tile.M1 {
			t.Fatal("ClosedTiles() must return a copy")
		}
	})
}

func TestQueries(t *testing.T) {
	h := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z", mt.Chi("4m 5m 6m"))

	t.Run("IsMenzen is false with a called meld", func(t *testing.T) {
		if h.IsMenzen() {
			t.Fatal("menzen")
		}
	})
	t.Run("an ankan keeps the hand menzen", func(t *testing.T) {
		if !mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z", mt.Ankan("5m 5m 5m 5m")).IsMenzen() {
			t.Fatal("not menzen")
		}
	})
	t.Run("AllTiles returns closed tiles plus meld tiles", func(t *testing.T) {
		if got := len(h.AllTiles()); got != 10+3 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("Melds returns the melds in order", func(t *testing.T) {
		if got := kindsOf(h.Melds()); !reflect.DeepEqual(got, []hand.MeldKind{hand.Chi}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("String renders tiles and melds", func(t *testing.T) {
		if got := h.String(); got != "[1m 2m 3m 4p 5p 6p 7s 8s 9s 1z] [chi[4m 5m 6m]]" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestCanHold(t *testing.T) {
	t.Run("is true below four of a kind", func(t *testing.T) {
		h := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")
		if !h.CanHold(tile.Haku) || !h.CanHold(tile.M1) {
			t.Fatal("cannot hold")
		}
	})
	t.Run("is false at four of a kind", func(t *testing.T) {
		if mt.Hand("1m 1m 1m 1m 4p 5p 6p 7s 8s 9s 1z 2z 3z").CanHold(tile.M1) {
			t.Fatal("holds a fifth")
		}
	})
	t.Run("counts meld tiles", func(t *testing.T) {
		if mt.Hand("1m 4p 5p 6p 7s 8s 9s 1z 2z 3z", mt.Pon("1m 1m 1m")).CanHold(tile.M1) {
			t.Fatal("holds a fifth")
		}
	})
	t.Run("counts a red five as a plain five", func(t *testing.T) {
		h := mt.Hand("0m 5m 5m 5m 4p 5p 6p 7s 8s 9s 1z 2z 3z")
		if h.CanHold(tile.M5) || h.CanHold(tile.M5R) {
			t.Fatal("holds a fifth")
		}
	})
}

func TestEqual(t *testing.T) {
	a := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")
	b := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")
	if !a.Equal(b) {
		t.Error("same tiles must be equal")
	}
	if a.Equal(mt.Hand("2m 1m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")) {
		t.Error("tile order is part of identity")
	}
	if a.Equal(mt.Hand("4p 5p 6p 7s 8s 9s 1z 2z 3z 4z", mt.Chi("1m 2m 3m"))) {
		t.Error("melds are part of identity")
	}
	if mt.Hand("4p 5p 6p 7s 8s 9s 1z 2z 3z 4z", mt.Chi("1m 2m 3m")).Equal(mt.Hand("4p 5p 6p 7s 8s 9s 1z 2z 3z 4z", mt.Chi("2m 3m 4m"))) {
		t.Error("meld tiles are part of identity")
	}
}

func TestImprovingTilesAndWaits(t *testing.T) {
	t.Run("lists the tiles that lower the shanten", func(t *testing.T) {
		got := labelsOf(mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 2z 3z").ImprovingTiles())
		if !strings.Contains(got, "2z") || !strings.Contains(got, "3z") {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("leaves out a kind the hand already holds four of", func(t *testing.T) {
		got := labelsOf(mt.Hand("1z 1z 1z 1z 2m 3m 4p 5p 6p 7s 8s 9s 9s").ImprovingTiles())
		if strings.Contains(got, "1z") {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("waits are the improving tiles of a tenpai hand", func(t *testing.T) {
		h := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 2z 2z")
		if !reflect.DeepEqual(h.Waits(), h.ImprovingTiles()) {
			t.Fatal("waits differ from improving tiles")
		}
		if got := labelsOf(h.Waits()); got != "1z 2z" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("a hand that is not tenpai has no waits but still improves", func(t *testing.T) {
		h := mt.Hand("1m 3m 5m 7m 9m 1p 3p 5p 7p 9p 1s 3s 5s")
		if len(h.Waits()) != 0 || len(h.ImprovingTiles()) == 0 {
			t.Fatalf("waits %v improving %v", h.Waits(), h.ImprovingTiles())
		}
	})
	t.Run("answers waits with melds", func(t *testing.T) {
		h := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 2z", mt.Pon("1z 1z 1z"))
		if got := labelsOf(h.Waits()); got != "2z" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestDiscard(t *testing.T) {
	h := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")

	t.Run("removes the discard and keeps the added tile", func(t *testing.T) {
		after, err := h.Discard(tile.M1, tile.Haku)
		if err != nil {
			t.Fatal(err)
		}
		got := labelsOf(after.ClosedTiles())
		if strings.Contains(got, "1m") || !strings.Contains(got, "5z") {
			t.Fatalf("got %q", got)
		}
		if len(after.ClosedTiles()) != 13 {
			t.Fatalf("got %d tiles", len(after.ClosedTiles()))
		}
	})
	t.Run("discarding the added tile leaves the hand unchanged", func(t *testing.T) {
		after, err := h.Discard(tile.Haku, tile.Haku)
		if err != nil || !after.Equal(h) {
			t.Fatalf("got %v, %v", after, err)
		}
	})
	t.Run("keeps the count with melds", func(t *testing.T) {
		melded := mt.Hand("1m 2m 3m 4p 5p 6p 7s 8s 9s 1z", mt.Pon("2z 2z 2z"))
		after, err := melded.Discard(tile.M1, tile.Haku)
		if err != nil || len(after.ClosedTiles()) != 10 {
			t.Fatalf("got %v, %v", after, err)
		}
	})
	t.Run("rejects a tile the hand does not hold", func(t *testing.T) {
		if _, err := h.Discard(tile.P9, tile.Haku); !errors.Is(err, hand.ErrTileNotInHand) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("falls back to the same kind: a plain five for a red five", func(t *testing.T) {
		red := mt.Hand("0m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z")
		after, err := red.Discard(tile.M5, tile.Haku)
		if err != nil || strings.Contains(labelsOf(after.ClosedTiles()), "0m") {
			t.Fatalf("got %v, %v", after, err)
		}
	})
	t.Run("rejects an invalid added tile", func(t *testing.T) {
		if _, err := h.Discard(tile.M1, 0); !errors.Is(err, hand.ErrInvalidTile) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("leaves the original hand untouched", func(t *testing.T) {
		if _, err := h.Discard(tile.M1, tile.Haku); err != nil {
			t.Fatal(err)
		}
		if got := labelsOf(h.ClosedTiles()); got != "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestChiAndPon(t *testing.T) {
	h := mt.Hand("1m 2m 4p 5p 6p 7s 8s 9s 1z 1z 2z 3z 4z")

	t.Run("chi takes two tiles and the discard out of the hand", func(t *testing.T) {
		after, err := h.Chi(tile.M3, mt.Tiles("1m 2m"), tile.North)
		if err != nil {
			t.Fatal(err)
		}
		if got := kindsOf(after.Melds()); !reflect.DeepEqual(got, []hand.MeldKind{hand.Chi}) {
			t.Fatalf("melds %v", got)
		}
		got := labelsOf(after.ClosedTiles())
		if len(after.ClosedTiles()) != 10 || strings.Contains(got, "1m") || strings.Contains(got, "2m") || strings.Contains(got, "4z") {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("pon opens the hand", func(t *testing.T) {
		after, err := h.Pon(tile.East, mt.Tiles("1z 1z"), tile.North)
		if err != nil || after.IsMenzen() {
			t.Fatalf("got %v, %v", after, err)
		}
	})
	t.Run("rejects consumed tiles the hand does not hold", func(t *testing.T) {
		if _, err := h.Pon(tile.Chun, mt.Tiles("7z 7z"), tile.North); !errors.Is(err, hand.ErrTileNotInHand) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects a discard the hand does not hold", func(t *testing.T) {
		if _, err := h.Chi(tile.M3, mt.Tiles("1m 2m"), tile.Chun); !errors.Is(err, hand.ErrTileNotInHand) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("cannot discard a tile used in the call", func(t *testing.T) {
		if _, err := h.Chi(tile.M3, mt.Tiles("1m 2m"), tile.M1); !errors.Is(err, hand.ErrTileNotInHand) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects tiles that do not form the meld", func(t *testing.T) {
		if _, err := h.Chi(tile.M9, mt.Tiles("1m 2m"), tile.North); !errors.Is(err, hand.ErrInvalidMeld) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("leaves the original hand untouched", func(t *testing.T) {
		if _, err := h.Chi(tile.M3, mt.Tiles("1m 2m"), tile.North); err != nil {
			t.Fatal(err)
		}
		if len(h.Melds()) != 0 || len(h.ClosedTiles()) != 13 {
			t.Fatal("mutated")
		}
	})
}

func TestMinkan(t *testing.T) {
	h := mt.Hand("1z 1z 1z 4p 5p 6p 7s 8s 9s 2z 3z 4z 5z")
	after, err := h.Minkan(tile.East, mt.Tiles("1z 1z 1z"))
	if err != nil {
		t.Fatal(err)
	}
	if got := kindsOf(after.Melds()); !reflect.DeepEqual(got, []hand.MeldKind{hand.Minkan}) {
		t.Fatalf("melds %v", got)
	}
	if len(after.ClosedTiles()) != 10 || after.IsMenzen() {
		t.Fatalf("got %v", after)
	}
}

func TestAnkan(t *testing.T) {
	t.Run("uses the added tile among the four and stays menzen", func(t *testing.T) {
		h := mt.Hand("1z 1z 1z 4p 5p 6p 7s 8s 9s 2z 3z 4z 5z")
		after, err := h.Ankan(mt.Tiles("1z 1z 1z 1z"), tile.East)
		if err != nil {
			t.Fatal(err)
		}
		if got := kindsOf(after.Melds()); !reflect.DeepEqual(got, []hand.MeldKind{hand.Ankan}) {
			t.Fatalf("melds %v", got)
		}
		if len(after.ClosedTiles()) != 10 || !after.IsMenzen() {
			t.Fatalf("got %v", after)
		}
	})
	t.Run("can leave the added tile in the hand", func(t *testing.T) {
		h := mt.Hand("1z 1z 1z 1z 4p 5p 6p 7s 8s 9s 2z 3z 4z")
		after, err := h.Ankan(mt.Tiles("1z 1z 1z 1z"), tile.Haku)
		if err != nil || len(after.ClosedTiles()) != 10 || !strings.Contains(labelsOf(after.ClosedTiles()), "5z") {
			t.Fatalf("got %v, %v", after, err)
		}
	})
	t.Run("needs all four tiles", func(t *testing.T) {
		h := mt.Hand("1z 1z 1z 4p 5p 6p 7s 8s 9s 2z 3z 4z 5z")
		if _, err := h.Ankan(mt.Tiles("1z 1z 1z 1z"), tile.Hatsu); !errors.Is(err, hand.ErrTileNotInHand) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("rejects tiles that do not form a quad", func(t *testing.T) {
		h := mt.Hand("1z 1z 1z 4p 5p 6p 7s 8s 9s 2z 3z 4z 5z")
		if _, err := h.Ankan(mt.Tiles("1z 1z 1z 2z"), tile.East); !errors.Is(err, hand.ErrInvalidMeld) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestKakan(t *testing.T) {
	h := mt.Hand("4p 5p 6p 7s 8s 9s 2z 3z 4z 5z", mt.Pon("1z 1z 1z"))

	t.Run("upgrades the pon to a minkan and stays open", func(t *testing.T) {
		after, err := h.Kakan(tile.East, tile.East)
		if err != nil {
			t.Fatal(err)
		}
		melds := after.Melds()
		if len(melds) != 1 || melds[0].Kind() != hand.Minkan || len(melds[0].Tiles()) != 4 {
			t.Fatalf("melds %v", melds)
		}
		if len(after.ClosedTiles()) != 10 || after.IsMenzen() {
			t.Fatalf("got %v", after)
		}
	})
	t.Run("keeps the meld position", func(t *testing.T) {
		two := mt.Hand("4p 5p 6p 7s 8s 9s 2z", mt.Chi("1m 2m 3m"), mt.Pon("1z 1z 1z"))
		after, err := two.Kakan(tile.East, tile.East)
		if err != nil {
			t.Fatal(err)
		}
		if got := kindsOf(after.Melds()); !reflect.DeepEqual(got, []hand.MeldKind{hand.Chi, hand.Minkan}) {
			t.Fatalf("melds %v", got)
		}
	})
	t.Run("rejects a tile that was not ponned", func(t *testing.T) {
		if _, err := h.Kakan(tile.Hatsu, tile.Hatsu); !errors.Is(err, hand.ErrNoPonToKakan) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("matches the pon by kind, so a red five upgrades a plain-five pon", func(t *testing.T) {
		red := mt.Hand("4p 5p 6p 7s 8s 9s 2z 3z 4z 5z", mt.Pon("5m 5m 5m"))
		after, err := red.Kakan(tile.M5R, tile.M5R)
		if err != nil || after.Melds()[0].Kind() != hand.Minkan {
			t.Fatalf("got %v, %v", after, err)
		}
	})
}
