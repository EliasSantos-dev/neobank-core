package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/api"
	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/EliasSantos-dev/neobank-core/internal/risk"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/internal/worker"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func newSrvWithWorker(t *testing.T) (*httptest.Server, *worker.Worker) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	gw := gateway.NewService(s, gateway.NewFakeProvider())
	srv := httptest.NewServer(api.NewServer(s, []byte("test-secret"), "admin-token", gw, fx.NewFakeRateProvider(50)))
	t.Cleanup(srv.Close)
	return srv, worker.New(s, risk.NewEngine(risk.RuleReasoningAdvisor{}, 50))
}

func strReader(s string) *strings.Reader { return strings.NewReader(s) }

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
}

func TestE2E_LowRiskAutoCompletes(t *testing.T) {
	srv, w := newSrvWithWorker(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	register(t, srv.URL, "c@d.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 5000)
	authedPost(t, srv.URL, "/transfers", tok, map[string]any{"to_email": "c@d.com", "amount": 100}).Body.Close()

	_, err := w.ProcessOnce(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 4900, meBalance(t, srv.URL, tok))
}

func TestE2E_HighRiskReviewApprove(t *testing.T) {
	srv, w := newSrvWithWorker(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	register(t, srv.URL, "c@d.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 10000)
	// drena 95% + destinatário novo -> high
	authedPost(t, srv.URL, "/transfers", tok, map[string]any{"to_email": "c@d.com", "amount": 9500}).Body.Close()
	_, _ = w.ProcessOnce(context.Background())

	req, _ := http.NewRequest("GET", srv.URL+"/me/transfers", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var intents []struct {
		ID     string `json:"ID"`
		Status string `json:"Status"`
	}
	decodeJSON(t, resp, &intents)
	require.Len(t, intents, 1)
	require.Equal(t, "under_review", intents[0].Status)

	rr, _ := http.NewRequest("POST", srv.URL+"/admin/transfers/"+intents[0].ID+"/review", strReader(`{"decision":"approve"}`))
	rr.Header.Set("X-Admin-Token", "admin-token")
	resp2, err := http.DefaultClient.Do(rr)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	require.EqualValues(t, 500, meBalance(t, srv.URL, tok)) // 10000-9500 efetivou
}

func TestE2E_HighRiskReviewReject(t *testing.T) {
	srv, w := newSrvWithWorker(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	register(t, srv.URL, "c@d.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 10000)
	authedPost(t, srv.URL, "/transfers", tok, map[string]any{"to_email": "c@d.com", "amount": 9500}).Body.Close()
	_, _ = w.ProcessOnce(context.Background())

	req, _ := http.NewRequest("GET", srv.URL+"/me/transfers", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var intents []struct {
		ID string `json:"ID"`
	}
	decodeJSON(t, resp, &intents)
	require.Len(t, intents, 1)

	rr, _ := http.NewRequest("POST", srv.URL+"/admin/transfers/"+intents[0].ID+"/review", strReader(`{"decision":"reject"}`))
	rr.Header.Set("X-Admin-Token", "admin-token")
	resp2, err := http.DefaultClient.Do(rr)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	require.EqualValues(t, 10000, meBalance(t, srv.URL, tok)) // nada se moveu
}

func TestE2E_AdminTokenRequired(t *testing.T) {
	srv, _ := newSrvWithWorker(t)
	rr, _ := http.NewRequest("POST", srv.URL+"/admin/transfers/00000000-0000-0000-0000-000000000000/review", strReader(`{"decision":"approve"}`))
	resp, err := http.DefaultClient.Do(rr)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
