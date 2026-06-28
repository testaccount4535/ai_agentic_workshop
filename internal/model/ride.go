// Package model defines the core domain types for ride hailing data.
package model

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidRide is the sentinel error returned when ride data fails validation.
// Callers can match it with errors.Is to translate validation failures into a
// 400-class response without depending on the concrete error text.
var ErrInvalidRide = errors.New("invalid ride")

// RideKind enumerates the supported ride sharing modes.
type RideKind string

const (
	RideKindShared  RideKind = "shared"
	RideKindPrivate RideKind = "private"
)

// Valid reports whether the kind is one of the known ride kinds.
func (k RideKind) Valid() bool {
	switch k {
	case RideKindShared, RideKindPrivate:
		return true
	default:
		return false
	}
}

// RideStart captures the information recorded when a ride begins.
type RideStart struct {
	ID       string    `json:"id"`
	DriverID string    `json:"driver_id"`
	Kind     RideKind  `json:"kind"`
	Time     time.Time `json:"time"`
}

// Validate ensures every field carries a sensible value. All failures wrap
// ErrInvalidRide so the transport layer can map them to a single status code.
func (r RideStart) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidRide)
	}
	if r.DriverID == "" {
		return fmt.Errorf("%w: driver_id is required", ErrInvalidRide)
	}
	if !r.Kind.Valid() {
		return fmt.Errorf("%w: kind must be %q or %q, got %q", ErrInvalidRide, RideKindShared, RideKindPrivate, r.Kind)
	}
	if r.Time.IsZero() {
		return fmt.Errorf("%w: time is required", ErrInvalidRide)
	}
	return nil
}

// RideEnd captures the information recorded when a ride completes. ID refers to
// the ID of the RideStart this end belongs to.
type RideEnd struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Distance float64   `json:"distance"`
}

// Validate ensures every field carries a sensible value. All failures wrap
// ErrInvalidRide so the transport layer can map them to a single status code.
// A zero distance is allowed (e.g. a ride cancelled at the pickup point); a
// negative distance is not.
func (r RideEnd) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidRide)
	}
	if r.Time.IsZero() {
		return fmt.Errorf("%w: time is required", ErrInvalidRide)
	}
	if r.Distance < 0 {
		return fmt.Errorf("%w: distance must not be negative, got %v", ErrInvalidRide, r.Distance)
	}
	return nil
}
