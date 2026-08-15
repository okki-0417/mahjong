package winning

import (
	"fmt"

	"github.com/okki-0417/mahjong/ruleset"
)

type rank struct {
	label      string
	minHan     int
	basePoints int
}

var (
	mangan = rank{"満貫", 5, 2000}
	// Mangan and above ignore fu; base points come from han alone.
	ranks = []rank{
		{"数え役満", 13, 8000},
		{"三倍満", 11, 6000},
		{"倍満", 8, 4000},
		{"跳満", 6, 3000},
		mangan,
	}
	yakumanLabels = []string{"役満", "ダブル役満", "トリプル役満", "四倍役満", "五倍役満", "六倍役満"}
)

const (
	yakumanBasePoints = 8000
	// 4 han 30 fu and 3 han 60 fu fall just short of mangan; round-up mangan
	// lifts them.
	roundUpManganBasePoints = 1920
	paymentUnit             = 100
	nonDealerCount          = 3
)

// Payments is who pays how much for a win. Ron is paid by the discarder
// alone; tsumo is split, with the dealer paying twice a non-dealer's share.
// A share that does not apply is 0.
type Payments struct {
	FromLoser     int
	FromDealer    int
	FromNonDealer int
}

// Total is the winner's income.
func (p Payments) Total() int {
	if p.FromLoser > 0 {
		return p.FromLoser
	}
	if p.FromDealer > 0 {
		return p.FromDealer + p.FromNonDealer*(nonDealerCount-1)
	}
	return p.FromNonDealer * nonDealerCount
}

func paymentsOf(basePoints int, s Situation) Payments {
	pay := func(multiplier int) int {
		return (basePoints*multiplier + paymentUnit - 1) / paymentUnit * paymentUnit
	}
	switch {
	case s.IsRon() && s.IsDealer():
		return Payments{FromLoser: pay(6)}
	case s.IsRon():
		return Payments{FromLoser: pay(4)}
	case s.IsDealer():
		return Payments{FromNonDealer: pay(2)}
	default:
		return Payments{FromDealer: pay(2), FromNonDealer: pay(1)}
	}
}

// Score is the value of a win on one reading: the yaku, han, fu, and what
// is paid. Fu are only consulted below mangan; a yakuman ignores dora.
type Score struct {
	yakus     []Yaku
	form      *Form
	fu        Fu
	situation Situation
	doraCount int
	ruleSet   ruleset.RuleSet
}

func scoreOf(yakus []Yaku, fu Fu, s Situation, doraCount int, rs ruleset.RuleSet) Score {
	return Score{yakus: yakus, form: fu.form, fu: fu, situation: s, doraCount: doraCount, ruleSet: rs}
}

// Yakus lists the yaku that scored, in catalog order.
func (s Score) Yakus() []Yaku {
	return append([]Yaku(nil), s.yakus...)
}

// Form returns the reading that was scored.
func (s Score) Form() *Form {
	return s.form
}

// IsYakuman reports whether the win is a yakuman.
func (s Score) IsYakuman() bool {
	return s.YakumanCount() > 0
}

// YakumanCount is how many yakuman the win is worth (2 for a double).
func (s Score) YakumanCount() int {
	return yakumanTotal(s.yakus)
}

// DoraCount is the dora counted into the han; 0 for a yakuman.
func (s Score) DoraCount() int {
	if s.IsYakuman() {
		return 0
	}
	return s.doraCount
}

// Han is the yaku han plus dora; 0 for a yakuman.
func (s Score) Han() int {
	if s.IsYakuman() {
		return 0
	}
	return hanTotal(s.yakus) + s.doraCount
}

// Fu is the rounded fu; 0 for a yakuman.
func (s Score) Fu() int {
	if s.IsYakuman() {
		return 0
	}
	return s.fu.Total()
}

// BasePoints is the value before the ron/tsumo multipliers.
func (s Score) BasePoints() int {
	if s.IsYakuman() {
		return yakumanBasePoints * s.YakumanCount()
	}
	if r, ok := s.rank(); ok {
		return r.basePoints
	}
	return s.rawBasePoints()
}

// RankLabel names the limit reached (満貫, 跳満, ..., 役満), or "" below
// mangan.
func (s Score) RankLabel() string {
	if s.IsYakuman() {
		n := s.YakumanCount()
		if n <= len(yakumanLabels) {
			return yakumanLabels[n-1]
		}
		return fmt.Sprintf("%d倍役満", n)
	}
	if r, ok := s.rank(); ok {
		return r.label
	}
	return ""
}

// Payments is who pays what.
func (s Score) Payments() Payments {
	return paymentsOf(s.BasePoints(), s.situation)
}

// Total is the winner's income.
func (s Score) Total() int {
	return s.Payments().Total()
}

func (s Score) rank() (rank, bool) {
	han := s.Han()
	for _, r := range ranks {
		if han >= r.minHan {
			return r, true
		}
	}
	raw := s.rawBasePoints()
	if raw >= mangan.basePoints || (s.ruleSet.RoundUpMangan() && raw >= roundUpManganBasePoints) {
		return mangan, true
	}
	return rank{}, false
}

func (s Score) rawBasePoints() int {
	return s.Fu() * (1 << (2 + s.Han()))
}
