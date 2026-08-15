package kyoku

import (
	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/winning"
)

// LegalActions lists every action some seat may take right now: responses
// to a claimable tile while a window is open, otherwise the discarding
// seat's choices. Illegal actions are already left out, so a caller picks
// one and Take accepts it.
func (k *Kyokumen) LegalActions() []Action {
	if k.finished() {
		return nil
	}
	if k.awaitingClaim() {
		return k.claimOptions()
	}
	return k.turnActions()
}

func (k *Kyokumen) claimOptions() []Action {
	responded := k.respondedSeats()
	var out []Action
	for _, seat := range k.claimableSeats() {
		if containsSeat(responded, seat) {
			continue
		}
		out = append(out, k.claimsBy(seat)...)
		out = append(out, NewPass(seat))
	}
	return out
}

func (k *Kyokumen) claimsBy(seat int) []Action {
	var out []Action
	if r, ok := k.ron(seat); ok {
		out = append(out, r)
	}
	if k.IsChankanWindow() {
		return out
	}
	out = append(out, k.pon(seat)...)
	if m, ok := k.minkan(seat); ok {
		out = append(out, m)
	}
	return append(out, k.chi(seat)...)
}

// ron is offered when the seat can win on the claimed tile and is not
// furiten. The shape is checked first because furiten needs every wait.
func (k *Kyokumen) ron(seat int) (Action, bool) {
	if seat == k.claimedFrom() {
		return Action{}, false
	}
	if _, err := k.winningFor(seat); err != nil {
		return Action{}, false
	}
	if k.seats[seat].IsFuriten() {
		return Action{}, false
	}
	return NewRon(seat), true
}

func (k *Kyokumen) winningFor(seat int) (*winning.Winning, error) {
	s := k.seats[seat]
	return winning.New(s.hand, k.claimedTile(), winning.Situation{
		WinKind: winning.Ron, RoundWind: k.roundWind, SeatWind: k.SeatWind(seat),
		Riichi: s.singleRiichi(), DoubleRiichi: s.IsDoubleRiichi(), Ippatsu: s.IsIppatsu(),
		Chankan: k.IsChankanWindow(), Houtei: !k.IsChankanWindow() && k.DrawsExhausted(),
	}, k.rules)
}

// pon offers each distinct pair of matching tiles; a red five makes a
// different choice from a plain one.
func (k *Kyokumen) pon(seat int) []Action {
	if !k.callable(seat) {
		return nil
	}
	same := k.sameKindTiles(seat, k.claimedTile())
	var out []Action
	for i := 0; i < len(same); i++ {
		for j := i + 1; j < len(same); j++ {
			out = appendUnique(out, NewPon(seat, k.claimedTile(), []tile.Tile{same[i], same[j]}))
		}
	}
	return out
}

func (k *Kyokumen) minkan(seat int) (Action, bool) {
	if !k.callable(seat) || !k.kanAvailable() || len(k.sameKindTiles(seat, k.claimedTile())) < 3 {
		return Action{}, false
	}
	return NewMinkan(seat, k.claimedTile()), true
}

// chi is only for the seat to the right of the discarder, on a numeric tile.
func (k *Kyokumen) chi(seat int) []Action {
	if !k.callable(seat) || k.Shimocha(k.claimedFrom()) != seat || !k.claimedTile().IsNumeric() {
		return nil
	}
	var out []Action
	for _, pair := range k.sequencePairs(seat) {
		out = appendUnique(out, NewChi(seat, k.claimedTile(), pair))
	}
	return out
}

// sequencePairs are the two-tile combinations that make a sequence with the
// claimed tile: below it, around it, and above it.
func (k *Kyokumen) sequencePairs(seat int) [][]tile.Tile {
	n := k.claimedTile().EffectiveNumber()
	var out [][]tile.Tile
	for _, shape := range [][2]int{{n - 2, n - 1}, {n - 1, n + 1}, {n + 1, n + 2}} {
		if shape[0] < 1 || shape[1] > 9 {
			continue
		}
		for _, lower := range k.numericTiles(seat, shape[0]) {
			for _, upper := range k.numericTiles(seat, shape[1]) {
				out = append(out, []tile.Tile{lower, upper})
			}
		}
	}
	return out
}

func (k *Kyokumen) turnActions() []Action {
	seat, ok := k.DiscardingSeat()
	if !ok {
		return nil
	}
	var out []Action
	if a, ok := k.tsumo(seat); ok {
		out = append(out, a)
	}
	out = append(out, k.ankan(seat)...)
	out = append(out, k.kakan(seat)...)
	if a, ok := k.kyushukyuhai(seat); ok {
		out = append(out, a)
	}
	out = append(out, k.riichi(seat)...)
	return append(out, k.discard(seat)...)
}

func (k *Kyokumen) tsumo(seat int) (Action, bool) {
	drawn, ok := k.Drawn()
	if !ok {
		return Action{}, false
	}
	s := k.seats[seat]
	_, err := winning.New(s.hand, drawn, winning.Situation{
		WinKind: winning.Tsumo, RoundWind: k.roundWind, SeatWind: k.SeatWind(seat),
		Riichi: s.singleRiichi(), DoubleRiichi: s.IsDoubleRiichi(), Ippatsu: s.IsIppatsu(),
		Rinshan: k.IsRinshanDraw(), Haitei: k.IsHaiteiDraw(),
	}, k.rules)
	if err != nil {
		return Action{}, false
	}
	return NewTsumo(seat), true
}

// ankan is offered for each kind the seat holds four of, drawn tile
// included. After riichi only a kan that keeps the waits is allowed.
func (k *Kyokumen) ankan(seat int) []Action {
	if _, ok := k.Drawn(); !ok || !k.kanAvailable() {
		return nil
	}
	var out []Action
	for _, kind := range k.ankanKinds(seat) {
		out = append(out, NewAnkan(seat, kind))
	}
	return out
}

func (k *Kyokumen) ankanKinds(seat int) []tile.Tile {
	counts := map[tile.Tile]int{}
	var order []tile.Tile
	for _, t := range k.turnTiles(seat) {
		if counts[t.Kind()] == 0 {
			order = append(order, t.Kind())
		}
		counts[t.Kind()]++
	}
	var out []tile.Tile
	for _, kind := range order {
		if counts[kind] >= tile.CopiesPerKind && k.ankanKeepsWait(seat, kind) {
			out = append(out, kind)
		}
	}
	return out
}

func (k *Kyokumen) ankanKeepsWait(seat int, kind tile.Tile) bool {
	s := k.seats[seat]
	if !s.IsRiichi() {
		return true
	}
	drawn := k.drawn()
	if !drawn.SameKind(kind) {
		return false
	}
	var inHand []tile.Tile
	for _, t := range s.hand.ClosedTiles() {
		if t.SameKind(kind) {
			inHand = append(inHand, t)
		}
	}
	after, err := s.hand.Ankan(append(inHand, drawn), drawn)
	if err != nil {
		return false
	}
	return sameTiles(tile.Sorted(s.hand.Waits()), tile.Sorted(after.Waits()))
}

func sameTiles(a, b []tile.Tile) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// kakan is offered for each held tile that matches a pon; not after riichi.
func (k *Kyokumen) kakan(seat int) []Action {
	if _, ok := k.Drawn(); !ok || !k.kanAvailable() || k.seats[seat].IsRiichi() {
		return nil
	}
	var ponKinds []tile.Tile
	for _, m := range k.seats[seat].hand.Melds() {
		if m.Kind() == hand.Pon {
			ponKinds = append(ponKinds, m.Tiles()[0])
		}
	}
	var out []Action
	for _, t := range k.turnTiles(seat) {
		for _, kind := range ponKinds {
			if kind.SameKind(t) {
				out = appendUnique(out, NewKakan(seat, t))
				break
			}
		}
	}
	return out
}

func (k *Kyokumen) kyushukyuhai(seat int) (Action, bool) {
	if !k.firstUninterruptedTurn(seat) {
		return Action{}, false
	}
	kinds := map[tile.Tile]bool{}
	for _, t := range k.turnTiles(seat) {
		if t.IsTerminalOrHonor() {
			kinds[t.Kind()] = true
		}
	}
	if len(kinds) < 9 {
		return Action{}, false
	}
	return NewKyushukyuhai(seat), true
}

// riichi is offered for each discard that keeps a concealed hand tenpai,
// given the points and enough draws left.
func (k *Kyokumen) riichi(seat int) []Action {
	s := k.seats[seat]
	_, pending := k.PendingCall(seat)
	if !s.IsMenzen() || pending || s.IsRiichi() || s.score < riichiStick || k.RemainingDraws() < Seats {
		return nil
	}
	held := k.turnTiles(seat)
	melds := s.hand.Melds()
	var out []Action
	for i, t := range held {
		rest := append(append([]tile.Tile(nil), held[:i]...), held[i+1:]...)
		if h, err := hand.New(rest, melds); err == nil && h.IsTenpai() {
			out = appendUnique(out, NewRiichi(seat, t))
		}
	}
	return out
}

// discard offers every held tile; only the drawn tile after riichi, and not
// a kuikae tile right after a call.
func (k *Kyokumen) discard(seat int) []Action {
	var out []Action
	for _, t := range k.discardableTiles(seat) {
		out = appendUnique(out, NewDiscard(seat, t))
	}
	return out
}

func (k *Kyokumen) discardableTiles(seat int) []tile.Tile {
	if k.seats[seat].IsRiichi() {
		if d, ok := k.Drawn(); ok {
			return []tile.Tile{d}
		}
		return nil
	}
	call, ok := k.PendingCall(seat)
	if !ok {
		return k.turnTiles(seat)
	}
	forbidden := k.kuikaeTiles(call)
	var out []tile.Tile
	for _, t := range withoutPledged(k.turnTiles(seat), call.Tiles()) {
		blocked := false
		for _, kind := range forbidden {
			if kind.SameKind(t) {
				blocked = true
			}
		}
		if !blocked {
			out = append(out, t)
		}
	}
	return out
}

// kuikaeTiles are the tiles that would swap the called meld back into the
// same shape: the called tile itself, and for a ryanmen chi the other end
// of the sequence.
func (k *Kyokumen) kuikaeTiles(call Action) []tile.Tile {
	out := []tile.Tile{call.called}
	if call.kind != ActionChi {
		return out
	}
	called := call.called.EffectiveNumber()
	consumed := call.Tiles()
	a, b := consumed[0].EffectiveNumber(), consumed[1].EffectiveNumber()
	opposite := 0
	switch {
	case a == called+1 && b == called+2:
		opposite = called + 3
	case a == called-2 && b == called-1:
		opposite = called - 3
	}
	if opposite >= 1 && opposite <= 9 {
		out = append(out, tile.Of(call.called.Suit(), opposite))
	}
	return out
}

func (k *Kyokumen) callable(seat int) bool {
	return !k.seats[seat].IsRiichi() && !k.DrawsExhausted()
}

func (k *Kyokumen) sameKindTiles(seat int, t tile.Tile) []tile.Tile {
	var out []tile.Tile
	for _, held := range k.seats[seat].hand.ClosedTiles() {
		if held.SameKind(t) {
			out = append(out, held)
		}
	}
	return out
}

func (k *Kyokumen) numericTiles(seat int, number int) []tile.Tile {
	var out []tile.Tile
	for _, held := range k.seats[seat].hand.ClosedTiles() {
		if held.IsNumeric() && held.Suit() == k.claimedTile().Suit() && held.EffectiveNumber() == number {
			out = append(out, held)
		}
	}
	return out
}

func appendUnique(actions []Action, a Action) []Action {
	for _, existing := range actions {
		if existing == a {
			return actions
		}
	}
	return append(actions, a)
}
