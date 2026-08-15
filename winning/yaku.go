package winning

import (
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

// YakuID identifies a yaku. The string form is stable and safe to serialize.
type YakuID string

// Every yaku the engine scores.
const (
	YakuRiichi                YakuID = "riichi"
	YakuDoubleRiichi          YakuID = "double_riichi"
	YakuIppatsu               YakuID = "ippatsu"
	YakuMenzenTsumo           YakuID = "menzen_tsumo"
	YakuHaitei                YakuID = "haitei"
	YakuHoutei                YakuID = "houtei"
	YakuRinshan               YakuID = "rinshan"
	YakuChankan               YakuID = "chankan"
	YakuHaku                  YakuID = "yakuhai_haku"
	YakuHatsu                 YakuID = "yakuhai_hatsu"
	YakuChun                  YakuID = "yakuhai_chun"
	YakuRoundWind             YakuID = "yakuhai_round_wind"
	YakuSeatWind              YakuID = "yakuhai_seat_wind"
	YakuTanyao                YakuID = "tanyao"
	YakuPinfu                 YakuID = "pinfu"
	YakuIipeikou              YakuID = "iipeikou"
	YakuSanshokuDoujun        YakuID = "sanshoku_doujun"
	YakuIttsu                 YakuID = "ittsu"
	YakuToitoi                YakuID = "toitoi"
	YakuSanankou              YakuID = "sanankou"
	YakuChanta                YakuID = "chanta"
	YakuJunchan               YakuID = "junchan"
	YakuHonroutou             YakuID = "honroutou"
	YakuHonitsu               YakuID = "honitsu"
	YakuChinitsu              YakuID = "chinitsu"
	YakuRyanpeikou            YakuID = "ryanpeikou"
	YakuSanshokuDoukou        YakuID = "sanshoku_doukou"
	YakuSankantsu             YakuID = "sankantsu"
	YakuShousangen            YakuID = "shousangen"
	YakuChiitoitsu            YakuID = "chiitoitsu"
	YakuKokushimusou          YakuID = "kokushimusou"
	YakuSuuankou              YakuID = "suuankou"
	YakuDaisangen             YakuID = "daisangen"
	YakuTsuuiisou             YakuID = "tsuuiisou"
	YakuRyuuiisou             YakuID = "ryuuiisou"
	YakuChinroutou            YakuID = "chinroutou"
	YakuChuurenpoutou         YakuID = "chuurenpoutou"
	YakuSuukantsu             YakuID = "suukantsu"
	YakuShousuushi            YakuID = "shousuushi"
	YakuDaisuushi             YakuID = "daisuushi"
	YakuTenhou                YakuID = "tenhou"
	YakuChiihou               YakuID = "chiihou"
	YakuKokushimusouJuusanmen YakuID = "kokushimusou_juusanmen"
	YakuSuuankouTanki         YakuID = "suuankou_tanki"
	YakuJunseiChuurenpoutou   YakuID = "junsei_chuurenpoutou"
)

// Yaku is a yaku that applies to a winning form, with the han it is worth
// there (0 for a yakuman; see Yakuman).
type Yaku struct {
	ID      YakuID
	Name    string
	Han     int
	Yakuman int
}

// IsYakuman reports whether the yaku is a yakuman.
func (y Yaku) IsYakuman() bool {
	return y.Yakuman > 0
}

type yakuDef struct {
	id         YakuID
	name       string
	hanClosed  int
	hanOpen    int
	yakuman    int
	menzenOnly func(ruleset.RuleSet) bool
	supersedes []YakuID
	satisfied  func(*Form, Situation, ruleset.RuleSet) bool
}

var closedOnly = func(ruleset.RuleSet) bool { return true }

func (d yakuDef) of(f *Form, s Situation, rs ruleset.RuleSet) (Yaku, bool) {
	if d.menzenOnly != nil && d.menzenOnly(rs) && !f.IsMenzen() {
		return Yaku{}, false
	}
	if !d.satisfied(f, s, rs) {
		return Yaku{}, false
	}
	y := Yaku{ID: d.id, Name: d.name, Yakuman: d.yakuman}
	if d.yakuman == 0 {
		if f.IsMenzen() {
			y.Han = d.hanClosed
		} else {
			y.Han = d.hanOpen
		}
	}
	return y, true
}

// appliedYaku resolves the yaku of one form: every matching yaku, except
// that any yakuman drops all regular yaku, and a yaku that supersedes
// another (junsei chuurenpoutou over chuurenpoutou) drops it.
func appliedYaku(f *Form, s Situation, rs ruleset.RuleSet) []Yaku {
	var matched []Yaku
	var yakuman []Yaku
	for _, d := range yakuCatalog {
		if y, ok := d.of(f, s, rs); ok {
			matched = append(matched, y)
			if y.IsYakuman() {
				yakuman = append(yakuman, y)
			}
		}
	}
	if len(yakuman) == 0 {
		return matched
	}
	superseded := map[YakuID]bool{}
	for _, y := range yakuman {
		for _, id := range supersededBy(y.ID) {
			superseded[id] = true
		}
	}
	kept := make([]Yaku, 0, len(yakuman))
	for _, y := range yakuman {
		if !superseded[y.ID] {
			kept = append(kept, y)
		}
	}
	return kept
}

func supersededBy(id YakuID) []YakuID {
	for _, d := range yakuCatalog {
		if d.id == id {
			return d.supersedes
		}
	}
	return nil
}

func hanTotal(yakus []Yaku) int {
	total := 0
	for _, y := range yakus {
		if y.IsYakuman() {
			return 0
		}
		total += y.Han
	}
	return total
}

func yakumanTotal(yakus []Yaku) int {
	total := 0
	for _, y := range yakus {
		total += y.Yakuman
	}
	return total
}

func countMentsu(f *Form, pred func(Mentsu) bool) int {
	n := 0
	for _, m := range f.mentsu {
		if pred(m) {
			n++
		}
	}
	return n
}

func anyMentsu(f *Form, pred func(Mentsu) bool) bool {
	return countMentsu(f, pred) > 0
}

func allMentsu(f *Form, pred func(Mentsu) bool) bool {
	return countMentsu(f, pred) == len(f.mentsu)
}

func hasTripletOf(f *Form, t tile.Tile) bool {
	return anyMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && m.representative() == t })
}

// shuntsuStarts groups the starting numbers of sequences by numeric suit.
func shuntsuStarts(f *Form) map[tile.Suit][]int {
	return startsBy(f, func(m Mentsu) bool { return m.kind == Shuntsu })
}

func tripletNumbers(f *Form) map[tile.Suit][]int {
	return startsBy(f, func(m Mentsu) bool { return m.IsTriplet() })
}

func startsBy(f *Form, pred func(Mentsu) bool) map[tile.Suit][]int {
	out := map[tile.Suit][]int{}
	for _, m := range f.mentsu {
		if pred(m) && m.suit().IsNumeric() {
			out[m.suit()] = append(out[m.suit()], m.number())
		}
	}
	return out
}

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func sanshoku(byNumber map[tile.Suit][]int) bool {
	for _, n := range byNumber[tile.Man] {
		if containsInt(byNumber[tile.Pin], n) && containsInt(byNumber[tile.Sou], n) {
			return true
		}
	}
	return false
}

// shuntsuSignatureCounts counts each distinct sequence (suit, start);
// iipeikou needs exactly one to appear twice, ryanpeikou two.
func shuntsuSignatureCounts(f *Form) map[[2]int]int {
	counts := map[[2]int]int{}
	for _, m := range f.mentsu {
		if m.kind == Shuntsu {
			counts[[2]int{int(m.suit()), m.number()}]++
		}
	}
	return counts
}

func numberCounts(f *Form) map[int]int {
	counts := map[int]int{}
	for _, t := range f.allTiles() {
		counts[t.EffectiveNumber()]++
	}
	return counts
}

var chuurenShape = map[int]int{1: 3, 2: 1, 3: 1, 4: 1, 5: 1, 6: 1, 7: 1, 8: 1, 9: 3}

func chuurenpoutou(f *Form, _ Situation, _ ruleset.RuleSet) bool {
	if f.kind != Standard || f.hasHonor() {
		return false
	}
	if !f.oneNumericSuit() {
		return false
	}
	counts := numberCounts(f)
	for n, need := range chuurenShape {
		if counts[n] < need {
			return false
		}
	}
	return true
}

func suuankou(f *Form, _ Situation, _ ruleset.RuleSet) bool {
	return f.kind == Standard && countMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && !m.open }) == 4
}

var greenTiles = map[tile.Tile]bool{tile.S2: true, tile.S3: true, tile.S4: true, tile.S6: true, tile.S8: true, tile.Hatsu: true}

func pinfu(f *Form, s Situation, _ ruleset.RuleSet) bool {
	return f.kind == Standard &&
		f.waitKind == Ryanmen &&
		allMentsu(f, func(m Mentsu) bool { return m.kind == Shuntsu }) &&
		!s.IsYakuhai(f.pairTile)
}

func yakuhaiOf(t tile.Tile) func(*Form, Situation, ruleset.RuleSet) bool {
	return func(f *Form, _ Situation, _ ruleset.RuleSet) bool {
		return f.kind == Standard && hasTripletOf(f, t)
	}
}

func situational(pick func(Situation) bool) func(*Form, Situation, ruleset.RuleSet) bool {
	return func(_ *Form, s Situation, _ ruleset.RuleSet) bool { return pick(s) }
}

func standardOnly(pred func(*Form) bool) func(*Form, Situation, ruleset.RuleSet) bool {
	return func(f *Form, _ Situation, _ ruleset.RuleSet) bool { return f.kind == Standard && pred(f) }
}

func standardOrChiitoitsu(pred func(*Form) bool) func(*Form, Situation, ruleset.RuleSet) bool {
	return func(f *Form, _ Situation, _ ruleset.RuleSet) bool {
		return (f.kind == Standard || f.kind == Chiitoitsu) && pred(f)
	}
}

func notKokushi(pred func(*Form) bool) func(*Form, Situation, ruleset.RuleSet) bool {
	return func(f *Form, _ Situation, _ ruleset.RuleSet) bool { return f.kind != Kokushi && pred(f) }
}

// yakuCatalog lists every yaku in evaluation order. Its order is the order
// yaku are reported in.
var yakuCatalog = []yakuDef{
	{id: YakuRiichi, name: "立直", hanClosed: 1, menzenOnly: closedOnly,
		satisfied: situational(func(s Situation) bool { return s.Riichi && !s.DoubleRiichi })},
	{id: YakuDoubleRiichi, name: "ダブル立直", hanClosed: 2, menzenOnly: closedOnly,
		satisfied: situational(func(s Situation) bool { return s.DoubleRiichi })},
	{id: YakuIppatsu, name: "一発", hanClosed: 1, menzenOnly: closedOnly,
		satisfied: situational(func(s Situation) bool { return s.Ippatsu })},
	{id: YakuMenzenTsumo, name: "門前清自摸和", hanClosed: 1, menzenOnly: closedOnly,
		satisfied: situational(Situation.IsTsumo)},
	{id: YakuHaitei, name: "海底摸月", hanClosed: 1, hanOpen: 1,
		satisfied: situational(func(s Situation) bool { return s.Haitei })},
	{id: YakuHoutei, name: "河底撈魚", hanClosed: 1, hanOpen: 1,
		satisfied: situational(func(s Situation) bool { return s.Houtei })},
	{id: YakuRinshan, name: "嶺上開花", hanClosed: 1, hanOpen: 1,
		satisfied: situational(func(s Situation) bool { return s.Rinshan })},
	{id: YakuChankan, name: "槍槓", hanClosed: 1, hanOpen: 1,
		satisfied: situational(func(s Situation) bool { return s.Chankan })},

	{id: YakuHaku, name: "役牌（白）", hanClosed: 1, hanOpen: 1, satisfied: yakuhaiOf(tile.Haku)},
	{id: YakuHatsu, name: "役牌（發）", hanClosed: 1, hanOpen: 1, satisfied: yakuhaiOf(tile.Hatsu)},
	{id: YakuChun, name: "役牌（中）", hanClosed: 1, hanOpen: 1, satisfied: yakuhaiOf(tile.Chun)},
	{id: YakuRoundWind, name: "場風", hanClosed: 1, hanOpen: 1,
		satisfied: func(f *Form, s Situation, _ ruleset.RuleSet) bool {
			return f.kind == Standard && hasTripletOf(f, s.RoundWind.Tile())
		}},
	{id: YakuSeatWind, name: "自風", hanClosed: 1, hanOpen: 1,
		satisfied: func(f *Form, s Situation, _ ruleset.RuleSet) bool {
			return f.kind == Standard && hasTripletOf(f, s.SeatWind.Tile())
		}},

	{id: YakuTanyao, name: "断么九", hanClosed: 1, hanOpen: 1,
		menzenOnly: func(rs ruleset.RuleSet) bool { return !rs.Kuitan() },
		satisfied: func(f *Form, _ Situation, _ ruleset.RuleSet) bool {
			return f.allTilesSatisfy(func(t tile.Tile) bool { return !t.IsTerminalOrHonor() })
		}},
	{id: YakuPinfu, name: "平和", hanClosed: 1, menzenOnly: closedOnly, satisfied: pinfu},
	{id: YakuIipeikou, name: "一盃口", hanClosed: 1, menzenOnly: closedOnly,
		satisfied: standardOnly(func(f *Form) bool {
			pairs := 0
			for _, c := range shuntsuSignatureCounts(f) {
				if c >= 2 {
					pairs++
				}
			}
			return pairs == 1
		})},
	{id: YakuSanshokuDoujun, name: "三色同順", hanClosed: 2, hanOpen: 1,
		satisfied: standardOnly(func(f *Form) bool { return sanshoku(shuntsuStarts(f)) })},
	{id: YakuIttsu, name: "一気通貫", hanClosed: 2, hanOpen: 1,
		satisfied: standardOnly(func(f *Form) bool {
			for _, starts := range shuntsuStarts(f) {
				if containsInt(starts, 1) && containsInt(starts, 4) && containsInt(starts, 7) {
					return true
				}
			}
			return false
		})},
	{id: YakuToitoi, name: "対々和", hanClosed: 2, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool { return allMentsu(f, Mentsu.IsTriplet) })},
	{id: YakuSanankou, name: "三暗刻", hanClosed: 2, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool {
			return countMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && !m.open }) >= 3
		})},
	{id: YakuChanta, name: "混全帯么九", hanClosed: 2, hanOpen: 1,
		satisfied: standardOnly(func(f *Form) bool {
			return f.hasHonor() &&
				f.pairTile.IsTerminalOrHonor() &&
				allMentsu(f, Mentsu.containsTerminalOrHonor) &&
				anyMentsu(f, func(m Mentsu) bool { return m.kind == Shuntsu })
		})},
	{id: YakuJunchan, name: "純全帯么九", hanClosed: 3, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool {
			return !f.hasHonor() &&
				f.pairTile.IsTerminal() &&
				allMentsu(f, Mentsu.containsTerminal) &&
				anyMentsu(f, func(m Mentsu) bool { return m.kind == Shuntsu })
		})},
	{id: YakuHonroutou, name: "混老頭", hanClosed: 2, hanOpen: 2,
		satisfied: standardOrChiitoitsu(func(f *Form) bool {
			return f.hasHonor() && f.allTilesSatisfy(tile.Tile.IsTerminalOrHonor)
		})},
	{id: YakuHonitsu, name: "混一色", hanClosed: 3, hanOpen: 2,
		satisfied: standardOrChiitoitsu(func(f *Form) bool { return f.hasHonor() && f.oneNumericSuit() })},
	{id: YakuChinitsu, name: "清一色", hanClosed: 6, hanOpen: 5,
		satisfied: standardOrChiitoitsu(func(f *Form) bool { return !f.hasHonor() && f.oneNumericSuit() })},
	{id: YakuRyanpeikou, name: "二盃口", hanClosed: 3, menzenOnly: closedOnly,
		satisfied: standardOnly(func(f *Form) bool {
			counts := shuntsuSignatureCounts(f)
			if len(counts) != 2 {
				return false
			}
			for _, c := range counts {
				if c != 2 {
					return false
				}
			}
			return true
		})},
	{id: YakuSanshokuDoukou, name: "三色同刻", hanClosed: 2, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool { return sanshoku(tripletNumbers(f)) })},
	{id: YakuSankantsu, name: "三槓子", hanClosed: 2, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool {
			return countMentsu(f, func(m Mentsu) bool { return m.kind == Kantsu }) >= 3
		})},
	{id: YakuShousangen, name: "小三元", hanClosed: 2, hanOpen: 2,
		satisfied: standardOnly(func(f *Form) bool {
			return f.pairTile.IsDragon() &&
				countMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && m.representative().IsDragon() }) == 2
		})},

	{id: YakuChiitoitsu, name: "七対子", hanClosed: 2, menzenOnly: closedOnly,
		satisfied: func(f *Form, _ Situation, _ ruleset.RuleSet) bool { return f.kind == Chiitoitsu }},

	{id: YakuKokushimusou, name: "国士無双", yakuman: 1, menzenOnly: closedOnly,
		satisfied: func(f *Form, _ Situation, _ ruleset.RuleSet) bool { return f.kind == Kokushi }},
	{id: YakuSuuankou, name: "四暗刻", yakuman: 1, menzenOnly: closedOnly, satisfied: suuankou},
	{id: YakuDaisangen, name: "大三元", yakuman: 1,
		satisfied: standardOnly(func(f *Form) bool {
			return hasTripletOf(f, tile.Haku) && hasTripletOf(f, tile.Hatsu) && hasTripletOf(f, tile.Chun)
		})},
	{id: YakuTsuuiisou, name: "字一色", yakuman: 1,
		satisfied: notKokushi(func(f *Form) bool { return f.allTilesSatisfy(tile.Tile.IsHonor) })},
	{id: YakuRyuuiisou, name: "緑一色", yakuman: 1,
		satisfied: notKokushi(func(f *Form) bool {
			return f.allTilesSatisfy(func(t tile.Tile) bool { return greenTiles[t.Kind()] })
		})},
	{id: YakuChinroutou, name: "清老頭", yakuman: 1,
		satisfied: notKokushi(func(f *Form) bool { return f.allTilesSatisfy(tile.Tile.IsTerminal) })},
	{id: YakuChuurenpoutou, name: "九蓮宝燈", yakuman: 1, menzenOnly: closedOnly, satisfied: chuurenpoutou},
	{id: YakuSuukantsu, name: "四槓子", yakuman: 1,
		satisfied: standardOnly(func(f *Form) bool {
			return countMentsu(f, func(m Mentsu) bool { return m.kind == Kantsu }) == 4
		})},
	{id: YakuShousuushi, name: "小四喜", yakuman: 1,
		satisfied: standardOnly(func(f *Form) bool {
			return f.pairTile.IsWind() &&
				countMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && m.representative().IsWind() }) == 3
		})},
	{id: YakuDaisuushi, name: "大四喜", yakuman: 1,
		satisfied: standardOnly(func(f *Form) bool {
			return countMentsu(f, func(m Mentsu) bool { return m.IsTriplet() && m.representative().IsWind() }) == 4
		})},
	{id: YakuTenhou, name: "天和", yakuman: 1, menzenOnly: closedOnly,
		satisfied: situational(func(s Situation) bool { return s.Tenhou })},
	{id: YakuChiihou, name: "地和", yakuman: 1, menzenOnly: closedOnly,
		satisfied: situational(func(s Situation) bool { return s.Chiihou })},

	{id: YakuKokushimusouJuusanmen, name: "国士無双13面待ち", yakuman: 2, menzenOnly: closedOnly,
		supersedes: []YakuID{YakuKokushimusou},
		satisfied: func(f *Form, _ Situation, _ ruleset.RuleSet) bool {
			return f.kind == Kokushi && f.pairTile.SameKind(f.winningTile)
		}},
	{id: YakuSuuankouTanki, name: "四暗刻単騎", yakuman: 2, menzenOnly: closedOnly,
		supersedes: []YakuID{YakuSuuankou},
		satisfied: func(f *Form, s Situation, rs ruleset.RuleSet) bool {
			return suuankou(f, s, rs) && f.waitKind == Tanki
		}},
	{id: YakuJunseiChuurenpoutou, name: "純正九蓮宝燈", yakuman: 2, menzenOnly: closedOnly,
		supersedes: []YakuID{YakuChuurenpoutou},
		satisfied: func(f *Form, s Situation, rs ruleset.RuleSet) bool {
			if !chuurenpoutou(f, s, rs) {
				return false
			}
			counts := numberCounts(f)
			counts[f.winningTile.EffectiveNumber()]--
			for n := 1; n <= 9; n++ {
				if counts[n] != chuurenShape[n] {
					return false
				}
			}
			return true
		}},
}
