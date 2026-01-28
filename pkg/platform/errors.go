package platform

import "errors"

// Sentinel errors for platform operations.
var (
	// ErrClosed is returned when operating on a closed channel or stream.
	ErrClosed = errors.New("platform: channel closed")

	// ErrNotConnected is returned when the platform bridge is not connected.
	ErrNotConnected = errors.New("platform: not connected")
)
