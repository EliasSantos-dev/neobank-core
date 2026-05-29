package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/api"
	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/EliasSantos-dev/neobank-core/internal/metrics"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func newSrv(t *testing.T) *httptest.Server {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	gw := gateway.NewService(s, gateway.NewFakeProvider())
	srv := httptest.NewServer(api.NewServer(s, []byte("test-secret"), "admin-token", gw, fx.NewFakeRateProvider(50), metrics.New(), pool))
	t.Cleanup(srv.Close)
	return srv
}

func register(t *testing.T, base, email, pass string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"email": email, "password": pass})
	resp, err := http.Post(base+"/auth/register", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func login(t *testing.T, base, email, pass string) string {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"email": email, "password": pass})
	resp, err := http.Post(base+"/auth/login", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.NotEmpty(t, out.AccessToken)
	return out.AccessToken
}

func TestRegisterLoginMe(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")

	req, _ := http.NewRequest("GET", srv.URL+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me struct {
		Email   string `json:"email"`
		Balance int64  `json:"balance"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.Equal(t, "a@b.com", me.Email)
	require.EqualValues(t, 0, me.Balance)
}

func TestRegisterDuplicateEmail(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	b, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "segredo123"})
	resp, err := http.Post(srv.URL+"/auth/register", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestMeRequiresAuth(t *testing.T) {
	srv := newSrv(t)
	resp, err := http.Get(srv.URL + "/me")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLoginBadCredentials(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	b, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "errada99"})
	resp, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
