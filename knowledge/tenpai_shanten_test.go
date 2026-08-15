package knowledge_test

import (
	"testing"

	"github.com/okki-0417/mahjong/hand"
	mt "github.com/okki-0417/mahjong/mahjongtest"
)

func shantenOf(closed string, melds ...hand.Meld) hand.Shanten {
	return mt.Hand(closed, melds...).Shanten()
}

// 向聴数（シャンテン）= 和了まであと何枚必要か。標準形・七対子・国士の最小値を採る。
func TestShanten(t *testing.T) {
	t.Run("基準", func(t *testing.T) {
		t.Run("和了している手はアガリ（テンパイより1つ進んだ状態）となること", func(t *testing.T) {
			// 和了形は14枚なので、13-3N 枚しか持てない手牌ではなく向聴そのものに問う。
			complete := hand.ShantenOf(mt.Tiles("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s 5s"), nil)
			if complete != hand.Agari {
				t.Fatalf("got %d", complete)
			}
		})
		t.Run("あと1枚で和了できる手はテンパイ（0向聴）となること", func(t *testing.T) {
			if !shantenOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 3p 5s").IsTenpai() {
				t.Fatal("not tenpai")
			}
		})
		t.Run("標準形・七対子・国士無双 の3通りを評価し、最小の向聴数を採ること", func(t *testing.T) {
			// 七対子としては1向聴、標準形としてはもっと遠い手。
			if !shantenOf("1m 1m 2m 2m 3m 3m 4m 4m 6m 6m 7m 7m 9p").IsTenpai() {
				t.Fatal("not tenpai")
			}
		})
	})

	t.Run("標準形", func(t *testing.T) {
		t.Run("4面子1雀頭に向けて、面子・対子・塔子の不足から向聴数を数えること", func(t *testing.T) {
			if got := shantenOf("1m 2m 3m 4m 5m 6m 7m 8m 9m 1p 2p 4p 5s"); got != 1 {
				t.Fatalf("got %d", got)
			}
		})
		t.Run("副露した面子は完成面子として数えること", func(t *testing.T) {
			if !shantenOf("1m 2m 3m 4m 5m 6m 5s", mt.Pon("1z 1z 1z"), mt.Pon("2z 2z 2z")).IsTenpai() {
				t.Fatal("not tenpai")
			}
		})
	})

	t.Run("七対子", func(t *testing.T) {
		t.Run("そろった対子が n 組のとき 6 - n 向聴となること", func(t *testing.T) {
			if got := shantenOf("1m 1m 3m 3m 5m 5m 7m 7m 9p 2s 4s 6s 8s"); got != 2 {
				t.Errorf("four pairs: %d", got)
			}
			if got := shantenOf("1m 1m 3m 3m 5m 5m 7m 7m 9p 9p 2s 4s 6s"); got != 1 {
				t.Errorf("five pairs: %d", got)
			}
		})
		t.Run("同種4枚は2対子として数えない（対子1組ぶんにしかならない）こと", func(t *testing.T) {
			four := shantenOf("1m 1m 1m 1m 3m 3m 5m 5m 7z 1p 4p 7p 9s")
			twoKinds := shantenOf("1m 1m 2m 2m 3m 3m 5m 5m 7z 1p 4p 7p 9s")
			if four <= twoKinds {
				t.Fatalf("four of a kind %d, two kinds %d", four, twoKinds)
			}
		})
	})

	t.Run("国士無双", func(t *testing.T) {
		t.Run("そろった么九牌の種類数と、么九牌の雀頭の有無から向聴数を数えること", func(t *testing.T) {
			if !shantenOf("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 7z").IsTenpai() {
				t.Error("thirteen kinds")
			}
			if !shantenOf("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 6z 6z").IsTenpai() {
				t.Error("twelve kinds with a pair")
			}
			if got := shantenOf("1m 9m 1p 9p 1s 9s 1z 2z 3z 4z 5z 5m 6m"); got != 2 {
				t.Errorf("eleven kinds: %d", got)
			}
		})
	})
}
