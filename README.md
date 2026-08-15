# mahjong

A riichi mahjong rules engine in Go: tiles, hands, shanten, ukeire, yaku, fu,
scoring, and the progression of a kyoku. Pure Go, no dependencies outside the
standard library.

> **Status:** early port in progress. The API is not stable until v1.0.0.
> Ported so far: `tile`, `hand` (melds, shanten, waits).
> Coming: `winning` (yaku / fu / score), `ruleset`, `kyoku`, `cpu`, and the
> `mahjongd` HTTP server.

## Install

```sh
go get github.com/okki-0417/mahjong
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
)

func main() {
	closed, _ := tile.ParseAll([]string{"1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "1z", "1z", "2z", "2z"})
	h, _ := hand.New(closed, nil)

	fmt.Println(h.Shanten())  // 0
	fmt.Println(h.IsTenpai()) // true
	fmt.Println(h.Waits())    // [1z 2z]

	after, _ := h.Discard(tile.East, tile.Chun)
	fmt.Println(after.Shanten()) // 1
}
```

Tiles are labelled `1m`–`9m`, `1p`–`9p`, `1s`–`9s` for the numeric suits,
`0m` / `0p` / `0s` for red fives, and `1z`–`7z` for east, south, west, north,
haku, hatsu, chun.

## Packages

| Package | What it models |
| --- | --- |
| `tile` | The 34 kinds plus red fives; suit, number, honor / terminal, dora |
| `hand` | A player's hand: concealed tiles, melds, calls, shanten, waits |
| `mahjongtest` | Helpers that build tiles, melds, and hands from labels in tests |
| `knowledge` | Domain knowledge tests: the rules written as a specification |

## Testing

```sh
go test ./...
go test ./knowledge/ -v   # prints the rules as a specification
```

`testdata/*.jsonl` are golden fixtures recorded from the Ruby implementation
this engine is being ported from; the Go code must agree with them exactly.

## License

MIT
