package kyoku

import (
	"errors"
	"fmt"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// ErrIllegalAction is returned when an action cannot be taken in the
// current position.
var ErrIllegalAction = errors.New("kyoku: illegal action")

// DrawKind is a way a kyoku ends without a win.
type DrawKind uint8

// Every kind of draw. DrawRyukyoku is the exhaustive draw; the others abort the
// kyoku mid-way.
const (
	DrawRyukyoku DrawKind = iota + 1
	DrawSanchaho
	DrawSuufonRenda
	DrawSuuchaRiichi
	DrawSuukaikan
	DrawKyushukyuhai
)

var drawKindNames = map[DrawKind]string{
	DrawRyukyoku: "ryukyoku", DrawSanchaho: "sanchaho", DrawSuufonRenda: "suufon_renda",
	DrawSuuchaRiichi: "suucha_riichi", DrawSuukaikan: "suukaikan", DrawKyushukyuhai: "kyushukyuhai",
}

// String returns the lowercase romaji name.
func (k DrawKind) String() string {
	if n, ok := drawKindNames[k]; ok {
		return n
	}
	return fmt.Sprintf("DrawKind(%d)", uint8(k))
}

// Kyokumen is the whole state of a kyoku at one moment: walls, seats, and
// what happened last. It is derived by folding actions and is never
// stored. It knows every hidden tile, so hand it outside only through
// SeenBy.
type Kyokumen struct {
	live         liveWall
	dead         deadWall
	seats        [Seats]SeatState
	last         Action
	claims       []Action
	dealerSeat   int
	roundWind    tile.Wind
	kyokuNumber  int
	honba        int
	riichiSticks int
	junme        int
	rules        ruleset.RuleSet
}

func dealKyokumen(w Wall, dealerSeat int, roundWind tile.Wind, scores [Seats]int, kyokuNumber, honba, riichiSticks int, rules ruleset.RuleSet) Kyokumen {
	hands := w.Hands()
	var seats [Seats]SeatState
	for i := 0; i < Seats; i++ {
		seats[i] = SeatState{hand: hand.Must(hands[(i-dealerSeat+Seats)%Seats], nil), score: scores[i]}
	}
	return Kyokumen{
		live: liveWall{tiles: w.DrawTiles()}, dead: newDeadWall(w.DeadTiles()), seats: seats,
		dealerSeat: dealerSeat, roundWind: roundWind, kyokuNumber: kyokuNumber,
		honba: honba, riichiSticks: riichiSticks, junme: 1, rules: rules,
	}
}

// Seat returns the state of a seat.
func (k *Kyokumen) Seat(seat int) SeatState { return k.seats[seat] }

// Score returns a seat's points.
func (k *Kyokumen) Score(seat int) int { return k.seats[seat].score }

// IsTenpai reports whether a seat's hand is tenpai.
func (k *Kyokumen) IsTenpai(seat int) bool { return k.seats[seat].IsTenpai() }

// DealerSeat returns the dealer's seat.
func (k *Kyokumen) DealerSeat() int { return k.dealerSeat }

// RoundWind returns the prevailing wind.
func (k *Kyokumen) RoundWind() tile.Wind { return k.roundWind }

// KyokuNumber returns which kyoku of the wind this is (1-4).
func (k *Kyokumen) KyokuNumber() int { return k.kyokuNumber }

// Honba returns the repeat counter.
func (k *Kyokumen) Honba() int { return k.honba }

// RiichiSticks returns the riichi sticks on the table.
func (k *Kyokumen) RiichiSticks() int { return k.riichiSticks }

// Junme returns the turn number, counted each time play returns to the
// dealer.
func (k *Kyokumen) Junme() int { return k.junme }

// RuleSet returns the table rules.
func (k *Kyokumen) RuleSet() ruleset.RuleSet { return k.rules }

// RemainingDraws returns how many tiles are left to draw.
func (k *Kyokumen) RemainingDraws() int { return k.live.remaining() }

// DrawsExhausted reports whether the live wall is empty.
func (k *Kyokumen) DrawsExhausted() bool { return k.live.empty() }

// DoraIndicators returns the revealed dora indicators.
func (k *Kyokumen) DoraIndicators() []tile.Tile { return k.dead.doraIndicatorTiles() }

// UradoraIndicators returns the uradora indicators, one per revealed dora.
func (k *Kyokumen) UradoraIndicators() []tile.Tile { return k.dead.uradoraIndicatorTiles() }

// SeatWind returns a seat's wind, by distance from the dealer.
func (k *Kyokumen) SeatWind(seat int) tile.Wind {
	return tile.EastWind + tile.Wind((seat-k.dealerSeat+Seats)%Seats)
}

// IsDealer reports whether the seat is the dealer.
func (k *Kyokumen) IsDealer(seat int) bool { return seat == k.dealerSeat }

// Shimocha returns the seat to the right, the only one that may chi.
func (k *Kyokumen) Shimocha(seat int) int { return (seat + 1) % Seats }

// LastAction returns the last action folded in, if any.
func (k *Kyokumen) LastAction() (Action, bool) { return k.last, !k.last.IsZero() }

// IsOpening reports whether this is the position right after the deal: no
// action folded, no discards, no melds.
func (k *Kyokumen) IsOpening() bool {
	if !k.last.IsZero() {
		return false
	}
	for _, s := range k.seats {
		if len(s.discards) > 0 || len(s.hand.Melds()) > 0 {
			return false
		}
	}
	return true
}

// SeenBy projects the position onto what one seat can see.
func (k *Kyokumen) SeenBy(seat int) Sight { return Sight{k: k, seat: seat} }

// TileSupply returns the tiles a seat has seen: its own hand, every meld
// and river, and the dora indicators.
func (k *Kyokumen) TileSupply(seat int) tile.Supply {
	seen := append([]tile.Tile(nil), k.seats[seat].hand.ClosedTiles()...)
	for _, s := range k.seats {
		seen = append(seen, s.openTiles()...)
	}
	seen = append(seen, k.DoraIndicators()...)
	return tile.MustSupply(seen)
}

// DiscardingSeat returns the seat that must discard next; ok is false while
// calls are pending or the kyoku is over.
func (k *Kyokumen) DiscardingSeat() (seat int, ok bool) {
	if k.finished() || k.awaitingClaim() {
		return 0, false
	}
	if k.last.IsZero() {
		return k.dealerSeat, true
	}
	return k.last.seat, true
}

// Drawn returns the tile the discarding seat holds beyond its hand: the
// tile it drew from the wall or dead wall. There is none right after a chi
// or pon, since the tiles given up balance the meld.
func (k *Kyokumen) Drawn() (tile.Tile, bool) {
	if _, ok := k.DiscardingSeat(); !ok {
		return 0, false
	}
	if k.last.IsZero() {
		return k.live.nextDraw()
	}
	switch k.last.kind {
	case ActionChi, ActionPon:
		return 0, false
	case ActionRinshan:
		return k.dead.nextRinshan()
	}
	return k.live.nextDraw()
}

// IsRinshanDraw reports whether the drawn tile came from the dead wall.
func (k *Kyokumen) IsRinshanDraw() bool {
	if _, ok := k.Drawn(); !ok {
		return false
	}
	return k.last.kind == ActionRinshan
}

// IsHaiteiDraw reports whether the drawn tile is the last of the live wall.
func (k *Kyokumen) IsHaiteiDraw() bool {
	_, ok := k.Drawn()
	return ok && !k.IsRinshanDraw() && k.RemainingDraws() == 1
}

// ClaimedTile returns the tile other seats may call: the last discard, or
// the tile being added by kakan.
func (k *Kyokumen) ClaimedTile() (tile.Tile, bool) {
	if !k.awaitingClaim() {
		return 0, false
	}
	if k.IsChankanWindow() {
		return k.last.tiles[0], true
	}
	discards := k.seats[k.last.seat].discards
	if len(discards) == 0 {
		return 0, false
	}
	return discards[len(discards)-1].tile, true
}

// ClaimedFrom returns the seat whose tile may be called.
func (k *Kyokumen) ClaimedFrom() (int, bool) {
	if !k.awaitingClaim() {
		return 0, false
	}
	return k.last.seat, true
}

// IsChankanWindow reports whether a kakan is waiting to be robbed.
func (k *Kyokumen) IsChankanWindow() bool {
	return !k.last.IsZero() && k.last.kind == ActionKakan
}

// PendingCall returns a seat's declared but unfinished chi or pon: until it
// discards, the meld is not in its hand and the promised tiles still are.
func (k *Kyokumen) PendingCall(seat int) (Action, bool) {
	if k.last.IsZero() || (k.last.kind != ActionChi && k.last.kind != ActionPon) || k.last.seat != seat {
		return Action{}, false
	}
	return k.last, true
}

// ConcealedTiles returns the tiles a seat still holds in hand: promised
// tiles of a pending call already lie on the table.
func (k *Kyokumen) ConcealedTiles(seat int) []tile.Tile {
	tiles := k.seats[seat].hand.ClosedTiles()
	call, ok := k.PendingCall(seat)
	if !ok {
		return tiles
	}
	return withoutPledged(tiles, call.Tiles())
}

// ConcealedCount returns how many tiles a seat holds in hand.
func (k *Kyokumen) ConcealedCount(seat int) int { return len(k.ConcealedTiles(seat)) }

// DrawKind returns how the position has ended in a draw, if it has. Every
// draw is decided by a discard that nobody claimed.
func (k *Kyokumen) DrawKind() (DrawKind, bool) {
	if !k.awaitingClaim() {
		return 0, false
	}
	if k.sanchaho() {
		return DrawSanchaho, true
	}
	if k.IsChankanWindow() {
		return 0, false
	}
	switch {
	case k.suufonRenda():
		return DrawSuufonRenda, true
	case k.suuchaRiichi():
		return DrawSuuchaRiichi, true
	case k.suukaikan():
		return DrawSuukaikan, true
	case k.DrawsExhausted():
		return DrawRyukyoku, true
	}
	return 0, false
}

// NagashiManganSeats lists the seats that have nagashi mangan; none when
// the rules do not adopt it.
func (k *Kyokumen) NagashiManganSeats() []int {
	if !k.rules.NagashiMangan() {
		return nil
	}
	var out []int
	for i, s := range k.seats {
		if s.IsNagashiMangan() {
			out = append(out, i)
		}
	}
	return out
}

func (k *Kyokumen) awaitingClaim() bool {
	if k.last.IsZero() {
		return false
	}
	return k.last.kind == ActionDiscard || k.last.kind == ActionRiichi || k.last.kind == ActionKakan
}

func (k *Kyokumen) finished() bool {
	if _, drawn := k.DrawKind(); drawn {
		return true
	}
	if k.last.IsZero() {
		return false
	}
	return k.last.kind == ActionTsumo || k.last.kind == ActionRon || k.last.kind == ActionKyushukyuhai
}

func (k *Kyokumen) sanchaho() bool {
	return k.awaitingClaim() && len(k.claims) > 0 && claimSet{k.claims}.aborts()
}

func (k *Kyokumen) suufonRenda() bool {
	if k.anyMeld() {
		return false
	}
	for _, s := range k.seats {
		if len(s.discards) != 1 {
			return false
		}
	}
	first := k.seats[0].discards[0].tile
	if !first.IsWind() {
		return false
	}
	for _, s := range k.seats {
		if !s.discards[0].tile.SameKind(first) {
			return false
		}
	}
	return true
}

func (k *Kyokumen) suuchaRiichi() bool {
	for _, s := range k.seats {
		if !s.IsRiichi() {
			return false
		}
	}
	return true
}

func (k *Kyokumen) suukaikan() bool {
	total, holders := 0, 0
	for _, s := range k.seats {
		n := 0
		for _, m := range s.hand.Melds() {
			if m.IsKan() {
				n++
			}
		}
		total += n
		if n > 0 {
			holders++
		}
	}
	return total >= tile.CopiesPerKind && holders >= 2
}

func (k *Kyokumen) anyMeld() bool {
	for _, s := range k.seats {
		if len(s.hand.Melds()) > 0 {
			return true
		}
	}
	return false
}

func (k *Kyokumen) kanAvailable() bool {
	if k.dead.rinshanExhausted() {
		return false
	}
	need := 0
	if _, ok := k.Drawn(); ok {
		need = 1
	}
	return k.RemainingDraws() > need
}

func (k *Kyokumen) claimedTile() tile.Tile {
	t, _ := k.ClaimedTile()
	return t
}

func (k *Kyokumen) claimedFrom() int {
	return k.last.seat
}

func (k *Kyokumen) drawn() tile.Tile {
	t, _ := k.Drawn()
	return t
}

// advance applies one action: as a response while calls are pending,
// otherwise as the discarding seat's turn.
func (k Kyokumen) advance(a Action) Kyokumen {
	var next Kyokumen
	if k.awaitingClaim() {
		next = k.advanceClaim(a)
	} else {
		next = k.advanceTurn(a)
	}
	if next.unanswerableClaim() {
		return next.settleClaims()
	}
	return next
}

// takenOn is the position the action is really taken in: while calls are
// pending, a non-response is taken after the pending calls are settled.
func (k Kyokumen) takenOn(a Action) Kyokumen {
	if k.awaitingClaim() && !isClaimResponse(a) {
		return k.settleClaims()
	}
	return k
}

func isClaimResponse(a Action) bool {
	switch a.kind {
	case ActionRon, ActionPon, ActionMinkan, ActionChi, ActionPass:
		return true
	}
	return false
}

func (k Kyokumen) advanceClaim(a Action) Kyokumen {
	if !isClaimResponse(a) {
		return k.settleClaims().advance(a)
	}
	return k.recordClaim(a)
}

func (k Kyokumen) recordClaim(a Action) Kyokumen {
	recorded := k.rebuild(k.last)
	recorded.claims = append(append([]Action(nil), k.claims...), a)
	if recorded.claimsOpen() {
		return recorded
	}
	return recorded.settleClaims()
}

func (k *Kyokumen) respondedSeats() []int {
	out := make([]int, 0, len(k.claims))
	for _, c := range k.claims {
		out = append(out, c.seat)
	}
	return out
}

func containsSeat(seats []int, seat int) bool {
	for _, s := range seats {
		if s == seat {
			return true
		}
	}
	return false
}

// claimableSeats are the seats that could respond to the claimed tile.
func (k *Kyokumen) claimableSeats() []int {
	var out []int
	for seat := 0; seat < Seats; seat++ {
		if seat == k.claimedFrom() || len(k.claimsBy(seat)) == 0 {
			continue
		}
		out = append(out, seat)
	}
	return out
}

func (k *Kyokumen) claimsOpen() bool {
	responded := k.respondedSeats()
	for _, seat := range k.claimableSeats() {
		if !containsSeat(responded, seat) {
			return true
		}
	}
	return false
}

func (k *Kyokumen) unanswerableClaim() bool {
	if !k.awaitingClaim() || len(k.claims) > 0 || len(k.claimableSeats()) > 0 {
		return false
	}
	_, drawn := k.DrawKind()
	return !drawn
}

// settleClaims judges the pending claims and closes the window; seats
// that did not answer passed. DrawSanchaho stays open, since nobody wins.
func (k Kyokumen) settleClaims() Kyokumen {
	if !k.awaitingClaim() || k.sanchaho() {
		return k
	}
	taken, ok := claimSet{k.claims}.taken(k.claimedFrom())
	if !ok {
		return k.advanceUnclaimed()
	}
	switch taken.kind {
	case ActionPon, ActionChi:
		return k.advanceCall(taken)
	case ActionMinkan:
		return k.advanceMinkan(taken)
	default:
		return k.rebuild(taken)
	}
}

func (k Kyokumen) advanceTurn(a Action) Kyokumen {
	switch a.kind {
	case ActionDiscard:
		return k.advanceDiscard(a)
	case ActionRiichi:
		return k.advanceRiichi(a)
	case ActionAnkan:
		return k.advanceAnkan(a)
	case ActionKakan:
		return k.rebuild(a)
	default:
		return k.rebuild(a)
	}
}

func (k Kyokumen) advanceDiscard(a Action) Kyokumen {
	if !k.last.IsZero() && (k.last.kind == ActionChi || k.last.kind == ActionPon) {
		return k.advanceCallDiscard(a)
	}
	live, dead := k.wallsAfterDraw()
	next := k.rebuild(a)
	next.seats[a.seat] = k.discardedSeat(a, false).withoutIppatsu()
	next.live, next.dead = live, dead
	return next
}

func (k Kyokumen) advanceRiichi(a Action) Kyokumen {
	declaring := k.discardedSeat(a, true).declareRiichi(k.junme, k.firstUninterruptedTurn(a.seat))
	live, dead := k.wallsAfterDraw()
	next := k.rebuild(a)
	next.seats[a.seat] = declaring
	next.live, next.dead = live, dead
	next.riichiSticks = k.riichiSticks + 1
	return next
}

func (k *Kyokumen) discardedSeat(a Action, declaration bool) SeatState {
	incoming := k.drawn()
	discarded := a.tiles[0]
	s := k.seats[a.seat]
	h, err := s.hand.Discard(discarded, incoming)
	if err != nil {
		panic(err)
	}
	return s.placeDiscard(h, discarded, discarded == incoming, declaration)
}

// advanceCall records a chi or pon: the hand does not change yet, the
// caller's discard will complete it; the discard is marked called and play
// moves to the caller.
func (k Kyokumen) advanceCall(a Action) Kyokumen {
	next := k.rebuild(a)
	next.seats = k.seatsAfterClaim(a.seat)
	next.junme = k.junmeAfterHandoff(a.seat)
	return next
}

func (k Kyokumen) advanceCallDiscard(a Action) Kyokumen {
	caller := k.last.seat
	discarded := a.tiles[0]
	s := k.seats[caller]
	var melded hand.Hand
	var err error
	if k.last.kind == ActionPon {
		melded, err = s.hand.Pon(k.last.called, k.last.Tiles(), discarded)
	} else {
		melded, err = s.hand.Chi(k.last.called, k.last.Tiles(), discarded)
	}
	if err != nil {
		panic(err)
	}
	next := k.rebuild(NewDiscard(caller, discarded))
	next.seats[caller] = s.placeDiscard(melded, discarded, false, false)
	return next
}

func (k Kyokumen) advanceMinkan(a Action) Kyokumen {
	caller := a.seat
	claimed := k.claimedTile()
	var consumed []tile.Tile
	for _, t := range k.seats[caller].hand.ClosedTiles() {
		if t.SameKind(claimed) {
			consumed = append(consumed, t)
		}
	}
	kanHand, err := k.seats[caller].hand.Minkan(claimed, consumed)
	if err != nil {
		panic(err)
	}
	next := k.rebuild(NewRinshan(caller))
	next.seats = k.seatsAfterClaim(caller)
	next.seats[caller] = next.seats[caller].withHand(kanHand)
	next.dead = k.dead.withKanDoraRevealed()
	next.junme = k.junmeAfterHandoff(caller)
	return next
}

func (k Kyokumen) advanceAnkan(a Action) Kyokumen {
	var four []tile.Tile
	for _, t := range k.turnTiles(a.seat) {
		if t.SameKind(a.tiles[0]) && len(four) < tile.CopiesPerKind {
			four = append(four, t)
		}
	}
	kanHand, err := k.seats[a.seat].hand.Ankan(four, k.drawn())
	if err != nil {
		panic(err)
	}
	live, dead := k.wallsAfterDraw()
	next := k.rebuild(NewRinshan(a.seat))
	for i := range next.seats {
		if i == a.seat {
			next.seats[i] = k.seats[i].withHand(kanHand)
		}
		next.seats[i] = next.seats[i].withoutIppatsu()
	}
	next.live, next.dead = live, dead.withKanDoraRevealed()
	return next
}

// confirmKakan completes a kakan nobody robbed: the held-back draw enters
// the hand, the kan is made, a kan dora is revealed, and play goes to the
// rinshan draw.
func (k Kyokumen) confirmKakan() Kyokumen {
	seat := k.last.seat
	drawn, _ := k.live.nextDraw()
	kanHand, err := k.seats[seat].hand.Kakan(k.last.tiles[0], drawn)
	if err != nil {
		panic(err)
	}
	next := k.rebuild(NewRinshan(seat))
	for i := range next.seats {
		s := k.seats[i]
		if k.missedRon(i) {
			s = s.missedRon()
		}
		if i == seat {
			s = s.withHand(kanHand)
		}
		next.seats[i] = s.withoutIppatsu()
	}
	next.live = k.live.draw()
	next.dead = k.dead.withKanDoraRevealed()
	return next
}

// seatsAfterClaim is every seat once a call closes the window: ippatsu is
// gone for all, seats that could have ron'd but did not are furiten, the
// discarder's tile is marked called, and a yakuman made certain by the call
// puts liability on the discarder.
func (k *Kyokumen) seatsAfterClaim(caller int) [Seats]SeatState {
	liable, hasLiable := k.liableYakumanFor(caller)
	var out [Seats]SeatState
	for i, s := range k.seats {
		if i != caller && k.missedRon(i) {
			s = s.missedRon()
		}
		s = s.withoutIppatsu()
		switch {
		case i == k.claimedFrom():
			s = s.markLastCalled(caller)
		case i == caller && hasLiable:
			s = s.withLiability(liable, k.claimedFrom())
		}
		out[i] = s
	}
	return out
}

func (k *Kyokumen) liableYakumanFor(caller int) (winning.YakuID, bool) {
	claimed := k.claimedTile()
	dragons, winds := 0, 0
	for _, m := range k.seats[caller].hand.Melds() {
		first := m.Tiles()[0]
		if first.IsDragon() {
			dragons++
		}
		if first.IsWind() {
			winds++
		}
	}
	if claimed.IsDragon() && dragons == 2 {
		return winning.YakuDaisangen, true
	}
	if claimed.IsWind() && winds == 3 {
		return winning.YakuDaisuushi, true
	}
	return "", false
}

func (k *Kyokumen) missedRon(seat int) bool {
	_, ok := k.ron(seat)
	return ok
}

// advanceUnclaimed closes a window nobody claimed and hands play to the
// next drawer; a chankan window that closes confirms the kakan instead.
func (k Kyokumen) advanceUnclaimed() Kyokumen {
	if k.IsChankanWindow() {
		return k.confirmKakan()
	}
	nextDrawer := k.Shimocha(k.last.seat)
	next := k.rebuild(NewHandoff(nextDrawer))
	for i, s := range k.seats {
		if k.missedRon(i) {
			s = s.missedRon()
		}
		if i == nextDrawer {
			s = s.onDraw()
		}
		next.seats[i] = s
	}
	next.junme = k.junmeAfterHandoff(nextDrawer)
	return next
}

// wallsAfterDraw consumes the drawn tile: a rinshan draw moves the haitei
// into the dead wall and takes the rinshan tile, a normal draw advances the
// live wall.
func (k *Kyokumen) wallsAfterDraw() (liveWall, deadWall) {
	if !k.IsRinshanDraw() {
		return k.live.draw(), k.dead
	}
	return k.live.dropHaitei(), k.dead.drawRinshan(k.live.haitei())
}

// rebuild copies the position with a new last action and no pending claims.
func (k Kyokumen) rebuild(last Action) Kyokumen {
	k.last = last
	k.claims = nil
	return k
}

// junmeAfterHandoff advances the turn count when play passes over (or to)
// the dealer on its way to seat to.
func (k *Kyokumen) junmeAfterHandoff(to int) int {
	from := k.claimedFrom()
	distance := (to - from + Seats) % Seats
	for step := 1; step <= distance; step++ {
		if (from+step)%Seats == k.dealerSeat {
			return k.junme + 1
		}
	}
	return k.junme
}

func (k *Kyokumen) firstUninterruptedTurn(seat int) bool {
	_, drawn := k.Drawn()
	return drawn && len(k.seats[seat].discards) == 0 && !k.anyMeld()
}

// turnTiles are the tiles the discarding seat holds now, drawn tile
// included.
func (k *Kyokumen) turnTiles(seat int) []tile.Tile {
	tiles := k.seats[seat].hand.ClosedTiles()
	if d, ok := k.Drawn(); ok {
		tiles = append(tiles, d)
	}
	return tiles
}

// withoutPledged removes each pledged tile once, so other copies of the
// same kind are kept.
func withoutPledged(tiles, pledged []tile.Tile) []tile.Tile {
	remaining := append([]tile.Tile(nil), tiles...)
	for _, p := range pledged {
		for i, t := range remaining {
			if t == p {
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
	}
	return remaining
}
