package mahjongtest

import (
	"fmt"
	"sort"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
)

const (
	defaultScore = 25_000
	handSize     = 13
	deadSize     = 14
	liveSize     = 136 - kyoku.Seats*handSize - deadSize
	uradoraIndex = 4
	doraIndex    = 9
	chunkSize    = 4
	chunkRounds  = 3
	seatsAtTable = kyoku.Seats
)

// KyokuSpec describes a kyoku to build. The wall is compiled backwards from
// the hands, draws, and indicators asked for, then dealt through kyoku.Deal
// so only positions reachable in play can be made; melds and riichi are
// produced by folding Actions.
type KyokuSpec struct {
	// Hands gives a seat's 13 dealt tiles by label. Other seats are dealt
	// from what remains.
	Hands map[int]string
	// Draws lists the tiles drawn from the live wall in order.
	Draws string
	// SeatDraws lists the tiles a seat draws, in order, assigned to that
	// seat's draw turns while the wall is burned and afterwards.
	SeatDraws map[int]string
	// Dora, Uradora are the (first) indicators; Rinshan lists rinshan tiles
	// in the order they are taken.
	Dora    string
	Uradora string
	Rinshan string
	Dealer  int
	// RoundWind defaults to east, KyokuNumber to 1.
	RoundWind   tile.Wind
	KyokuNumber int
	Scores      map[int]int
	Honba       int
	// RiichiSticks on the table at the start.
	RiichiSticks int
	// Wall, when set, is how many live-wall tiles to leave: everyone
	// tsumogiri-discards until then. Only 0 or a count whose burn is a
	// multiple of four is allowed, so seats stay aligned with SeatDraws.
	Wall *int
	// RuleSet defaults to ruleset.Default().
	RuleSet ruleset.RuleSet
	// Actions are folded after the burn.
	Actions []kyoku.Action
}

// WallOf is a convenience for KyokuSpec.Wall.
func WallOf(n int) *int { return &n }

// BuildKyoku compiles the spec into a wall and deals it, folding the
// actions. It panics when the spec cannot be satisfied.
func BuildKyoku(spec KyokuSpec) *kyoku.Kyoku {
	k, err := TryBuildKyoku(spec)
	if err != nil {
		panic(err)
	}
	return k
}

// TryBuildKyoku is BuildKyoku returning the error instead of panicking.
func TryBuildKyoku(spec KyokuSpec) (*kyoku.Kyoku, error) {
	burnCount := 0
	if spec.Wall != nil {
		burnCount = liveSize - *spec.Wall
		if *spec.Wall != 0 && burnCount%seatsAtTable != 0 {
			return nil, fmt.Errorf("mahjongtest: Wall must be 0 or leave a burn that is a multiple of four (2, 6, 10, ...)")
		}
	}
	pool := tile.FullSet(true)
	take := func(labels string) ([]tile.Tile, error) {
		var out []tile.Tile
		for _, t := range Tiles(labels) {
			idx := -1
			for i, candidate := range pool {
				if candidate == t {
					idx = i
					break
				}
			}
			if idx < 0 {
				return nil, fmt.Errorf("mahjongtest: not enough %v in the set", t)
			}
			pool = append(pool[:idx], pool[idx+1:]...)
			out = append(out, t)
		}
		return out, nil
	}

	fixed := map[int][]tile.Tile{}
	for seat, labels := range spec.Hands {
		held, err := take(labels)
		if err != nil {
			return nil, err
		}
		if len(held) != handSize {
			return nil, fmt.Errorf("mahjongtest: Hands[%d] must be the %d dealt tiles, got %d; make melds with actions", seat, handSize, len(held))
		}
		fixed[seat] = held
	}

	seatDraws := map[int][]tile.Tile{}
	for seat, labels := range spec.SeatDraws {
		drawn, err := take(labels)
		if err != nil {
			return nil, err
		}
		seatDraws[seat] = drawn
	}
	var flatDraws []tile.Tile
	if spec.Draws != "" {
		drawn, err := take(spec.Draws)
		if err != nil {
			return nil, err
		}
		flatDraws = drawn
	}
	specials := map[int]tile.Tile{}
	if spec.Rinshan != "" {
		rinshan, err := take(spec.Rinshan)
		if err != nil {
			return nil, err
		}
		for i, t := range rinshan {
			specials[i] = t
		}
	}
	if spec.Uradora != "" {
		t, err := take(spec.Uradora)
		if err != nil {
			return nil, err
		}
		specials[uradoraIndex] = t[0]
	}
	if spec.Dora != "" {
		t, err := take(spec.Dora)
		if err != nil {
			return nil, err
		}
		specials[doraIndex] = t[0]
	}

	protected := protectedKinds(fixed)
	burnTiles := make([]tile.Tile, 0, burnCount)
	for position := 0; position < burnCount; position++ {
		seat := (spec.Dealer + position) % seatsAtTable
		if len(seatDraws[seat]) > 0 {
			burnTiles = append(burnTiles, seatDraws[seat][0])
			seatDraws[seat] = seatDraws[seat][1:]
			continue
		}
		idx := indexWhere(pool, func(t tile.Tile) bool { return !t.IsTerminalOrHonor() && !protected[t.Kind()] })
		if idx < 0 {
			idx = indexWhere(pool, func(t tile.Tile) bool { return !protected[t.Kind()] })
		}
		if idx < 0 {
			return nil, fmt.Errorf("mahjongtest: not enough tiles to burn the wall without a fixed hand's waits")
		}
		burnTiles = append(burnTiles, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}
	if spec.Wall != nil && len(flatDraws) > *spec.Wall {
		return nil, fmt.Errorf("mahjongtest: Draws do not fit in the %d tiles left in the wall", *spec.Wall)
	}

	dealt := map[int][]tile.Tile{}
	var dealtSeats []int
	for seat := 0; seat < seatsAtTable; seat++ {
		if _, ok := fixed[seat]; !ok {
			dealtSeats = append(dealtSeats, seat)
		}
	}
	for i := 0; i < handSize; i++ {
		for _, seat := range dealtSeats {
			dealt[seat] = append(dealt[seat], pool[0])
			pool = pool[1:]
		}
	}

	deadCount := deadSize - len(specials)
	dead := append([]tile.Tile(nil), pool[len(pool)-deadCount:]...)
	pool = pool[:len(pool)-deadCount]
	specialIndexes := make([]int, 0, len(specials))
	for i := range specials {
		specialIndexes = append(specialIndexes, i)
	}
	sort.Ints(specialIndexes)
	for _, i := range specialIndexes {
		dead = append(dead[:i], append([]tile.Tile{specials[i]}, dead[i:]...)...)
	}

	liveTail := make([]tile.Tile, 0, liveSize-burnCount)
	for position := burnCount; position < liveSize; position++ {
		seat := (spec.Dealer + position) % seatsAtTable
		switch {
		case len(seatDraws[seat]) > 0:
			liveTail = append(liveTail, seatDraws[seat][0])
			seatDraws[seat] = seatDraws[seat][1:]
		case len(flatDraws) > 0:
			liveTail = append(liveTail, flatDraws[0])
			flatDraws = flatDraws[1:]
		default:
			liveTail = append(liveTail, pool[0])
			pool = pool[1:]
		}
	}
	for seat, rest := range seatDraws {
		if len(rest) > 0 {
			return nil, fmt.Errorf("mahjongtest: SeatDraws[%d] do not fit in the wall", seat)
		}
	}

	rows := make([][]tile.Tile, seatsAtTable)
	for slot := 0; slot < seatsAtTable; slot++ {
		seat := (slot + spec.Dealer) % seatsAtTable
		if held, ok := fixed[seat]; ok {
			rows[slot] = held
		} else {
			rows[slot] = dealt[seat]
		}
	}
	tiles := chunkLayout(rows)
	tiles = append(tiles, burnTiles...)
	tiles = append(tiles, liveTail...)
	tiles = append(tiles, dead...)
	wall, err := kyoku.NewWall(tiles)
	if err != nil {
		return nil, err
	}

	var scores [kyoku.Seats]int
	for seat := range scores {
		if s, ok := spec.Scores[seat]; ok {
			scores[seat] = s
		} else {
			scores[seat] = defaultScore
		}
	}
	k, err := kyoku.Deal(kyoku.Setup{
		Wall: &wall, DealerSeat: spec.Dealer, RoundWind: spec.RoundWind, KyokuNumber: spec.KyokuNumber,
		Scores: &scores, Honba: spec.Honba, RiichiSticks: spec.RiichiSticks, RuleSet: spec.RuleSet,
	})
	if err != nil {
		return nil, err
	}
	for position, t := range burnTiles {
		if k, err = takeSettling(k, kyoku.NewDiscard((spec.Dealer+position)%seatsAtTable, t)); err != nil {
			return nil, err
		}
	}
	if burnCount > 0 && *spec.Wall > 0 {
		k = AfterOthersPass(k)
	}
	for _, a := range spec.Actions {
		if k, err = takeSettling(k, a); err != nil {
			return nil, err
		}
	}
	return k, nil
}

// takeSettling takes an action, first closing an open claim window with
// passes when the action is not itself a response to it. Specs list only
// the moves that matter; the seats that merely decline are implied.
func takeSettling(k *kyoku.Kyoku, a kyoku.Action) (*kyoku.Kyoku, error) {
	switch a.Kind() {
	case kyoku.ActionRon, kyoku.ActionPon, kyoku.ActionMinkan, kyoku.ActionChi, kyoku.ActionPass:
	default:
		k = AfterOthersPass(k)
	}
	return k.Take(a)
}

// AfterOthersPass takes every pending pass until no seat is left to
// answer, for tests that are not about the claim window itself.
func AfterOthersPass(k *kyoku.Kyoku) *kyoku.Kyoku {
	for {
		var pass *kyoku.Action
		for _, a := range k.Kyokumen().LegalActions() {
			if a.Kind() == kyoku.ActionPass {
				pass = &a
				break
			}
		}
		if pass == nil {
			return k
		}
		next, err := k.Take(*pass)
		if err != nil {
			panic(err)
		}
		k = next
	}
}

// protectedKinds are the waits of every fixed tenpai hand: burned tiles
// must avoid them so no furiten or stray ron window is written into the
// kyoku.
func protectedKinds(fixed map[int][]tile.Tile) map[tile.Tile]bool {
	out := map[tile.Tile]bool{}
	for _, held := range fixed {
		h := hand.Must(held, nil)
		if h.IsTenpai() {
			for _, w := range h.Waits() {
				out[w.Kind()] = true
			}
		}
	}
	return out
}

func indexWhere(tiles []tile.Tile, pred func(tile.Tile) bool) int {
	for i, t := range tiles {
		if pred(t) {
			return i
		}
	}
	return -1
}

// chunkLayout is the deal in reverse: four at a time for three rounds from
// row 0, then one each, so dealing the wall gives back the rows.
func chunkLayout(rows [][]tile.Tile) []tile.Tile {
	var out []tile.Tile
	for round := 0; round < chunkRounds; round++ {
		for _, row := range rows {
			out = append(out, row[round*chunkSize:(round+1)*chunkSize]...)
		}
	}
	for _, row := range rows {
		out = append(out, row[handSize-1])
	}
	return out
}

// DiscardAction builds a discard from a label.
func DiscardAction(seat int, label string) kyoku.Action { return kyoku.NewDiscard(seat, T(label)) }

// RiichiAction builds a riichi declaration from a label.
func RiichiAction(seat int, label string) kyoku.Action { return kyoku.NewRiichi(seat, T(label)) }

// AnkanAction builds an ankan declaration from a label.
func AnkanAction(seat int, label string) kyoku.Action { return kyoku.NewAnkan(seat, T(label)) }

// KakanAction builds a kakan declaration from a label.
func KakanAction(seat int, label string) kyoku.Action { return kyoku.NewKakan(seat, T(label)) }

// MinkanAction builds a minkan call from the called tile's label.
func MinkanAction(seat int, called string) kyoku.Action { return kyoku.NewMinkan(seat, T(called)) }

// PonAction builds a pon call from labels.
func PonAction(seat int, called, consumed string) kyoku.Action {
	return kyoku.NewPon(seat, T(called), Tiles(consumed))
}

// ChiAction builds a chi call from labels.
func ChiAction(seat int, called, consumed string) kyoku.Action {
	return kyoku.NewChi(seat, T(called), Tiles(consumed))
}
