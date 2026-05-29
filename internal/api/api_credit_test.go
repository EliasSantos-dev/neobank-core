package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreditEndpoint(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 50_000)

	req, _ := http.NewRequest("GET", srv.URL+"/me/credit", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Score int    `json:"score"`
		Band  string `json:"band"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.GreaterOrEqual(t, out.Score, 0)
	require.NotEmpty(t, out.Band)
}

func TestMetricsAndHealthz(t *testing.T) {
	srv := newSrv(t)
	http.Get(srv.URL + "/healthz")

	mresp, err := http.Get(srv.URL + "/metrics")
	require.NoError(t, err)
	defer mresp.Body.Close()
	require.Equal(t, http.StatusOK, mresp.StatusCode)

	hresp, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	defer hresp.Body.Close()
	require.Equal(t, http.StatusOK, hresp.StatusCode)
}
