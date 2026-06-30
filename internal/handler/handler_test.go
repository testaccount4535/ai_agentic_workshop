package handler_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testaccount4535/ai_agentic_workshop/internal/handler"
	"github.com/testaccount4535/ai_agentic_workshop/internal/store"
)

func newTestServer(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(path, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, st.Close()) })
	return handler.New(st, nil).Routes(), st
}

func post(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const validBody = `{"id":"ride-1","driver_id":"driver-1","kind":"shared","time":"2026-06-25T12:00:00Z","request_time":"2026-06-25T11:58:00Z"}`

// Pass criterion 1: posting a ride start writes to the db.
func TestStartRide_Success(t *testing.T) {
	h, st := newTestServer(t)

	rec := post(t, h, validBody)
	require.Equal(t, http.StatusCreated, rec.Code)

	ride, found, err := st.GetRideStart("ride-1")
	require.NoError(t, err)
	require.True(t, found, "ride should be persisted to db")
	require.Equal(t, "driver-1", ride.DriverID)
}

// Pass criterion 2: posting the same id twice fails.
func TestStartRide_DuplicateConflict(t *testing.T) {
	h, _ := newTestServer(t)

	require.Equal(t, http.StatusCreated, post(t, h, validBody).Code)

	rec := post(t, h, validBody)
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "already exists")
}

// Pass criterion 3: invalid data is rejected.
func TestStartRide_RejectsInvalid(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing id", `{"driver_id":"d1","kind":"shared","time":"2026-06-25T12:00:00Z"}`},
		{"missing driver", `{"id":"r1","kind":"shared","time":"2026-06-25T12:00:00Z"}`},
		{"invalid kind", `{"id":"r1","driver_id":"d1","kind":"carpool","time":"2026-06-25T12:00:00Z"}`},
		{"missing time", `{"id":"r1","driver_id":"d1","kind":"shared","request_time":"2026-06-25T11:58:00Z"}`},
		{"missing request_time", `{"id":"r1","driver_id":"d1","kind":"shared","time":"2026-06-25T12:00:00Z"}`},
		{"malformed json", `{"id":`},
		{"unknown field", `{"id":"r1","driver_id":"d1","kind":"shared","time":"2026-06-25T12:00:00Z","foo":1}`},
		{"empty body", ``},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, st := newTestServer(t)
			rec := post(t, h, tc.body)
			require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())

			// Nothing should have been written for invalid input.
			_, found, err := st.GetRideStart("r1")
			require.NoError(t, err)
			require.False(t, found)
		})
	}
}

func TestStartRide_RejectsWrongMethod(t *testing.T) {
	h, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/rides", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
