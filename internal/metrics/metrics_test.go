package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareAndHandler(t *testing.T) {
	m := metrics.New()
	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	m.IncOperation("deposit")

	mrec := httptest.NewRecorder()
	m.Handler().ServeHTTP(mrec, httptest.NewRequest("GET", "/metrics", nil))
	body := mrec.Body.String()
	require.True(t, strings.Contains(body, "neobank_http_requests_total"), body)
	require.True(t, strings.Contains(body, "neobank_operations_total"), body)
}
