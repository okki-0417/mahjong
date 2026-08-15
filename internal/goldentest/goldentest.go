// Package goldentest reads the JSON Lines fixtures under testdata/ that were
// recorded from the Ruby implementation this engine is ported from.
package goldentest

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

// Meld is a meld as recorded in a fixture.
type Meld struct {
	Kind  string   `json:"kind"`
	Tiles []string `json:"tiles"`
}

// Hand builds the hand from recorded closed tiles and melds.
func Hand(t *testing.T, closed []string, melds []Meld) hand.Hand {
	t.Helper()
	closedTiles, err := tile.ParseAll(closed)
	if err != nil {
		t.Fatal(err)
	}
	built := make([]hand.Meld, 0, len(melds))
	for _, m := range melds {
		kind, err := hand.ParseMeldKind(m.Kind)
		if err != nil {
			t.Fatal(err)
		}
		tiles, err := tile.ParseAll(m.Tiles)
		if err != nil {
			t.Fatal(err)
		}
		built = append(built, hand.MustMeld(kind, tiles))
	}
	return hand.Must(closedTiles, built)
}

// Each decodes every line of testdata/<name> into a fresh T and calls fn
// with its 1-based line number.
func Each[T any](t *testing.T, name string, fn func(line int, record T)) {
	t.Helper()
	_, here, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(here), "..", "..", "testdata", name)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 64<<20)
	line := 0
	for scanner.Scan() {
		line++
		var record T
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("%s line %d: %v", name, line, err)
		}
		fn(line, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if line == 0 {
		t.Fatalf("%s: no records", name)
	}
}

// SameLabels compares label lists, treating nil and empty as equal.
func SameLabels(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
