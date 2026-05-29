package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/metrics"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestHealthz_OK(t *testing.T) {
	pool := itest.NewPostgres(t)
	h := metrics.Healthz(pool)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	require.Equal(t, http.StatusOK, rec.Code)
}
