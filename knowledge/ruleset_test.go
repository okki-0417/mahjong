package knowledge_test

import (
	"testing"

	"github.com/okki-0417/mahjong/ruleset"
)

// 採用ルール。麻雀が一意に答えを持たない部分は卓ごとに取り決めるので、
// 何を採用するかは牌からも手牌からも決まらず、外から与えられる。
func TestRuleSet(t *testing.T) {
	t.Run("何も指定しなければ、広く採用されている取り決めが選ばれること", func(t *testing.T) {
		adopted := ruleset.Default()
		if !adopted.Kuitan() || adopted.RoundUpMangan() {
			t.Fatalf("kuitan %v round-up %v", adopted.Kuitan(), adopted.RoundUpMangan())
		}
	})
	t.Run("配給原点も取り決めであり、広く採用されているのは25000点持ちであること", func(t *testing.T) {
		if ruleset.Default().StartingScore() != 25_000 {
			t.Fatal("default")
		}
		rs, err := ruleset.Default().WithStartingScore(30_000)
		if err != nil || rs.StartingScore() != 30_000 {
			t.Fatalf("got %d, %v", rs.StartingScore(), err)
		}
	})
}
