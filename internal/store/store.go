// Package store provides bbolt-backed persistence for ride hailing data.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/testaccount4535/ai_agentic_workshop/internal/model"
)

// ErrDuplicateRide is returned when a record is saved with an ID that already
// exists in the relevant bucket (e.g. starting or ending the same ride twice).
var ErrDuplicateRide = errors.New("ride already exists")

// ErrRideNotStarted is returned when ending a ride that has no corresponding
// ride start on record.
var ErrRideNotStarted = errors.New("ride has not been started")

// Buckets holding records keyed by ride ID.
var (
	bucketRideStarts = []byte("ride_starts")
	bucketRideEnds   = []byte("ride_ends")
)

// Store wraps a bbolt database and exposes ride persistence operations.
type Store struct {
	db  *bolt.DB
	log *slog.Logger
}

// Open opens (or creates) a bbolt database at path and ensures required buckets
// exist. The caller is responsible for calling Close.
func Open(path string, log *slog.Logger) (*Store, error) {
	if log == nil {
		log = slog.Default()
	}

	log.Info("opening database", "path", path)
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt at %q: %w", path, err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketRideStarts, bucketRideEnds} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create buckets: %w", err)
	}

	return &Store{db: db, log: log}, nil
}

// Close releases the underlying database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveRideStart persists a ride start. It returns ErrDuplicateRide if a ride
// with the same ID already exists, leaving the existing record untouched.
func (s *Store) SaveRideStart(r model.RideStart) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal ride %q: %w", r.ID, err)
	}

	key := []byte(r.ID)
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRideStarts)
		if b == nil {
			return fmt.Errorf("bucket %q missing", bucketRideStarts)
		}
		if existing := b.Get(key); existing != nil {
			return ErrDuplicateRide
		}
		return b.Put(key, data)
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateRide) {
			s.log.Warn("rejected duplicate ride start", "ride_id", r.ID)
			return err
		}
		s.log.Error("failed to save ride start", "ride_id", r.ID, "error", err)
		return fmt.Errorf("save ride %q: %w", r.ID, err)
	}

	s.log.Info("saved ride start", "ride_id", r.ID, "driver_id", r.DriverID, "kind", r.Kind)
	return nil
}

// GetRideStart fetches a ride start by ID. The boolean result is false when no
// ride exists for the given ID.
func (s *Store) GetRideStart(id string) (model.RideStart, bool, error) {
	var (
		ride  model.RideStart
		found bool
	)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRideStarts)
		if b == nil {
			return fmt.Errorf("bucket %q missing", bucketRideStarts)
		}
		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}
		found = true
		return json.Unmarshal(data, &ride)
	})
	if err != nil {
		return model.RideStart{}, false, fmt.Errorf("get ride %q: %w", id, err)
	}
	return ride, found, nil
}

// SaveRideEnd persists a ride end. It returns ErrRideNotStarted if no ride start
// exists for the ID, and ErrDuplicateRide if the ride has already been ended.
// Both checks and the write happen in a single transaction so concurrent ends
// of the same ride cannot both succeed.
func (s *Store) SaveRideEnd(r model.RideEnd) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal ride end %q: %w", r.ID, err)
	}

	key := []byte(r.ID)
	err = s.db.Update(func(tx *bolt.Tx) error {
		starts := tx.Bucket(bucketRideStarts)
		ends := tx.Bucket(bucketRideEnds)
		if starts == nil || ends == nil {
			return fmt.Errorf("buckets %q/%q missing", bucketRideStarts, bucketRideEnds)
		}
		if starts.Get(key) == nil {
			return ErrRideNotStarted
		}
		if existing := ends.Get(key); existing != nil {
			return ErrDuplicateRide
		}
		return ends.Put(key, data)
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrRideNotStarted):
			s.log.Warn("rejected ride end for ride that was never started", "ride_id", r.ID)
			return err
		case errors.Is(err, ErrDuplicateRide):
			s.log.Warn("rejected duplicate ride end", "ride_id", r.ID)
			return err
		default:
			s.log.Error("failed to save ride end", "ride_id", r.ID, "error", err)
			return fmt.Errorf("save ride end %q: %w", r.ID, err)
		}
	}

	s.log.Info("saved ride end", "ride_id", r.ID, "distance", r.Distance)
	return nil
}

// GetRideEnd fetches a ride end by ID. The boolean result is false when no
// ride end exists for the given ID.
func (s *Store) GetRideEnd(id string) (model.RideEnd, bool, error) {
	var (
		ride  model.RideEnd
		found bool
	)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRideEnds)
		if b == nil {
			return fmt.Errorf("bucket %q missing", bucketRideEnds)
		}
		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}
		found = true
		return json.Unmarshal(data, &ride)
	})
	if err != nil {
		return model.RideEnd{}, false, fmt.Errorf("get ride end %q: %w", id, err)
	}
	return ride, found, nil
}
