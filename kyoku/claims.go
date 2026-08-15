package kyoku

// claimSet is the responses gathered on one discard. Only one claim can be
// taken; a ron beats a call, a pon or minkan beats a chi, and among equal
// claims the seat closest after the discarder wins (atamahane).
type claimSet struct {
	actions []Action
}

const sanchahoCount = 3

func claimPriority(kind ActionKind) int {
	switch kind {
	case ActionRon:
		return 0
	case ActionPon, ActionMinkan:
		return 1
	default:
		return 2
	}
}

// taken returns the claim that goes through, if any: none when everyone
// passed or when three rons abort the kyoku.
func (c claimSet) taken(discarder int) (Action, bool) {
	if c.aborts() {
		return Action{}, false
	}
	var best Action
	found := false
	for _, a := range c.actions {
		if a.kind == ActionPass {
			continue
		}
		if !found || claimLess(a, best, discarder) {
			best, found = a, true
		}
	}
	return best, found
}

func claimLess(a, b Action, discarder int) bool {
	pa, pb := claimPriority(a.kind), claimPriority(b.kind)
	if pa != pb {
		return pa < pb
	}
	return (a.seat-discarder+Seats)%Seats < (b.seat-discarder+Seats)%Seats
}

// aborts reports sanchaho: three seats declared ron on one discard.
func (c claimSet) aborts() bool {
	rons := 0
	for _, a := range c.actions {
		if a.kind == ActionRon {
			rons++
		}
	}
	return rons >= sanchahoCount
}
