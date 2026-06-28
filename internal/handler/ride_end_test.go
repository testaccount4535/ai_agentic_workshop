package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testaccount4535/ai_agentic_workshop/internal/model"
	"github.com/testaccount4535/ai_agentic_workshop/internal/store"
)

func postEnd(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/rides/end", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func seedRideStart(t *testing.T, st *store.Store, id string) {
	t.Helper()
	require.NoError(t, st.SaveRideStart(model.RideStart{
		ID:       id,
		DriverID: "driver-1",
		Kind:     model.RideKindShared,
		Time:     time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
	}))
}

const validEndBody = `{"id":"ride-1","time":"2026-06-25T12:30:00Z","distance":4.2}`

// Posting a ride end for a started ride writes to the db.
func TestEndRide_Success(t *testing.T) {
	h, st := newTestServer(t)
	seedRideStart(t, st, "ride-1")

	rec := postEnd(t, h, validEndBody)
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	end, found, err := st.GetRideEnd("ride-1")
	require.NoError(t, err)
	require.True(t, found, "ride end should be persisted to db")
	require.Equal(t, 4.2, end.Distance)
}

// Ending a ride that was never started is rejected.
func TestEndRide_RequiresStart(t *testing.T) {
	h, st := newTestServer(t)

	rec := postEnd(t, h, validEndBody)
	require.Equal(t, http.StatusNotFound, rec.Code)

	_, found, err := st.GetRideEnd("ride-1")
	require.NoError(t, err)
	require.False(t, found)
}

// Ending the same ride twice fails.
func TestEndRide_DuplicateConflict(t *testing.T) {
	h, st := newTestServer(t)
	seedRideStart(t, st, "ride-1")

	require.Equal(t, http.StatusCreated, postEnd(t, h, validEndBody).Code)

	rec := postEnd(t, h, validEndBody)
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "already exists")
}

// Invalid ride end data is rejected.
func TestEndRide_RejectsInvalid(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing id", `{"time":"2026-06-25T12:30:00Z","distance":4.2}`},
		{"missing time", `{"id":"ride-1","distance":4.2}`},
		{"negative distance", `{"id":"ride-1","time":"2026-06-25T12:30:00Z","distance":-1}`},
		{"malformed json", `{"id":`},
		{"unknown field", `{"id":"ride-1","time":"2026-06-25T12:30:00Z","distance":4.2,"foo":1}`},
		{"empty body", ``},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, st := newTestServer(t)
			seedRideStart(t, st, "ride-1")

			rec := postEnd(t, h, tc.body)
			require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())

			_, found, err := st.GetRideEnd("ride-1")
			require.NoError(t, err)
			require.False(t, found, "nothing should be written for invalid input")
		})
	}
}

func TestEndRide_RejectsWrongMethod(t *testing.T) {
	h, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/rides/end", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
