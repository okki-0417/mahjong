package hand

import (
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/tile"
)

func tilesOf(t *testing.T, labels string) []tile.Tile {
	t.Helper()
	fields := strings.Fields(labels)
	out := make([]tile.Tile, 0, len(fields))
	for _, l := range fields {
		out = append(out, tile.MustParse(l))
	}
	return out
}

func meldOf(t *testing.T, kind MeldKind, labels string) Meld {
	t.Helper()
	return MustMeld(kind, tilesOf(t, labels))
}

func TestChiitoitsuShanten(t *testing.T) {
	cases := []struct {
		name   string
		closed string
		want   Shanten
	}{
		{"seven pairs (14 tiles) is agari", "1m 1m 3m 3m 5p 5p 7p 7p 9s 9s 1z 1z 3z 3z", -1},
		{"six pairs and a single is tenpai", "1m 1m 3m 3m 5p 5p 7p 7p 9s 9s 1z 1z 3z", 0},
		{"five pairs and three singles is 1", "1m 1m 3m 3m 5p 5p 7p 7p 9s 9s 1z 3z 5z", 1},
		{"four pairs and five singles is 2", "1m 1m 3m 3m 5p 5p 7p 7p 9s 1z 3z 5z 7z", 2},
		{"no pairs over 13 kinds is 6", "1m 2m 3m 4m 5p 6p 7p 8p 1s 2s 3s 4s 5s", 6},
		{"fewer than seven kinds is a penalty: four pairs over five kinds is 4", "1m 1m 1m 2m 2m 2m 3m 3m 3m 4m 4m 4m 5m", 4},
		{"a red five pairs with a plain five", "1m 1m 0m 5m 3p 3p 5p 5p 7s 7s 1z 1z 3z", 0},
		{"honors only, six pairs and a single is tenpai", "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := chiitoitsuShanten(tilesOf(t, c.closed)); got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

func TestKokushiShanten(t *testing.T) {
	cases := []struct {
		name   string
		closed string
		want   Shanten
	}{
		{"thirteen kinds is the thirteen-sided tenpai", "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", 0},
		{"thirteen kinds plus a pair is agari", "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z 1m", -1},
		{"twelve kinds with a pair is a single wait tenpai", "1m 9m 1p 9p 1s 9s 1z 1z 2z 3z 4z 5z 6z", 0},
		{"twelve kinds without a pair is 1", "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 5p", 1},
		{"eleven kinds with a pair is 1", "1m 9m 1p 9p 1s 9s 1z 1z 2z 3z 4z 5z 5p", 1},
		{"no terminals or honors is 13", "2m 3m 4m 5m 6m 2p 3p 4p 5p 2s 3s 4s 5s", 13},
		{"a red five is not a terminal", "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 0m", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := kokushiShanten(tilesOf(t, c.closed)); got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

type meldSpec struct {
	kind   MeldKind
	labels string
}

func TestShantenOf(t *testing.T) {
	cases := []struct {
		name   string
		closed string
		melds  []meldSpec
		want   Shanten
	}{
		{name: "[agari] standard form", closed: "1m 2m 3m 4p 5p 6p 7s 8s 9s 9m 9m 1z 1z 1z", want: -1},
		{name: "[agari] all triplets", closed: "2m 2m 2m 5p 5p 5p 7s 7s 7s 9s 9s 9s 4z 4z", want: -1},
		{name: "[agari] chiitoitsu", closed: "1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z 7z", want: -1},
		{name: "[agari] kokushi", closed: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z 1m", want: -1},
		{name: "[agari] kokushi thirteen-sided", closed: "1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", want: -1},

		{name: "[tenpai] tanki", closed: "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 1z 1z 9m", want: 0},
		{name: "[tenpai] ryanmen", closed: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 2s 5z 5z", want: 0},
		{name: "[tenpai] kanchan", closed: "1m 2m 3m 4m 5m 6m 7p 8p 9p 1s 3s 5z 5z", want: 0},
		{name: "[tenpai] penchan", closed: "1m 2m 3m 4m 5m 6m 7p 8p 9p 8s 9s 5z 5z", want: 0},
		{name: "[tenpai] shanpon", closed: "1m 2m 3m 4m 5m 6m 7p 8p 9p 5z 5z 9s 9s", want: 0},
		{name: "[tenpai] chiitoitsu", closed: "1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 8s 7z", want: 0},
		{name: "[tenpai] kokushi thirteen-sided", closed: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z", want: 0},
		{name: "[tenpai] kokushi tanki", closed: "1m 1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z", want: 0},

		{name: "[1] standard form", closed: "1m 2m 3m 4m 5m 6m 7p 8p 1s 2s 5z 5z 9s", want: 1},
		{name: "[1] chiitoitsu ahead of standard form", closed: "1m 1m 4p 4p 7s 7s 2m 2m 5p 5p 8s 9s 7z", want: 1},
		{name: "[1] kokushi ahead of standard form", closed: "1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 5p", want: 1},

		{name: "[2] three sets and loose honors", closed: "1m 2m 3m 4p 5p 6p 7s 8s 9s 1z 2z 3z 4z", want: 2},
		{name: "[3] M=2 T=1", closed: "1m 2m 3m 4m 5m 6m 7p 8p 1s 4s 7s 1z 2z", want: 3},
		{name: "[4] M=1 T=2", closed: "1m 2m 3m 4p 5p 7s 8s 1z 2z 3z 4z 5z 6z", want: 4},
		{name: "[5] M=0 T=3", closed: "1m 2m 4p 5p 7s 8s 1z 2z 3z 4z 5z 6z 7z", want: 5},
		{name: "[6] thirteen isolated tiles: standard 8, chiitoitsu and kokushi 6", closed: "1m 4m 7m 1p 4p 7p 1s 4s 7s 1z 3z 5z 7z", want: 6},

		{name: "[1 meld] chi and a tanki tenpai", closed: "4p 5p 6p 7s 8s 9s 1z 1z 1z 9m", melds: []meldSpec{{Chi, "1m 2m 3m"}}, want: 0},
		{name: "[1 meld] pon and 1-shanten", closed: "4p 5p 6p 7s 8s 9s 1z 1z 2z 3z", melds: []meldSpec{{Pon, "5z 5z 5z"}}, want: 1},
		{name: "[2 melds] tenpai", closed: "7p 8p 9p 1s 2s 3s 4z", melds: []meldSpec{{Chi, "1m 2m 3m"}, {Chi, "4m 5m 6m"}}, want: 0},
		{name: "[3 melds] tenpai", closed: "1m 2m 3m 4z", melds: []meldSpec{{Chi, "1p 2p 3p"}, {Chi, "4p 5p 6p"}, {Pon, "5z 5z 5z"}}, want: 0},
		{name: "[4 melds] tanki tenpai", closed: "4z", melds: []meldSpec{{Chi, "1m 2m 3m"}, {Chi, "4m 5m 6m"}, {Chi, "1p 2p 3p"}, {Pon, "5z 5z 5z"}}, want: 0},
		{name: "[ankan] counts as a set", closed: "1m 2m 3m 4p 5p 6p 7s 8s 9s 4z", melds: []meldSpec{{Ankan, "5m 5m 5m 5m"}}, want: 0},

		{name: "[red five] 0p joins a set as 5p", closed: "1m 2m 3m 4p 0p 6p 7s 8s 9s 5m 5m 1z 1z", want: 0},

		{name: "[multiple readings] pure chuuren shape with four of a kind is tenpai", closed: "1m 1m 1m 1m 2m 3m 4m 5m 6m 7m 8m 9m 9m", want: 0},
		{name: "[multiple readings] 1m×3 as triplet, pair, or sequence", closed: "1m 1m 1m 2m 3m 4p 5p 6p 7s 8s 9s 4z 5z", want: 1},

		{name: "[cap regression] 14 tiles with M=4 T=1 P=0 is tenpai, not agari", closed: "1m 2m 3m 4m 5m 6m 7p 8p 9p 4s 5s 5z 5z 5z", want: 0},

		{name: "[edge] honors only, six pairs", closed: "1z 1z 2z 2z 3z 3z 4z 4z 5z 5z 6z 6z 7z", want: 0},
		{name: "[edge] numbers only, kokushi irrelevant", closed: "2m 3m 4m 2m 3m 4m 2p 3p 4p 1s 1s 5s 6s", want: 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var melds []Meld
			for _, m := range c.melds {
				melds = append(melds, meldOf(t, m.kind, m.labels))
			}
			if got := ShantenOf(tilesOf(t, c.closed), melds); got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

func TestShantenValue(t *testing.T) {
	if !Tenpai.IsTenpai() || Agari.IsTenpai() || Shanten(1).IsTenpai() {
		t.Error("IsTenpai")
	}
	if Agari >= Tenpai || Tenpai >= Shanten(1) {
		t.Error("ordering")
	}
}
