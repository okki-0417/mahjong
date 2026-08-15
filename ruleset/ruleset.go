// Package ruleset holds the table rules that mahjong itself leaves open and
// that must be agreed before play: they come from outside the tiles and the
// hand, and they affect both winning (yaku, score) and the kyoku (results).
package ruleset

import "errors"

// ErrStartingScore is returned for a starting score that is not positive.
var ErrStartingScore = errors.New("ruleset: starting score must be positive")

const defaultStartingScore = 25_000

// RuleSet is an immutable set of table rules. The zero value (and Default)
// is the widely used rules; With* methods return a modified copy.
type RuleSet struct {
	noKuitan        bool
	roundUpMangan   bool
	noNagashiMangan bool
	startingScore   int
}

// Default returns the widely used rules: kuitan on, round-up mangan off,
// nagashi mangan on, 25,000 starting score.
func Default() RuleSet {
	return RuleSet{}
}

// Kuitan reports whether an open tanyao is a yaku.
func (r RuleSet) Kuitan() bool {
	return !r.noKuitan
}

// RoundUpMangan reports whether a hand just short of mangan (4 han 30 fu, 3
// han 60 fu) is scored as mangan.
func (r RuleSet) RoundUpMangan() bool {
	return r.roundUpMangan
}

// NagashiMangan reports whether a seat whose discards are all terminals and
// honors at an exhaustive draw is paid as mangan.
func (r RuleSet) NagashiMangan() bool {
	return !r.noNagashiMangan
}

// StartingScore is how many points each seat starts with.
func (r RuleSet) StartingScore() int {
	if r.startingScore == 0 {
		return defaultStartingScore
	}
	return r.startingScore
}

// WithKuitan returns a copy with kuitan set.
func (r RuleSet) WithKuitan(on bool) RuleSet {
	r.noKuitan = !on
	return r
}

// WithRoundUpMangan returns a copy with round-up mangan set.
func (r RuleSet) WithRoundUpMangan(on bool) RuleSet {
	r.roundUpMangan = on
	return r
}

// WithNagashiMangan returns a copy with nagashi mangan set.
func (r RuleSet) WithNagashiMangan(on bool) RuleSet {
	r.noNagashiMangan = !on
	return r
}

// WithStartingScore returns a copy with the starting score set. The score
// must be positive.
func (r RuleSet) WithStartingScore(score int) (RuleSet, error) {
	if score <= 0 {
		return RuleSet{}, ErrStartingScore
	}
	r.startingScore = score
	return r, nil
}
