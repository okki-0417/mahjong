package winning

import (
	"strings"

	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

// FuSourceKind names where fu came from.
type FuSourceKind string

// Every source of fu.
const (
	FuBase       FuSourceKind = "base"
	FuMenzenRon  FuSourceKind = "menzen_ron"
	FuTsumo      FuSourceKind = "tsumo"
	FuPair       FuSourceKind = "pair"
	FuWait       FuSourceKind = "wait"
	FuMentsu     FuSourceKind = "mentsu"
	FuOpenPinfu  FuSourceKind = "open_pinfu"
	FuChiitoitsu FuSourceKind = "chiitoitsu"
	FuPinfu      FuSourceKind = "pinfu"
)

// FuSource is one contribution to the fu: the base, a set, the pair, the
// wait, or a fixed total. Tiles names the tiles behind it when there are
// any.
type FuSource struct {
	Kind  FuSourceKind
	Label string
	Fu    int
	Tiles []tile.Tile
}

const (
	fuBase        = 20
	fuChiitoitsu  = 25
	fuMenzenRon   = 10
	fuOpenPinfu   = 10
	fuTsumo       = 2
	fuYakuhaiPair = 2
	fuWait        = 2
	fuPinfuTsumo  = 20
	fuPinfuRon    = 30
	fuCeiling     = 10
)

var fuWaitLabels = map[WaitKind]string{Kanchan: "嵌張待ち", Penchan: "辺張待ち", Tanki: "単騎待ち"}

// Fu is the fu of one form in one situation: the sources and the total
// rounded up to ten. Pinfu is fixed at 20 (tsumo) or 30 (ron), chiitoitsu at
// 25, and kokushi has none.
type Fu struct {
	form    *Form
	sources []FuSource
	fixed   bool
}

func fuOf(f *Form, s Situation, rs ruleset.RuleSet) Fu {
	fu := Fu{form: f}
	switch {
	case f.kind == Chiitoitsu:
		fu.fixed = true
		fu.sources = []FuSource{{Kind: FuChiitoitsu, Label: "七対子（25符固定）", Fu: fuChiitoitsu}}
	case f.kind == Kokushi:
	case f.IsMenzen() && pinfu(f, s, rs):
		fu.fixed = true
		if s.IsTsumo() {
			fu.sources = []FuSource{{Kind: FuPinfu, Label: "平和ツモ（20符固定）", Fu: fuPinfuTsumo}}
		} else {
			fu.sources = []FuSource{{Kind: FuPinfu, Label: "平和ロン（30符固定）", Fu: fuPinfuRon}}
		}
	default:
		fu.sources = collectFuSources(f, s)
	}
	return fu
}

// Form returns the reading the fu were counted on.
func (f Fu) Form() *Form {
	return f.form
}

// Total returns the fu rounded up to the next ten. Fixed fu (pinfu,
// chiitoitsu) are not rounded.
func (f Fu) Total() int {
	if f.fixed {
		return f.Subtotal()
	}
	return (f.Subtotal() + fuCeiling - 1) / fuCeiling * fuCeiling
}

// Subtotal returns the fu before rounding.
func (f Fu) Subtotal() int {
	total := 0
	for _, s := range f.sources {
		total += s.Fu
	}
	return total
}

// Sources lists each contribution.
func (f Fu) Sources() []FuSource {
	out := make([]FuSource, len(f.sources))
	for i, s := range f.sources {
		out[i] = s
		out[i].Tiles = append([]tile.Tile(nil), s.Tiles...)
	}
	return out
}

func collectFuSources(f *Form, s Situation) []FuSource {
	sources := []FuSource{{Kind: FuBase, Label: "副底", Fu: fuBase}}
	if f.IsMenzen() && s.IsRon() {
		sources = append(sources, FuSource{Kind: FuMenzenRon, Label: "門前加符（ロン）", Fu: fuMenzenRon})
	}
	if s.IsTsumo() {
		sources = append(sources, FuSource{Kind: FuTsumo, Label: "ツモ", Fu: fuTsumo})
	}
	if src, ok := pairFuSource(f, s); ok {
		sources = append(sources, src)
	}
	if label, ok := fuWaitLabels[f.waitKind]; ok {
		sources = append(sources, FuSource{Kind: FuWait, Label: label, Fu: fuWait})
	}
	for _, m := range f.mentsu {
		if m.kind != Shuntsu {
			sources = append(sources, mentsuFuSource(m))
		}
	}
	// A called hand whose only fu is the base (an open pinfu shape) is fixed
	// at 30 on ron rather than 20.
	if s.IsRon() && !f.IsMenzen() && sumFu(sources) == fuBase {
		sources = append(sources, FuSource{Kind: FuOpenPinfu, Label: "食い平和形（30符固定）", Fu: fuOpenPinfu})
	}
	return sources
}

func sumFu(sources []FuSource) int {
	total := 0
	for _, s := range sources {
		total += s.Fu
	}
	return total
}

// pairFuSource is 2 for a yakuhai pair. A pair that is both the round and
// seat wind still counts once.
func pairFuSource(f *Form, s Situation) (FuSource, bool) {
	pair := f.pairTile
	if !pair.IsHonor() {
		return FuSource{}, false
	}
	var parts []string
	if pair.IsDragon() {
		parts = append(parts, "三元牌")
	}
	if s.IsRoundWind(pair) {
		parts = append(parts, "場風")
	}
	if s.IsSeatWind(pair) {
		parts = append(parts, "自風")
	}
	if len(parts) == 0 {
		return FuSource{}, false
	}
	return FuSource{
		Kind:  FuPair,
		Label: "雀頭（" + strings.Join(parts, "・") + "）",
		Fu:    fuYakuhaiPair,
		Tiles: []tile.Tile{pair},
	}, true
}

func mentsuFuSource(m Mentsu) FuSource {
	var kindLabel string
	switch {
	case m.kind == Koutsu && m.open:
		kindLabel = "明刻"
	case m.kind == Koutsu:
		kindLabel = "暗刻"
	case m.open:
		kindLabel = "明槓"
	default:
		kindLabel = "暗槓"
	}
	tileLabel := "中張牌"
	if m.representative().IsTerminalOrHonor() {
		tileLabel = "么九牌"
	}
	return FuSource{
		Kind:  FuMentsu,
		Label: kindLabel + "（" + tileLabel + "）",
		Fu:    m.fu(),
		Tiles: m.Tiles(),
	}
}
