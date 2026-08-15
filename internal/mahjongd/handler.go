package mahjongd

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/okki-0417/mahjong/tile"
	"github.com/okki-0417/mahjong/ukeire"
)

// NewHandler returns the HTTP handler serving every endpoint under /v1,
// logging each request to logger (slog.Default when nil).
func NewHandler(logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /v1/shanten", handle(shanten))
	mux.HandleFunc("POST /v1/ukeire", handle(ukeireOf))
	mux.HandleFunc("POST /v1/waits", handle(waits))
	mux.HandleFunc("POST /v1/fu", handle(fu))
	mux.HandleFunc("POST /v1/score", handle(score))
	mux.HandleFunc("POST /v1/discard-analysis", handle(discardAnalysis))
	mux.HandleFunc("POST /v1/simulator/step", handle(simulatorStep))
	return logged(logger, mux)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logged(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration", time.Since(started))
	})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func errorResponseOf(err error) errorResponse {
	return errorResponse{Error: err.Error(), Code: errorCode(err)}
}

// handle wraps an endpoint: it reads JSON, runs it, and maps errors to
// statuses. Malformed input is 400; input the rules reject is 422.
func handle[Req any, Res any](endpoint func(Req) (Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req
		if err := decode(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponseOf(err))
			return
		}
		res, err := endpoint(req)
		if err != nil {
			status := http.StatusUnprocessableEntity
			if errors.Is(err, errBadRequest) {
				status = http.StatusBadRequest
			}
			writeJSON(w, status, errorResponseOf(err))
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type shantenResponse struct {
	Shanten int `json:"shanten"`
}

func shanten(req handRequest) (shantenResponse, error) {
	h, err := req.hand()
	if err != nil {
		return shantenResponse{}, err
	}
	return shantenResponse{Shanten: int(h.Shanten())}, nil
}

type ukeireResponse struct {
	Shanten int           `json:"shanten"`
	Ukeire  []ukeireEntry `json:"ukeire"`
}

func ukeireOf(req handRequest) (ukeireResponse, error) {
	h, err := req.hand()
	if err != nil {
		return ukeireResponse{}, err
	}
	return ukeireResponse{Shanten: int(h.Shanten()), Ukeire: ukeireEntries(ukeire.OfHand(h))}, nil
}

type waitsResponse struct {
	Shanten int           `json:"shanten"`
	Waits   []waitEntry   `json:"waits"`
	Ukeire  []ukeireEntry `json:"ukeire"`
}

// waits answers the winning tiles of a tenpai hand; a hand that is not
// tenpai gets its ukeire instead, since the waits are the answer only when
// there are any.
func waits(req handRequest) (waitsResponse, error) {
	h, err := req.hand()
	if err != nil {
		return waitsResponse{}, err
	}
	u := ukeire.OfHand(h)
	res := waitsResponse{Shanten: int(h.Shanten()), Waits: []waitEntry{}, Ukeire: []ukeireEntry{}}
	if h.IsTenpai() {
		res.Waits = waitEntries(u)
	} else {
		res.Ukeire = ukeireEntries(u)
	}
	return res, nil
}

func fu(req winningRequest) (fuResponse, error) {
	w, err := req.winning()
	if err != nil {
		return fuResponse{}, err
	}
	return fuResponseOf(w.Fu()), nil
}

func score(req winningRequest) (scoreResponse, error) {
	w, err := req.winning()
	if err != nil {
		return scoreResponse{}, err
	}
	return scoreResponseOf(w.Score(req.DoraCount), req.DoraCount), nil
}

type discardOption struct {
	DiscardedTile string        `json:"discarded_tile"`
	Shanten       int           `json:"shanten"`
	Ukeire        []ukeireEntry `json:"ukeire"`
}

type discardAnalysisResponse struct {
	CurrentShanten int             `json:"current_shanten"`
	Options        []discardOption `json:"options"`
}

// discardAnalysis takes a drawn hand (14 - 3N tiles, the last being the
// draw) and compares every discard.
func discardAnalysis(req handRequest) (discardAnalysisResponse, error) {
	tiles, err := tile.ParseAll(req.ClosedTiles)
	if err != nil {
		return discardAnalysisResponse{}, err
	}
	melds, err := req.melds()
	if err != nil {
		return discardAnalysisResponse{}, err
	}
	c, err := ukeire.CompareDrawnTiles(tiles, melds)
	if err != nil {
		return discardAnalysisResponse{}, err
	}
	res := discardAnalysisResponse{CurrentShanten: int(c.Shanten()), Options: []discardOption{}}
	for _, cand := range c.Candidates() {
		res.Options = append(res.Options, discardOption{DiscardedTile: cand.Tile.String(), Shanten: int(cand.Shanten), Ukeire: ukeireEntries(cand.Ukeire)})
	}
	return res, nil
}
