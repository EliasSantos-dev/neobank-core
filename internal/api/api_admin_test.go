package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReconcileRequiresAdmin(t *testing.T) {
	srv, _ := newSrvWithWorker(t)
	req, _ := http.NewRequest("POST", srv.URL+"/admin/reconcile", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestReconcileHealthy(t *testing.T) {
	srv, _ := newSrvWithWorker(t)
	req, _ := http.NewRequest("POST", srv.URL+"/admin/reconcile", nil)
	req.Header.Set("X-Admin-Token", "admin-token")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
