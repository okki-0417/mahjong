# mahjong

A riichi mahjong rules engine in Go: tiles, hands, shanten, ukeire, yaku, fu,
scoring, and the progression of a kyoku. Pure Go, no dependencies outside the
standard library.

> **Status:** the port from the Ruby original is complete and verified
> against it (golden fixtures under `testdata/`). The API is not stable
> until v1.0.0.

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
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

func main() {
	closed, _ := tile.ParseAll([]string{"1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "1z", "1z", "2z", "2z"})
	h, _ := hand.New(closed, nil)

	fmt.Println(h.Shanten())  // 0
	fmt.Println(h.IsTenpai()) // true
	fmt.Println(h.Waits())    // [1z 2z]

	w, _ := winning.New(h, tile.East, winning.Situation{
		WinKind: winning.Ron, RoundWind: tile.EastWind, SeatWind: tile.SouthWind, Riichi: true,
	}, ruleset.Default())
	score := w.Score(0)
	fmt.Println(score.Han(), score.Fu(), score.Total()) // 2 40 2600
	for _, y := range score.Yakus() {
		fmt.Println(y.ID, y.Han) // riichi 1, yakuhai_round_wind 1
	}
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
| `winning` | A win: readings of the hand, yaku, fu, score, and payments |
| `ukeire` | The improving tiles of a hand and how many remain unseen |
| `ruleset` | Table rules: kuitan, round-up mangan, nagashi mangan, starting score |
| `kyoku` | One deal: the wall, every legal action, claims, results, and the next deal |
| `cpu` | A computer player choosing from what its seat can see |
| `mahjongtest` | Helpers that build tiles, melds, and hands from labels in tests |
| `knowledge` | Domain knowledge tests: the rules written as a specification |

## Server

`cmd/mahjongd` serves the engine as stateless JSON endpoints (the handlers
live in `internal/mahjongd`; they are a delivery layer, not part of the
library API):

```sh
go run ./cmd/mahjongd -addr :8080
curl -s localhost:8080/v1/shanten -d '{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","4p","5s"]}'
# {"shanten":1}
```

| Endpoint | Body | Answers |
| --- | --- | --- |
| `POST /v1/shanten` | `closed_tiles`, `open_melds` | shanten |
| `POST /v1/ukeire` | same | shanten and the improving tiles with counts |
| `POST /v1/waits` | same | waits with their kinds (tenpai) or ukeire |
| `POST /v1/fu` | hand + `winning_tile`, `win_kind`, winds | fu breakdown and the reading |
| `POST /v1/score` | fu body + situation flags, `dora_count`, `round_up_mangan` | yaku, han, fu, payments |
| `POST /v1/discard-analysis` | a drawn hand (14 − 3N tiles) | every discard by shanten and ukeire |
| `POST /v1/simulator/step` | wall, setup, rules, `actions`, `user_seat`, `action` | replays, takes the human's action, plays the CPU seats, returns the human's sight |

The Dockerfile builds a static image for deployment.

## Testing

```sh
go test ./...
go test ./knowledge/ -v   # prints the rules as a specification
```

`testdata/*.jsonl` are golden fixtures recorded from the Ruby implementation
this engine is being ported from; the Go code must agree with them exactly.

## License

MIT
