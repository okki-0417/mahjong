package mahjongd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/okki-0417/mahjong/internal/mahjongd"
)

func post(t *testing.T, h http.Handler, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var v map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("%s: %v: %s", path, err, rec.Body.String())
	}
	return rec.Code, v
}

func newHandler() http.Handler {
	return mahjongd.NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestShanten(t *testing.T) {
	h := newHandler()
	t.Run("answers the shanten of a hand", func(t *testing.T) {
		status, v := post(t, h, "/v1/shanten", `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","4p","5s"],"open_melds":[]}`)
		if status != 200 || v["shanten"] != float64(1) {
			t.Fatalf("%d %v", status, v)
		}
	})
	t.Run("counts melds", func(t *testing.T) {
		status, v := post(t, h, "/v1/shanten", `{"closed_tiles":["4m","5m","6m","7m","8m","9m","1p","2p","3p","5s"],"open_melds":[{"kind":"chi","tiles":["1m","2m","3m"]}]}`)
		if status != 200 || v["shanten"] != float64(0) {
			t.Fatalf("%d %v", status, v)
		}
	})
	t.Run("rejects a bad tile with 422", func(t *testing.T) {
		status, v := post(t, h, "/v1/shanten", `{"closed_tiles":["9z"]}`)
		if status != 422 || v["error"] == nil {
			t.Fatalf("%d %v", status, v)
		}
	})
	t.Run("rejects malformed JSON with 400", func(t *testing.T) {
		status, _ := post(t, h, "/v1/shanten", `{`)
		if status != 400 {
			t.Fatalf("%d", status)
		}
	})
	t.Run("rejects the wrong tile count with 422", func(t *testing.T) {
		status, _ := post(t, h, "/v1/shanten", `{"closed_tiles":["1m","2m"]}`)
		if status != 422 {
			t.Fatalf("%d", status)
		}
	})
}

func TestUkeireAndWaits(t *testing.T) {
	h := newHandler()
	tenpai := `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","3p","5s"]}`
	t.Run("ukeire lists tiles with counts", func(t *testing.T) {
		status, v := post(t, h, "/v1/ukeire", tenpai)
		entries := v["ukeire"].([]any)
		if status != 200 || v["shanten"] != float64(0) || len(entries) != 1 || entries[0].(map[string]any)["count"] != float64(3) {
			t.Fatalf("%d %v", status, v)
		}
	})
	t.Run("waits of a tenpai hand carry their kinds and leave ukeire empty", func(t *testing.T) {
		status, v := post(t, h, "/v1/waits", tenpai)
		waits := v["waits"].([]any)
		if status != 200 || len(waits) != 1 || len(v["ukeire"].([]any)) != 0 {
			t.Fatalf("%d %v", status, v)
		}
		if kinds := waits[0].(map[string]any)["kinds"].([]any); len(kinds) != 1 || kinds[0] != "tanki" {
			t.Fatalf("kinds %v", kinds)
		}
	})
	t.Run("a hand that is not tenpai answers ukeire and no waits", func(t *testing.T) {
		status, v := post(t, h, "/v1/waits", `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","4p","5s"]}`)
		if status != 200 || len(v["waits"].([]any)) != 0 || len(v["ukeire"].([]any)) == 0 {
			t.Fatalf("%d %v", status, v)
		}
	})
}

func TestFuAndScore(t *testing.T) {
	h := newHandler()
	body := `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","3p","5s"],"winning_tile":"5s","win_kind":"ron","round_wind":"east","seat_wind":"south","riichi":true,"dora_count":1}`
	t.Run("fu answers subtotal, total, items, and the reading", func(t *testing.T) {
		status, v := post(t, h, "/v1/fu", body)
		if status != 200 || v["total"] != float64(40) || len(v["items"].([]any)) == 0 {
			t.Fatalf("%d %v", status, v)
		}
		parsed := v["parsed_hand"].(map[string]any)
		if parsed["kind"] != "standard" || parsed["wait_kind"] != "tanki" || parsed["pair_tile"] != "5s" {
			t.Fatalf("parsed %v", parsed)
		}
	})
	t.Run("score answers yaku, han, fu, payments, and the reading", func(t *testing.T) {
		status, v := post(t, h, "/v1/score", body)
		if status != 200 || v["han"] != float64(4) || v["fu"] != float64(40) || v["dora_count"] != float64(1) {
			t.Fatalf("%d %v", status, v)
		}
		payments := v["payments"].(map[string]any)
		if payments["from_loser"] != float64(8000) || payments["from_dealer"] != nil {
			t.Fatalf("payments %v", payments)
		}
		if v["rank_label"] != "満貫" {
			t.Fatalf("rank %v", v["rank_label"])
		}
	})
	t.Run("a hand without yaku is 422", func(t *testing.T) {
		status, v := post(t, h, "/v1/score", `{"closed_tiles":["1p","2p","3p","4p","5p","6p","7s","8s","9s","2z"],"open_melds":[{"kind":"chi","tiles":["1m","2m","3m"]}],"winning_tile":"2z","win_kind":"ron","round_wind":"east","seat_wind":"south"}`)
		if status != 422 || v["error"] == nil {
			t.Fatalf("%d %v", status, v)
		}
	})
	t.Run("a contradictory situation is 422", func(t *testing.T) {
		status, _ := post(t, h, "/v1/score", `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","3p","5s"],"winning_tile":"5s","win_kind":"ron","round_wind":"east","seat_wind":"south","ippatsu":true}`)
		if status != 422 {
			t.Fatalf("%d", status)
		}
	})
	t.Run("round-up mangan is a rule switch", func(t *testing.T) {
		near := `{"closed_tiles":["3m","4m","5m","6m","7m","8m","1p","2p","3p","4p","5p","7s","7s"],"winning_tile":"6p","win_kind":"ron","round_wind":"east","seat_wind":"south","riichi":true,"ippatsu":true,"dora_count":1,"round_up_mangan":%v}`
		_, without := post(t, h, "/v1/score", strings.Replace(near, "%v", "false", 1))
		_, with := post(t, h, "/v1/score", strings.Replace(near, "%v", "true", 1))
		if without["total"] != float64(7700) || with["total"] != float64(8000) {
			t.Fatalf("%v / %v", without["total"], with["total"])
		}
	})
}

func TestDiscardAnalysis(t *testing.T) {
	h := newHandler()
	t.Run("compares every discard of a drawn hand", func(t *testing.T) {
		status, v := post(t, h, "/v1/discard-analysis", `{"closed_tiles":["1m","2m","3m","4m","5m","6m","7m","8m","9m","1p","2p","4p","5s","3p"]}`)
		if status != 200 || v["current_shanten"] != float64(0) {
			t.Fatalf("%d %v", status, v)
		}
		options := v["options"].([]any)
		best := options[0].(map[string]any)
		if best["discarded_tile"] != "5s" || best["shanten"] != float64(0) || len(best["ukeire"].([]any)) == 0 {
			t.Fatalf("best %v", best)
		}
	})
	t.Run("rejects a hand that is not a drawn count", func(t *testing.T) {
		status, _ := post(t, h, "/v1/discard-analysis", `{"closed_tiles":["1m","2m","3m"]}`)
		if status != 422 {
			t.Fatalf("%d", status)
		}
	})
}

func TestSimulatorStep(t *testing.T) {
	h := newHandler()

	t.Run("deals a wall when none is given and plays until the human is awaited", func(t *testing.T) {
		status, v := post(t, h, "/v1/simulator/step", `{"user_seat":1}`)
		if status != 200 {
			t.Fatalf("%d %v", status, v)
		}
		if len(v["wall"].([]any)) != 136 || v["finished"] != false {
			t.Fatalf("%v", v)
		}
		awaiting := v["awaiting"].([]any)
		if len(awaiting) == 0 || awaiting[0] != float64(1) {
			t.Fatalf("awaiting %v", awaiting)
		}
		taken := v["taken"].([]any)
		if len(taken) == 0 {
			t.Fatal("the dealer CPU should have played")
		}
		sight := v["sight"].(map[string]any)
		if sight["seat"] != float64(1) || len(sight["legal_actions"].([]any)) == 0 {
			t.Fatalf("sight %v", sight)
		}
	})

	t.Run("replays recorded actions, takes the human's action, and continues", func(t *testing.T) {
		_, first := post(t, h, "/v1/simulator/step", `{"user_seat":0}`)
		wall, _ := json.Marshal(first["wall"])
		sight := first["sight"].(map[string]any)
		legal := sight["legal_actions"].([]any)
		action, _ := json.Marshal(legal[len(legal)-1])
		var body bytes.Buffer
		body.WriteString(`{"user_seat":0,"wall":`)
		body.Write(wall)
		body.WriteString(`,"actions":[],"action":`)
		body.Write(action)
		body.WriteString(`}`)
		status, v := post(t, h, "/v1/simulator/step", body.String())
		if status != 200 {
			t.Fatalf("%d %v", status, v)
		}
		taken := v["taken"].([]any)
		if len(taken) < 1 || taken[0].(map[string]any)["seat"] != float64(0) {
			t.Fatalf("taken %v", taken)
		}
	})

	t.Run("rejects another seat's action with 422", func(t *testing.T) {
		status, v := post(t, h, "/v1/simulator/step", `{"user_seat":0,"action":{"seat":1,"kind":"pass","tiles":[]}}`)
		if status != 422 || v["error"] == nil {
			t.Fatalf("%d %v", status, v)
		}
	})

	t.Run("rejects an illegal action with 422", func(t *testing.T) {
		status, _ := post(t, h, "/v1/simulator/step", `{"user_seat":0,"action":{"seat":0,"kind":"tsumo","tiles":[]}}`)
		if status != 422 {
			t.Fatalf("%d", status)
		}
	})

	t.Run("rejects a bad user seat with 400", func(t *testing.T) {
		status, _ := post(t, h, "/v1/simulator/step", `{"user_seat":4}`)
		if status != 400 {
			t.Fatalf("%d", status)
		}
	})

	t.Run("plays a whole kyoku when the human only tsumogiris", func(t *testing.T) {
		_, v := post(t, h, "/v1/simulator/step", `{"user_seat":0}`)
		wall, _ := json.Marshal(v["wall"])
		var actions []any
		for i := 0; i < 400 && v["finished"] != true; i++ {
			actions = append(actions, v["taken"].([]any)...)
			sight := v["sight"].(map[string]any)
			legal := sight["legal_actions"].([]any)
			var chosen any
			for _, a := range legal {
				kind := a.(map[string]any)["kind"]
				if kind == "tsumo" || kind == "ron" || kind == "pass" {
					chosen = a
					break
				}
			}
			if chosen == nil {
				drawn := sight["drawn"]
				for _, a := range legal {
					m := a.(map[string]any)
					if m["kind"] == "discard" && m["tiles"].([]any)[0] == drawn {
						chosen = a
					}
				}
			}
			if chosen == nil {
				chosen = legal[0]
			}
			recorded, _ := json.Marshal(actions)
			action, _ := json.Marshal(chosen)
			body := `{"user_seat":0,"wall":` + string(wall) + `,"actions":` + string(recorded) + `,"action":` + string(action) + `}`
			var status int
			status, v = post(t, h, "/v1/simulator/step", body)
			if status != 200 {
				t.Fatalf("%d %v", status, v)
			}
		}
		if v["finished"] != true || v["result"] == nil {
			t.Fatalf("did not finish: %v", v)
		}
		if len(actions) < 20 {
			t.Fatalf("only %d actions before the end", len(actions))
		}
	})
}
