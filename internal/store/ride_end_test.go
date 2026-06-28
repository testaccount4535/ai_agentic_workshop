package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testaccount4535/ai_agentic_workshop/internal/model"
)

func sampleRideEnd() model.RideEnd {
	return model.RideEnd{
		ID:       "ride-1",
		Time:     time.Date(2026, 6, 25, 12, 30, 0, 0, time.UTC),
		Distance: 4.2,
	}
}

func TestSaveRideEnd_Persists(t *testing.T) {
	s := newTestStore(t)
	require.NoError(t, s.SaveRideStart(sampleRide()))

	end := sampleRideEnd()
	require.NoError(t, s.SaveRideEnd(end))

	got, found, err := s.GetRideEnd(end.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, end.ID, got.ID)
	require.Equal(t, end.Distance, got.Distance)
	require.True(t, end.Time.Equal(got.Time))
}

func TestSaveRideEnd_RequiresStart(t *testing.T) {
	s := newTestStore(t)

	err := s.SaveRideEnd(sampleRideEnd())
	require.ErrorIs(t, err, ErrRideNotStarted)

	_, found, err := s.GetRideEnd("ride-1")
	require.NoError(t, err)
	require.False(t, found, "nothing should be persisted when the ride was never started")
}

func TestSaveRideEnd_DuplicateFails(t *testing.T) {
	s := newTestStore(t)
	require.NoError(t, s.SaveRideStart(sampleRide()))
	require.NoError(t, s.SaveRideEnd(sampleRideEnd()))

	err := s.SaveRideEnd(sampleRideEnd())
	require.ErrorIs(t, err, ErrDuplicateRide)
}

func TestSaveRideEnd_DuplicateDoesNotOverwrite(t *testing.T) {
	s := newTestStore(t)
	require.NoError(t, s.SaveRideStart(sampleRide()))
	require.NoError(t, s.SaveRideEnd(sampleRideEnd()))

	updated := sampleRideEnd()
	updated.Distance = 99.9
	require.ErrorIs(t, s.SaveRideEnd(updated), ErrDuplicateRide)

	got, found, err := s.GetRideEnd("ride-1")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 4.2, got.Distance, "existing record must be untouched")
}

func TestGetRideEnd_NotFound(t *testing.T) {
	s := newTestStore(t)
	require.NoError(t, s.SaveRideStart(sampleRide()))

	_, found, err := s.GetRideEnd("ride-1")
	require.NoError(t, err)
	require.False(t, found, "a started but not-yet-ended ride has no end record")
}
