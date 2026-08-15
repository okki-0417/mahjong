package mahjongd

import (
	"github.com/okki-0417/mahjong/kyoku"
	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
	"github.com/okki-0417/mahjong/winning"
)

type ukeireEntry struct {
	Tile  string `json:"tile"`
	Count int    `json:"count"`
}

func ukeireEntries(u ukeire.Ukeire) []ukeireEntry {
	out := make([]ukeireEntry, 0, len(u.Entries()))
	for _, e := range u.Entries() {
		out = append(out, ukeireEntry{Tile: e.Tile.String(), Count: e.Remaining})
	}
	return out
}

type waitEntry struct {
	Tile  string   `json:"tile"`
	Count int      `json:"count"`
	Kinds []string `json:"kinds"`
}

func waitEntries(u ukeire.Ukeire) []waitEntry {
	out := make([]waitEntry, 0, len(u.Entries()))
	for _, e := range u.Entries() {
		waitKinds := u.WaitKinds(e.Tile)
		kinds := make([]string, 0, len(waitKinds))
		for _, k := range waitKinds {
			kinds = append(kinds, k.String())
		}
		out = append(out, waitEntry{Tile: e.Tile.String(), Count: e.Remaining, Kinds: kinds})
	}
	return out
}

type mentsuJSON struct {
	Kind  string   `json:"kind"`
	Tiles []string `json:"tiles"`
	Open  bool     `json:"open"`
}

type parsedHand struct {
	Kind        string       `json:"kind"`
	WinningTile string       `json:"winning_tile"`
	MentsuList  []mentsuJSON `json:"mentsu_list"`
	PairTile    *string      `json:"pair_tile"`
	PairTiles   []string     `json:"pair_tiles"`
	WaitKind    string       `json:"wait_kind"`
}

func parsedHandOf(f *winning.Form) parsedHand {
	p := parsedHand{
		Kind: f.Kind().String(), WinningTile: f.WinningTile().String(),
		MentsuList: []mentsuJSON{}, PairTiles: tile.Labels(f.PairTiles()), WaitKind: f.WaitKind().String(),
	}
	for _, m := range f.Mentsu() {
		p.MentsuList = append(p.MentsuList, mentsuJSON{Kind: m.Kind().String(), Tiles: tile.Labels(m.Tiles()), Open: m.IsOpen()})
	}
	if pair, ok := f.PairTile(); ok {
		label := pair.String()
		p.PairTile = &label
	}
	if p.PairTiles == nil {
		p.PairTiles = []string{}
	}
	return p
}

type fuItem struct {
	Kind  string   `json:"kind"`
	Label string   `json:"label"`
	Fu    int      `json:"fu"`
	Tiles []string `json:"tiles,omitempty"`
}

type fuResponse struct {
	Subtotal   int        `json:"subtotal"`
	Total      int        `json:"total"`
	Items      []fuItem   `json:"items"`
	ParsedHand parsedHand `json:"parsed_hand"`
}

func fuResponseOf(fu winning.Fu) fuResponse {
	res := fuResponse{Subtotal: fu.Subtotal(), Total: fu.Total(), Items: []fuItem{}, ParsedHand: parsedHandOf(fu.Form())}
	for _, s := range fu.Sources() {
		item := fuItem{Kind: string(s.Kind), Label: s.Label, Fu: s.Fu}
		if len(s.Tiles) > 0 {
			item.Tiles = tile.Labels(s.Tiles)
		}
		res.Items = append(res.Items, item)
	}
	return res
}

type yakuJSON struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Han     int    `json:"han"`
	Yakuman int    `json:"yakuman"`
}

func yakuList(score winning.Score) []yakuJSON {
	out := make([]yakuJSON, 0, len(score.Yakus()))
	for _, y := range score.Yakus() {
		out = append(out, yakuJSON{ID: string(y.ID), Name: y.Name, Han: y.Han, Yakuman: y.Yakuman})
	}
	return out
}

type paymentsJSON struct {
	FromLoser     *int `json:"from_loser"`
	FromDealer    *int `json:"from_dealer"`
	FromNonDealer *int `json:"from_non_dealer"`
}

func paymentsOf(p winning.Payments) paymentsJSON {
	optional := func(v int) *int {
		if v == 0 {
			return nil
		}
		return &v
	}
	return paymentsJSON{FromLoser: optional(p.FromLoser), FromDealer: optional(p.FromDealer), FromNonDealer: optional(p.FromNonDealer)}
}

type scoreResponse struct {
	Yakus        []yakuJSON   `json:"yakus"`
	Han          int          `json:"han"`
	Fu           int          `json:"fu"`
	YakumanCount int          `json:"yakuman_count"`
	DoraCount    int          `json:"dora_count"`
	Total        int          `json:"total"`
	RankLabel    *string      `json:"rank_label"`
	Payments     paymentsJSON `json:"payments"`
	ParsedHand   parsedHand   `json:"parsed_hand"`
}

func rankLabel(score winning.Score) *string {
	if score.RankLabel() == "" {
		return nil
	}
	label := score.RankLabel()
	return &label
}

func scoreResponseOf(score winning.Score, doraCount int) scoreResponse {
	return scoreResponse{
		Yakus: yakuList(score), Han: score.Han(), Fu: score.Fu(), YakumanCount: score.YakumanCount(),
		DoraCount: doraCount, Total: score.Total(), RankLabel: rankLabel(score),
		Payments: paymentsOf(score.Payments()), ParsedHand: parsedHandOf(score.Form()),
	}
}

type resultScoreJSON struct {
	Yakus        []yakuJSON `json:"yakus"`
	Han          int        `json:"han"`
	Fu           int        `json:"fu"`
	YakumanCount int        `json:"yakuman_count"`
	DoraCount    int        `json:"dora_count"`
	Total        int        `json:"total"`
	RankLabel    *string    `json:"rank_label"`
}

type resultJSON struct {
	Kind                string               `json:"kind"`
	Winner              *int                 `json:"winner"`
	Loser               *int                 `json:"loser"`
	Deltas              [kyoku.Seats]int     `json:"deltas"`
	Scores              [kyoku.Seats]int     `json:"scores"`
	DealerContinues     bool                 `json:"dealer_continues"`
	NextHonba           int                  `json:"next_honba"`
	CarriedRiichiSticks int                  `json:"carried_riichi_sticks"`
	RevealedHands       []kyoku.RevealedHand `json:"revealed_hands"`
	DoraIndicators      []string             `json:"dora_indicators"`
	UradoraIndicators   []string             `json:"uradora_indicators"`
	Score               *resultScoreJSON     `json:"score"`
}

func resultOf(r *kyoku.Result) resultJSON {
	res := resultJSON{
		Kind: r.Kind().String(), Deltas: r.Deltas(), Scores: r.Scores(),
		DealerContinues: r.DealerContinues(), NextHonba: r.NextHonba(), CarriedRiichiSticks: r.CarriedRiichiSticks(),
		RevealedHands: r.RevealedHands(), DoraIndicators: tile.Labels(r.DoraIndicators()), UradoraIndicators: tile.Labels(r.UradoraIndicators()),
	}
	if res.RevealedHands == nil {
		res.RevealedHands = []kyoku.RevealedHand{}
	}
	if res.UradoraIndicators == nil {
		res.UradoraIndicators = []string{}
	}
	if w, ok := r.Winner(); ok {
		res.Winner = &w
	}
	if l, ok := r.Loser(); ok {
		res.Loser = &l
	}
	if score, ok := r.Score(); ok {
		res.Score = &resultScoreJSON{
			Yakus: yakuList(score), Han: score.Han(), Fu: score.Fu(), YakumanCount: score.YakumanCount(),
			DoraCount: score.DoraCount(), Total: score.Total(), RankLabel: rankLabel(score),
		}
	}
	return res
}
