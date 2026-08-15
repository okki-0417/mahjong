package mahjongd

import (
	"errors"

	"github.com/okki-0417/mahjong/hand"
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/ruleset"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
	"github.com/okki-0417/mahjong/winning"
)

// errorCodes maps sentinel errors to the stable codes clients can branch or
// localize on. Wrapped errors are listed before what they wrap so the most
// specific code wins.
var errorCodes = []struct {
	err  error
	code string
}{
	{errBadRequest, "bad_request"},
	{tile.ErrInvalidLabel, "invalid_tile"},
	{tile.ErrOversupplied, "too_many_copies"},
	{hand.ErrInvalidMeld, "invalid_meld"},
	{hand.ErrTooManyMelds, "too_many_melds"},
	{hand.ErrClosedTileCount, "closed_tile_count"},
	{hand.ErrTooManyCopies, "too_many_copies"},
	{hand.ErrTileNotInHand, "tile_not_in_hand"},
	{hand.ErrNoPonToKakan, "no_pon_to_kakan"},
	{hand.ErrInvalidTile, "invalid_tile"},
	{winning.ErrNoForm, "no_form"},
	{winning.ErrNoYaku, "no_yaku"},
	{winning.ErrNotWinning, "not_winning"},
	{winning.ErrTileExhausted, "tile_exhausted"},
	{winning.ErrInvalidSituation, "invalid_situation"},
	{ruleset.ErrStartingScore, "invalid_starting_score"},
	{ukeire.ErrDrawnCount, "drawn_tile_count"},
	{kyoku.ErrInvalidAction, "invalid_action"},
	{kyoku.ErrIllegalAction, "illegal_action"},
	{kyoku.ErrInvalidWall, "invalid_wall"},
	{kyoku.ErrInvalidSetup, "invalid_setup"},
	{kyoku.ErrFinished, "finished"},
	{kyoku.ErrNotFinished, "not_finished"},
	{kyoku.ErrHanchanOver, "hanchan_over"},
}

const unknownErrorCode = "rejected"

func errorCode(err error) string {
	for _, e := range errorCodes {
		if errors.Is(err, e.err) {
			return e.code
		}
	}
	return unknownErrorCode
}
