package ruleset_test

import (
	"errors"
	"testing"

	"github.com/okki-0417/mahjong/ruleset"
)

func TestDefault(t *testing.T) {
	rs := ruleset.Default()
	if !rs.Kuitan() || rs.RoundUpMangan() || !rs.NagashiMangan() || rs.StartingScore() != 25_000 {
		t.Fatalf("got %+v", rs)
	}
	if rs != (ruleset.RuleSet{}) {
		t.Fatal("the zero value must be the default")
	}
}

func TestWith(t *testing.T) {
	base := ruleset.Default()
	if base.WithKuitan(false).Kuitan() || !base.WithKuitan(false).WithKuitan(true).Kuitan() {
		t.Error("WithKuitan")
	}
	if !base.WithRoundUpMangan(true).RoundUpMangan() {
		t.Error("WithRoundUpMangan")
	}
	if base.WithNagashiMangan(false).NagashiMangan() {
		t.Error("WithNagashiMangan")
	}
	if base.Kuitan() != true || base.RoundUpMangan() != false {
		t.Error("With* must not mutate the receiver")
	}
	rs, err := base.WithStartingScore(30_000)
	if err != nil || rs.StartingScore() != 30_000 {
		t.Errorf("WithStartingScore = %v, %v", rs, err)
	}
	for _, score := range []int{0, -1000} {
		if _, err := base.WithStartingScore(score); !errors.Is(err, ruleset.ErrStartingScore) {
			t.Errorf("WithStartingScore(%d) err = %v", score, err)
		}
	}
}

func TestEquality(t *testing.T) {
	a, b := ruleset.Default().WithKuitan(false), ruleset.Default().WithKuitan(false)
	if a != b {
		t.Error("equal rules must compare equal")
	}
	if ruleset.Default() == ruleset.Default().WithNagashiMangan(false) {
		t.Error("different rules must not compare equal")
	}
	other, _ := ruleset.Default().WithStartingScore(30_000)
	if ruleset.Default() == other {
		t.Error("different starting scores must not compare equal")
	}
	seen := map[ruleset.RuleSet]bool{ruleset.Default().WithKuitan(false): true}
	if !seen[ruleset.Default().WithKuitan(false)] {
		t.Error("usable as a map key")
	}
}
