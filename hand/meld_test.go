package hand_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
	"github.com/okki-0417/mahjong/tile"
)

func TestNewMeld(t *testing.T) {
	accepted := []struct {
		name  string
		kind  hand.MeldKind
		tiles string
	}{
		{"a sequence", hand.Chi, "3p 4p 5p"},
		{"a sequence with a red five", hand.Chi, "4m 0m 6m"},
		{"a sequence given out of order", hand.Chi, "5p 3p 4p"},
		{"a triplet", hand.Pon, "1z 1z 1z"},
		{"a triplet with a red five", hand.Pon, "5p 5p 0p"},
		{"an open quad", hand.Minkan, "5z 5z 5z 5z"},
		{"a concealed quad", hand.Ankan, "1m 1m 1m 1m"},
	}
	for _, c := range accepted {
		t.Run("accepts "+c.name, func(t *testing.T) {
			m, err := hand.NewMeld(c.kind, mt.Tiles(c.tiles))
			if err != nil {
				t.Fatal(err)
			}
			if m.Kind() != c.kind || len(m.Tiles()) != len(mt.Tiles(c.tiles)) {
				t.Fatalf("got %v", m)
			}
		})
	}
	t.Run("keeps the given tile order", func(t *testing.T) {
		m := hand.MustMeld(hand.Chi, mt.Tiles("5p 3p 4p"))
		if got := tile.Labels(m.Tiles()); got[0] != "5p" || got[1] != "3p" || got[2] != "4p" {
			t.Fatalf("got %v", got)
		}
	})

	rejected := []struct {
		name  string
		kind  hand.MeldKind
		tiles []tile.Tile
	}{
		{"an unknown kind", hand.MeldKind(9), mt.Tiles("1m 2m 3m")},
		{"the wrong tile count", hand.Chi, mt.Tiles("1m 2m")},
		{"a quad with three tiles", hand.Ankan, mt.Tiles("1m 1m 1m")},
		{"a sequence across suits", hand.Chi, mt.Tiles("1m 2m 3p")},
		{"a non-consecutive sequence", hand.Chi, mt.Tiles("1m 2m 4m")},
		{"a sequence of honors", hand.Chi, mt.Tiles("1z 2z 3z")},
		{"a triplet of different tiles", hand.Pon, mt.Tiles("1m 2m 3m")},
		{"a quad of different tiles", hand.Minkan, mt.Tiles("1m 1m 1m 2m")},
		{"an invalid tile", hand.Pon, []tile.Tile{tile.M1, tile.M1, 0}},
	}
	for _, c := range rejected {
		t.Run("rejects "+c.name, func(t *testing.T) {
			if _, err := hand.NewMeld(c.kind, c.tiles); !errors.Is(err, hand.ErrInvalidMeld) {
				t.Fatalf("err = %v", err)
			}
		})
	}
	t.Run("MustMeld panics on an invalid meld", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("did not panic")
			}
		}()
		hand.MustMeld(hand.Chi, mt.Tiles("1m 2m"))
	})
}

func TestMeldKind(t *testing.T) {
	cases := []struct {
		kind   hand.MeldKind
		name   string
		called bool
		kan    bool
	}{
		{hand.Chi, "chi", true, false},
		{hand.Pon, "pon", true, false},
		{hand.Minkan, "minkan", true, true},
		{hand.Ankan, "ankan", false, true},
	}
	for _, c := range cases {
		if c.kind.String() != c.name || c.kind.IsCalled() != c.called || c.kind.IsKan() != c.kan {
			t.Errorf("%v: %q called=%v kan=%v", c.kind, c.kind.String(), c.kind.IsCalled(), c.kind.IsKan())
		}
		parsed, err := hand.ParseMeldKind(c.name)
		if err != nil || parsed != c.kind {
			t.Errorf("ParseMeldKind(%q) = %v, %v", c.name, parsed, err)
		}
	}
	if _, err := hand.ParseMeldKind("kan"); !errors.Is(err, hand.ErrInvalidMeld) {
		t.Errorf("err = %v", err)
	}
	if got := hand.MeldKind(9).String(); got != "MeldKind(9)" {
		t.Errorf("got %q", got)
	}
}

func TestMeldQueries(t *testing.T) {
	if !mt.Chi("1p 2p 3p").IsCalled() || !mt.Pon("1z 1z 1z").IsCalled() || !mt.Minkan("1z 1z 1z 1z").IsCalled() {
		t.Error("chi/pon/minkan are calls")
	}
	if mt.Ankan("1z 1z 1z 1z").IsCalled() {
		t.Error("ankan is not a call")
	}
	if mt.Chi("1p 2p 3p").IsKan() || mt.Pon("1z 1z 1z").IsKan() {
		t.Error("chi/pon are not kan")
	}
	if !mt.Minkan("1z 1z 1z 1z").IsKan() || !mt.Ankan("1z 1z 1z 1z").IsKan() {
		t.Error("minkan/ankan are kan")
	}
	if got := mt.Chi("1m 2m 3m").String(); got != "chi[1m 2m 3m]" {
		t.Errorf("String() = %q", got)
	}
	m := mt.Pon("1z 1z 1z")
	m.Tiles()[0] = tile.M1
	if m.Tiles()[0] != tile.East {
		t.Error("Tiles() must return a copy")
	}
}

func TestMeldEquality(t *testing.T) {
	a, b := mt.Pon("1z 1z 1z"), mt.Pon("1z 1z 1z")
	if a != b {
		t.Error("same kind and tiles must be equal")
	}
	if mt.Pon("5p 5p 5p") == mt.Minkan("5p 5p 5p 5p") {
		t.Error("different kinds must not be equal")
	}
	if mt.Chi("1m 2m 3m") == mt.Chi("3m 2m 1m") {
		t.Error("tile order is part of identity")
	}
	seen := map[hand.Meld]bool{mt.Pon("1z 1z 1z"): true}
	if !seen[mt.Pon("1z 1z 1z")] {
		t.Error("usable as a map key")
	}
}
