package mahjongd

import (
	"fmt"

	"github.com/okki-0417/mahjong/cpu"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/tile"
)

// simulatorRequest replays a kyoku, optionally takes the human's action, and
// lets the CPU seats play until the human is awaited or the kyoku ends.
type simulatorRequest struct {
	kyokuRequest
	UserSeat int           `json:"user_seat"`
	Action   *kyoku.Action `json:"action"`
}

type simulatorResponse struct {
	Wall     []string       `json:"wall"`
	Taken    []kyoku.Action `json:"taken"`
	Finished bool           `json:"finished"`
	Awaiting []int          `json:"awaiting"`
	Sight    kyoku.Sight    `json:"sight"`
	Result   *resultJSON    `json:"result"`
}

func simulatorStep(req simulatorRequest) (simulatorResponse, error) {
	if req.UserSeat < 0 || req.UserSeat >= kyoku.Seats {
		return simulatorResponse{}, fmt.Errorf("%w: user_seat %d", errBadRequest, req.UserSeat)
	}
	k, err := req.deal()
	if err != nil {
		return simulatorResponse{}, err
	}
	taken := []kyoku.Action{}
	if req.Action != nil {
		if req.Action.Seat() != req.UserSeat {
			return simulatorResponse{}, fmt.Errorf("%w: the human sits at seat %d", kyoku.ErrIllegalAction, req.UserSeat)
		}
		if k, err = k.Take(*req.Action); err != nil {
			return simulatorResponse{}, err
		}
		taken = append(taken, *req.Action)
	}
	for {
		choice, ok := nextChoice(k, req.UserSeat)
		if !ok {
			break
		}
		if k, err = k.Take(choice); err != nil {
			return simulatorResponse{}, err
		}
		taken = append(taken, choice)
	}
	res := simulatorResponse{
		Wall: tile.Labels(k.Wall().Tiles()), Taken: taken, Finished: k.IsFinished(),
		Awaiting: k.AwaitingSeats(), Sight: k.SeenBy(req.UserSeat),
	}
	if res.Awaiting == nil {
		res.Awaiting = []int{}
	}
	if r, ok := k.Result(); ok {
		result := resultOf(r)
		res.Result = &result
	}
	return res, nil
}

// nextChoice is the next choice made without asking the human: a CPU seat's
// choice while no human answer is pending, or the human's only option on
// their own turn.
func nextChoice(k *kyoku.Kyoku, userSeat int) (kyoku.Action, bool) {
	awaited := k.AwaitingSeats()
	if len(awaited) == 0 {
		return kyoku.Action{}, false
	}
	for _, seat := range awaited {
		if seat == userSeat {
			return forcedChoice(k, userSeat)
		}
	}
	return cpu.Choose(k.SeenBy(awaited[0]))
}

// forcedChoice is the human's single legal action on their own turn: a
// tsumogiri after riichi, or the one tile left after kuikae. It is never
// taken in a claim window, since passing would mark a missed ron.
func forcedChoice(k *kyoku.Kyoku, userSeat int) (kyoku.Action, bool) {
	seen := k.SeenBy(userSeat)
	if !seen.IsMyTurn() {
		return kyoku.Action{}, false
	}
	options := seen.LegalActions()
	if len(options) != 1 {
		return kyoku.Action{}, false
	}
	return options[0], true
}
