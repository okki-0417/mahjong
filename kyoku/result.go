package kyoku

import (
	"encoding/json"
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// ResultKind is how a kyoku ended.
type ResultKind uint8

// Every way a kyoku ends: a win, or one of the draws.
const (
	ResultTsumo ResultKind = iota + 1
	ResultRon
	ResultKyushukyuhai
	ResultRyukyoku
	ResultSanchaho
	ResultSuufonRenda
	ResultSuuchaRiichi
	ResultSuukaikan
)

var resultKindNames = map[ResultKind]string{
	ResultTsumo: "tsumo", ResultRon: "ron", ResultKyushukyuhai: "kyushukyuhai", ResultRyukyoku: "ryukyoku",
	ResultSanchaho: "sanchaho", ResultSuufonRenda: "suufon_renda", ResultSuuchaRiichi: "suucha_riichi",
	ResultSuukaikan: "suukaikan",
}

// String returns the lowercase romaji name.
func (k ResultKind) String() string {
	if n, ok := resultKindNames[k]; ok {
		return n
	}
	return fmt.Sprintf("ResultKind(%d)", uint8(k))
}

// IsWin reports whether the kyoku ended in a win.
func (k ResultKind) IsWin() bool { return k == ResultTsumo || k == ResultRon }

func resultKindOfDraw(d DrawKind) ResultKind {
	switch d {
	case DrawRyukyoku:
		return ResultRyukyoku
	case DrawSanchaho:
		return ResultSanchaho
	case DrawSuufonRenda:
		return ResultSuufonRenda
	case DrawSuuchaRiichi:
		return ResultSuuchaRiichi
	default:
		return ResultSuukaikan
	}
}

const (
	honbaPoints       = 300
	notenPenalty      = 3000
	nagashiManganBase = 2000
)

// RevealedHand is a hand shown at the end: the winner's, or a tenpai seat's
// at an exhaustive draw. The winning tile is kept apart from the hand.
type RevealedHand struct {
	Seat        int
	Hand        hand.Hand
	WinningTile tile.Tile
	HasWinning  bool
}

// MarshalJSON renders {"seat", "closed_tiles", "melds", "winning_tile"}.
func (r RevealedHand) MarshalJSON() ([]byte, error) {
	melds := r.Hand.Melds()
	if melds == nil {
		melds = []hand.Meld{}
	}
	return json.Marshal(struct {
		Seat        int         `json:"seat"`
		ClosedTiles []string    `json:"closed_tiles"`
		Melds       []hand.Meld `json:"melds"`
		WinningTile *string     `json:"winning_tile"`
	}{r.Seat, tile.Labels(r.Hand.ClosedTiles()), melds, optionalTile(r.WinningTile, r.HasWinning)})
}

// Result is how a kyoku ended and what it did to the scores. It is derived
// from the position the terminal action was taken in, since after the end
// the winning tile and the discarder can no longer be read.
type Result struct {
	k       Kyokumen
	kind    ResultKind
	winner  int
	winning *winning.Winning
	score   *winning.Score
	deltas  [Seats]int
}

func resultOf(k Kyokumen, a Action) (*Result, bool) {
	switch a.kind {
	case ActionTsumo:
		return newResult(k, ResultTsumo, a.seat), true
	case ActionRon:
		return newResult(k, ResultRon, a.seat), true
	case ActionKyushukyuhai:
		return newResult(k, ResultKyushukyuhai, -1), true
	}
	return nil, false
}

func newResult(k Kyokumen, kind ResultKind, winner int) *Result {
	r := &Result{k: k, kind: kind, winner: winner}
	if kind.IsWin() {
		w, err := r.buildWinning()
		if err != nil {
			panic(err)
		}
		r.winning = w
		score := w.Score(r.doraCount())
		r.score = &score
		r.deltas = r.winDeltas()
	} else {
		r.deltas = r.drawDeltas()
	}
	return r
}

// Kind returns how the kyoku ended.
func (r *Result) Kind() ResultKind { return r.kind }

// Winner returns the winning seat, if the kyoku was won.
func (r *Result) Winner() (int, bool) { return r.winner, r.kind.IsWin() }

// Loser returns the seat that dealt into a ron, if any.
func (r *Result) Loser() (int, bool) {
	if r.kind != ResultRon {
		return 0, false
	}
	return r.k.claimedFrom(), true
}

// Score returns the winning hand's score, if the kyoku was won.
func (r *Result) Score() (winning.Score, bool) {
	if r.score == nil {
		return winning.Score{}, false
	}
	return *r.score, true
}

// Deltas returns each seat's change in points.
func (r *Result) Deltas() [Seats]int { return r.deltas }

// Scores returns each seat's points after the kyoku.
func (r *Result) Scores() [Seats]int {
	var out [Seats]int
	for i := range out {
		out[i] = r.k.seats[i].score + r.deltas[i]
	}
	return out
}

// DealerContinues reports whether the dealer keeps the deal: a win by the
// dealer, the dealer tenpai (or nagashi mangan) at an exhaustive draw, or
// any abortive draw.
func (r *Result) DealerContinues() bool {
	if r.kind == ResultRyukyoku {
		if containsSeat(r.k.NagashiManganSeats(), r.k.dealerSeat) {
			return true
		}
		return r.k.IsTenpai(r.k.dealerSeat)
	}
	if !r.kind.IsWin() {
		return true
	}
	return r.k.IsDealer(r.winner)
}

// NextHonba returns the repeat counter for the next kyoku.
func (r *Result) NextHonba() int {
	if !r.kind.IsWin() || r.DealerContinues() {
		return r.k.honba + 1
	}
	return 0
}

// CarriedRiichiSticks returns the riichi sticks left on the table for the
// next kyoku; a winner takes them all.
func (r *Result) CarriedRiichiSticks() int {
	if r.kind.IsWin() {
		return 0
	}
	return r.k.riichiSticks
}

// RevealedHands returns the hands shown: the winner's, or every tenpai hand
// at an exhaustive draw. Abortive draws show none.
func (r *Result) RevealedHands() []RevealedHand {
	var out []RevealedHand
	switch {
	case r.kind.IsWin():
		out = append(out, RevealedHand{Seat: r.winner, Hand: r.k.seats[r.winner].hand, WinningTile: r.winningTile(), HasWinning: true})
	case r.kind == ResultRyukyoku:
		for seat := 0; seat < Seats; seat++ {
			if r.k.IsTenpai(seat) {
				out = append(out, RevealedHand{Seat: seat, Hand: r.k.seats[seat].hand})
			}
		}
	}
	return out
}

// DoraIndicators returns the revealed dora indicators.
func (r *Result) DoraIndicators() []tile.Tile { return r.k.DoraIndicators() }

// UradoraIndicators returns the uradora indicators, revealed only when a
// riichi seat won.
func (r *Result) UradoraIndicators() []tile.Tile {
	if !r.kind.IsWin() || !r.k.seats[r.winner].IsRiichi() {
		return nil
	}
	return r.k.UradoraIndicators()
}

func (r *Result) buildWinning() (*winning.Winning, error) {
	s := r.k.seats[r.winner]
	kind := winning.Ron
	if r.kind == ResultTsumo {
		kind = winning.Tsumo
	}
	return winning.New(s.hand, r.winningTile(), winning.Situation{
		WinKind: kind, RoundWind: r.k.roundWind, SeatWind: r.k.SeatWind(r.winner),
		Riichi: s.singleRiichi(), DoubleRiichi: s.IsDoubleRiichi(), Ippatsu: s.IsIppatsu(),
		Rinshan: kind == winning.Tsumo && r.k.IsRinshanDraw(),
		Haitei:  kind == winning.Tsumo && r.k.IsHaiteiDraw(),
		Houtei:  kind == winning.Ron && r.k.DrawsExhausted(),
		Chankan: kind == winning.Ron && r.k.IsChankanWindow(),
	}, r.k.rules)
}

func (r *Result) winningTile() tile.Tile {
	if r.kind == ResultTsumo {
		return r.k.drawn()
	}
	return r.k.claimedTile()
}

func (r *Result) doraCount() int {
	h := r.k.seats[r.winner].hand
	tiles := append(h.AllTiles(), r.winningTile())
	return tile.DoraCount(tiles, append(r.DoraIndicators(), r.UradoraIndicators()...))
}

func (r *Result) winDeltas() [Seats]int {
	var deltas [Seats]int
	for payer, amount := range r.paymentsBySeat() {
		deltas[payer] -= amount
		deltas[r.winner] += amount
	}
	deltas[r.winner] += r.k.riichiSticks * riichiStick
	return deltas
}

func (r *Result) drawDeltas() [Seats]int {
	var deltas [Seats]int
	if r.kind != ResultRyukyoku {
		return deltas
	}
	if nagashi := r.k.NagashiManganSeats(); len(nagashi) > 0 {
		for _, w := range nagashi {
			for payer := 0; payer < Seats; payer++ {
				if payer == w {
					continue
				}
				amount := nagashiManganBase
				if r.k.IsDealer(payer) || r.k.IsDealer(w) {
					amount *= 2
				}
				deltas[payer] -= amount
				deltas[w] += amount
			}
		}
		return deltas
	}
	var tenpai, noten []int
	for seat := 0; seat < Seats; seat++ {
		if r.k.IsTenpai(seat) {
			tenpai = append(tenpai, seat)
		} else {
			noten = append(noten, seat)
		}
	}
	if len(tenpai) == 0 || len(noten) == 0 {
		return deltas
	}
	for _, seat := range tenpai {
		deltas[seat] = notenPenalty / len(tenpai)
	}
	for _, seat := range noten {
		deltas[seat] = -notenPenalty / len(noten)
	}
	return deltas
}

// paymentsBySeat is who pays what: the discarder pays a ron with the honba,
// tsumo is split with the honba shared; a seat liable for a yakuman pays
// that yakuman's share (halved with the discarder on ron).
func (r *Result) paymentsBySeat() map[int]int {
	payments := map[int]int{}
	for seat, amount := range r.liablePayments() {
		payments[seat] += amount
	}
	for seat, amount := range r.sharedPayments() {
		payments[seat] += amount
	}
	for seat, amount := range payments {
		if amount == 0 {
			delete(payments, seat)
		}
	}
	return payments
}

func (r *Result) liableSeat() (seat int, yaku winning.Yaku, ok bool) {
	winner := r.k.seats[r.winner]
	for _, y := range r.score.Yakus() {
		if from, liable := winner.liableSeatFor(y.ID); liable {
			return from, y, true
		}
	}
	return 0, winning.Yaku{}, false
}

func (r *Result) liablePayments() map[int]int {
	liable, yaku, ok := r.liableSeat()
	if !ok {
		return nil
	}
	total := r.score.Total() * yaku.Yakuman / r.score.YakumanCount()
	loser, isRon := r.Loser()
	if !isRon || loser == liable {
		return map[int]int{liable: total}
	}
	return map[int]int{liable: total / 2, loser: total / 2}
}

func (r *Result) sharedPayments() map[int]int {
	honba := r.k.honba * honbaPoints
	shares := r.sharedShares()
	if loser, isRon := r.Loser(); isRon {
		return map[int]int{loser: shares.FromLoser + honba}
	}
	out := map[int]int{}
	payers := Seats - 1
	for seat := 0; seat < Seats; seat++ {
		if seat == r.winner {
			continue
		}
		share := shares.FromNonDealer
		if r.k.IsDealer(seat) && !r.k.IsDealer(r.winner) {
			share = shares.FromDealer
		}
		out[seat] = share + honba/payers
	}
	return out
}

// sharedShares is what remains of the winner's payments after the liable
// yakuman's share is taken out.
func (r *Result) sharedShares() winning.Payments {
	p := r.score.Payments()
	_, yaku, ok := r.liableSeat()
	if !ok {
		return p
	}
	outside := func(amount int) int {
		return amount * (r.score.YakumanCount() - yaku.Yakuman) / r.score.YakumanCount()
	}
	return winning.Payments{FromLoser: outside(p.FromLoser), FromDealer: outside(p.FromDealer), FromNonDealer: outside(p.FromNonDealer)}
}
