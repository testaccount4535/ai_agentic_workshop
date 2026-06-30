package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testaccount4535/ai_agentic_workshop/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, s.Close()) })
	return s
}

func sampleRide() model.RideStart {
	return model.RideStart{
		ID:          "ride-1",
		DriverID:    "driver-1",
		Kind:        model.RideKindShared,
		Time:        time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
		RequestTime: time.Date(2026, 6, 25, 11, 58, 0, 0, time.UTC),
	}
}

func TestSaveRideStart_Persists(t *testing.T) {
	s := newTestStore(t)
	ride := sampleRide()

	require.NoError(t, s.SaveRideStart(ride))

	got, found, err := s.GetRideStart(ride.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ride.ID, got.ID)
	require.Equal(t, ride.DriverID, got.DriverID)
	require.Equal(t, ride.Kind, got.Kind)
	require.True(t, ride.Time.Equal(got.Time))
}

func TestSaveRideStart_DuplicateFails(t *testing.T) {
	s := newTestStore(t)
	ride := sampleRide()

	require.NoError(t, s.SaveRideStart(ride))

	err := s.SaveRideStart(ride)
	require.ErrorIs(t, err, ErrDuplicateRide)
}

func TestSaveRideStart_DuplicateDoesNotOverwrite(t *testing.T) {
	s := newTestStore(t)
	ride := sampleRide()
	require.NoError(t, s.SaveRideStart(ride))

	updated := ride
	updated.DriverID = "driver-2"
	require.ErrorIs(t, s.SaveRideStart(updated), ErrDuplicateRide)

	got, found, err := s.GetRideStart(ride.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "driver-1", got.DriverID, "existing record must be untouched")
}

func TestGetRideStart_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, found, err := s.GetRideStart("missing")
	require.NoError(t, err)
	require.False(t, found)
}

func TestErrDuplicateIsDistinct(t *testing.T) {
	require.False(t, errors.Is(ErrDuplicateRide, model.ErrInvalidRide))
}
