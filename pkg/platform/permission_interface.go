package platform

import "context"

// PermissionStatus represents the current state of a permission.
// This is an alias for PermissionResult for naming consistency.
type PermissionStatus = PermissionResult

// Permission provides access to a runtime permission for a platform feature.
// Use Status to check current state, Request to prompt the user, and Listen
// to observe changes.
//
// Context usage: The ctx parameter is used for cancellation and timeout on
// blocking operations (Request). For non-blocking operations (Status,
// IsGranted, IsDenied, ShouldShowRationale), ctx is accepted for API
// consistency but not currently used for cancellation.
type Permission interface {
	// Status returns the current permission status.
	// Note: ctx is accepted for API consistency but not used for cancellation.
	Status(ctx context.Context) (PermissionStatus, error)

	// Request prompts the user for permission and blocks until they respond
	// or the context is canceled/times out. If already in a terminal state,
	// returns immediately without showing a dialog.
	Request(ctx context.Context) (PermissionStatus, error)

	// IsGranted returns true if permission is granted.
	// Best-effort convenience: returns false on any error. Use Status for
	// precise error handling when error details matter.
	IsGranted(ctx context.Context) bool

	// IsDenied returns true if permission is denied or permanently denied.
	// Best-effort convenience: returns false on any error. Use Status for
	// precise error handling when error details matter.
	IsDenied(ctx context.Context) bool

	// ShouldShowRationale returns whether to show a rationale before requesting.
	// Android-specific; always returns (false, nil) on iOS.
	ShouldShowRationale(ctx context.Context) (bool, error)

	// Listen subscribes to permission status changes.
	// Returns an unsubscribe function. Multiple listeners receive all events.
	Listen(handler func(PermissionStatus)) (unsubscribe func())
}
