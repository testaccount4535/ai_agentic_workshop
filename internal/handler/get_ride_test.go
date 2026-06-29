package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// Getting a started ride by id returns it.
func TestGetRide_Success(t *testing.T) {
	h, st := newTestServer(t)
	seedRideStart(t, st, "ride-1")

	rec := get(t, h, "/rides/ride-1")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	require.Contains(t, rec.Body.String(), `"id":"ride-1"`)
	require.Contains(t, rec.Body.String(), `"driver_id":"driver-1"`)
}

// Getting an unknown id returns 404.
func TestGetRide_NotFound(t *testing.T) {
	h, _ := newTestServer(t)

	rec := get(t, h, "/rides/does-not-exist")
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "not found")
}

func TestGetRide_RejectsWrongMethod(t *testing.T) {
	h, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodDelete, "/rides/ride-1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
