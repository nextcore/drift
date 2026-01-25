package platform

import (
	"context"
	"sync"
	"time"

	"github.com/go-drift/drift/pkg/errors"
)

// PermissionResult represents the status of a permission.
type PermissionResult string

// Permission status constants.
const (
	// PermissionGranted indicates full access has been granted.
	PermissionGranted PermissionResult = "granted"

	// PermissionDenied indicates the user denied the permission. The app may request again.
	PermissionDenied PermissionResult = "denied"

	// PermissionPermanentlyDenied indicates the user denied with "don't ask again" (Android)
	// or denied twice (iOS). The app cannot request again; direct user to Settings.
	PermissionPermanentlyDenied PermissionResult = "permanently_denied"

	// PermissionRestricted indicates a system policy prevents granting (parental controls,
	// MDM, enterprise policy). The user cannot change this; no dialog will be shown.
	PermissionRestricted PermissionResult = "restricted"

	// PermissionLimited indicates partial access (iOS only). For Photos, this means the user
	// selected specific photos rather than granting full library access.
	PermissionLimited PermissionResult = "limited"

	// PermissionNotDetermined indicates the user has not yet been asked. Calling Request()
	// will show the system permission dialog.
	PermissionNotDetermined PermissionResult = "not_determined"

	// PermissionProvisional indicates provisional notification permission (iOS only).
	// Notifications are delivered quietly to Notification Center without alerting the user.
	PermissionProvisional PermissionResult = "provisional"

	// PermissionResultUnknown indicates the status could not be determined.
	PermissionResultUnknown PermissionResult = "unknown"
)

// NotificationOptions configures which notification capabilities to request (iOS only).
// On Android, all capabilities are granted together; these options are ignored.
type NotificationOptions struct {
	// Alert enables visible notifications (banners, alerts). Default: true.
	Alert bool
	// Sound enables notification sounds. Default: true.
	Sound bool
	// Badge enables badge count updates on the app icon. Default: true.
	Badge bool
	// Provisional requests provisional authorization (iOS 12+). Notifications are
	// delivered quietly to Notification Center. The user can later promote or disable.
	Provisional bool
}

// DefaultPermissionTimeout is the default timeout for permission requests.
const DefaultPermissionTimeout = 30 * time.Second

// isTerminalStatus returns true if the status is a terminal state that won't change
// from showing a permission dialog. This includes:
//   - granted: permission already granted
//   - permanently_denied: user denied with "don't ask again" (Android) or denied twice (iOS)
//   - restricted: system policy prevents granting (parental controls, MDM, etc.)
//   - limited: partial access granted (e.g., iOS Photos with selected photos only)
//   - provisional: provisional notifications granted (iOS); no upgrade prompt available
func isTerminalStatus(status PermissionResult) bool {
	switch status {
	case PermissionGranted, PermissionPermanentlyDenied, PermissionRestricted,
		PermissionLimited, PermissionProvisional:
		return true
	default:
		return false
	}
}

// Permissions provides access to runtime permission management.
// Each permission type offers Status(), Request(), IsGranted(), IsDenied(),
// ShouldShowRationale(), and Changes() methods.
var Permissions = struct {
	Camera       *permissionType
	Microphone   *permissionType
	Photos       *permissionType
	Location     *locationPermissionType
	Contacts     *permissionType
	Calendar     *permissionType
	Storage      *permissionType
	Notification *notificationPermissionType
}{
	Camera:       newPermission("camera"),
	Microphone:   newPermission("microphone"),
	Photos:       newPermission("photos"),
	Location:     newLocationPermission(),
	Contacts:     newPermission("contacts"),
	Calendar:     newPermission("calendar"),
	Storage:      newPermission("storage"),
	Notification: newNotificationPermission(),
}

var (
	permissionChangesOnce    sync.Once
	permissionChangesChannel *EventChannel
)

func getPermissionChangesChannel() *EventChannel {
	permissionChangesOnce.Do(func() {
		permissionChangesChannel = NewEventChannel("drift/permissions/changes")
	})
	return permissionChangesChannel
}

// permissionType provides methods for checking and requesting a single permission.
type permissionType struct {
	name    string
	channel *MethodChannel
	changes *EventChannel

	// Mutex to serialize permission requests (only one dialog can be shown at a time)
	requestMu sync.Mutex

	// Shared channel for Changes() - initialized once, returned to all callers
	changesOnce sync.Once
	changesChan chan PermissionResult
}

func newPermission(name string) *permissionType {
	return &permissionType{
		name:    name,
		channel: NewMethodChannel("drift/permissions"),
		changes: getPermissionChangesChannel(),
	}
}

// Status returns the current status of the permission.
func (p *permissionType) Status() (PermissionResult, error) {
	result, err := p.channel.Invoke("check", map[string]any{
		"permission": p.name,
	})
	if err != nil {
		return PermissionResultUnknown, err
	}
	return parsePermissionResult(result), nil
}

// Request requests the permission from the user and blocks until the user responds
// or the default timeout (DefaultPermissionTimeout) is exceeded.
//
// If the permission is already in a terminal state (granted, permanently_denied,
// restricted, limited, or provisional), returns immediately without showing a dialog.
//
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the user did not respond within the timeout period
//   - Other errors: platform communication failures
func (p *permissionType) Request() (PermissionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPermissionTimeout)
	defer cancel()
	return p.RequestWithContext(ctx)
}

// RequestWithContext requests the permission from the user and blocks until the user
// responds, the context is canceled, or the context deadline is exceeded.
//
// If the permission is already in a terminal state (granted, permanently_denied,
// restricted, limited, or provisional), returns immediately without showing a dialog.
//
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the context deadline was exceeded
//   - ErrCanceled: the context was canceled
//   - Other errors: platform communication failures
func (p *permissionType) RequestWithContext(ctx context.Context) (PermissionResult, error) {
	p.requestMu.Lock()
	defer p.requestMu.Unlock()

	// Return immediately if already in terminal state
	currentStatus, err := p.Status()
	if err != nil {
		return PermissionResultUnknown, err
	}
	if isTerminalStatus(currentStatus) {
		return currentStatus, nil
	}

	// Subscribe BEFORE triggering native request to avoid race condition
	resultChan := make(chan PermissionResult, 1)
	sub := p.changes.Listen(EventHandler{
		OnEvent: func(data any) {
			change, ok := parsePermissionChange(data)
			if ok && change.Permission == p.name {
				select {
				case resultChan <- change.Result:
				default:
				}
			}
		},
		OnError: func(err error) {
			errors.Report(&errors.DriftError{
				Op:      "permissions.request",
				Kind:    errors.KindPlatform,
				Channel: "drift/permissions/changes",
				Err:     err,
			})
		},
	})
	defer sub.Cancel()

	// Trigger native request
	_, err = p.channel.Invoke("request", map[string]any{"permission": p.name})
	if err != nil {
		return PermissionResultUnknown, err
	}

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		// Re-check status in case we missed the event
		if finalStatus, err := p.Status(); err == nil && isTerminalStatus(finalStatus) {
			return finalStatus, nil
		}
		if ctx.Err() == context.DeadlineExceeded {
			return PermissionResultUnknown, ErrTimeout
		}
		return PermissionResultUnknown, ErrCanceled
	}
}

// IsGranted returns true if the permission is currently granted.
func (p *permissionType) IsGranted() bool {
	status, err := p.Status()
	if err != nil {
		return false
	}
	return status == PermissionGranted
}

// IsDenied returns true if the permission is denied or permanently denied.
func (p *permissionType) IsDenied() bool {
	status, err := p.Status()
	if err != nil {
		return false
	}
	return status == PermissionDenied || status == PermissionPermanentlyDenied
}

// ShouldShowRationale returns whether the app should show a rationale for
// requesting this permission. This is Android-specific and always returns
// false on iOS.
func (p *permissionType) ShouldShowRationale() bool {
	result, err := p.channel.Invoke("shouldShowRationale", map[string]any{
		"permission": p.name,
	})
	if err != nil {
		return false
	}
	if m, ok := result.(map[string]any); ok {
		return parseBool(m["shouldShow"])
	}
	return false
}

// Changes returns a channel that receives permission status changes for this permission.
// The same channel is returned on each call. Only one goroutine should read from it;
// if multiple goroutines read, events will be split between them (not broadcast).
// The channel has a small buffer; slow consumers may miss events.
func (p *permissionType) Changes() <-chan PermissionResult {
	p.changesOnce.Do(func() {
		p.changesChan = make(chan PermissionResult, 4)
		p.changes.Listen(EventHandler{
			OnEvent: func(data any) {
				change, ok := parsePermissionChange(data)
				if !ok {
					errors.Report(&errors.DriftError{
						Op:      "permissions.parseChange",
						Kind:    errors.KindParsing,
						Channel: "drift/permissions/changes",
						Err: &errors.ParseError{
							Channel:  "drift/permissions/changes",
							DataType: "PermissionChange",
							Got:      data,
						},
					})
					return
				}
				// Only send changes for this specific permission
				if change.Permission == p.name {
					select {
					case p.changesChan <- change.Result:
					default:
						// Channel full, skip
					}
				}
			},
			OnError: func(err error) {
				errors.Report(&errors.DriftError{
					Op:      "permissions.streamError",
					Kind:    errors.KindPlatform,
					Channel: "drift/permissions/changes",
					Err:     err,
				})
			},
		})
	})

	return p.changesChan
}

// locationPermissionType extends permissionType with location-specific methods.
type locationPermissionType struct {
	*permissionType
	alwaysChannel *MethodChannel

	// Mutex to serialize always permission requests
	alwaysRequestMu sync.Mutex

	// Shared channel for ChangesAlways() - initialized once, returned to all callers
	alwaysChangesOnce sync.Once
	alwaysChangesChan chan PermissionResult
}

func newLocationPermission() *locationPermissionType {
	return &locationPermissionType{
		permissionType: newPermission("location"),
		alwaysChannel:  NewMethodChannel("drift/permissions"),
	}
}

// RequestWhenInUse requests permission to access location while the app is in use
// and blocks until the user responds or the default timeout is exceeded.
//
// This is equivalent to calling Request() on the location permission.
// See [permissionType.Request] for full documentation on blocking behavior and errors.
func (l *locationPermissionType) RequestWhenInUse() (PermissionResult, error) {
	return l.Request()
}

// RequestAlways requests permission to access location in the background ("always")
// and blocks until the user responds or the default timeout is exceeded.
//
// On iOS, this requires first obtaining "when in use" permission. On Android 10+,
// background location is requested separately from foreground location.
//
// If the permission is already in a terminal state, returns immediately.
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the user did not respond within the timeout period
//   - Other errors: platform communication failures
func (l *locationPermissionType) RequestAlways() (PermissionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPermissionTimeout)
	defer cancel()
	return l.RequestAlwaysWithContext(ctx)
}

// RequestAlwaysWithContext requests permission to access location in the background
// and blocks until the user responds, the context is canceled, or the deadline is exceeded.
//
// If the permission is already in a terminal state, returns immediately.
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the context deadline was exceeded
//   - ErrCanceled: the context was canceled
//   - Other errors: platform communication failures
func (l *locationPermissionType) RequestAlwaysWithContext(ctx context.Context) (PermissionResult, error) {
	l.alwaysRequestMu.Lock()
	defer l.alwaysRequestMu.Unlock()

	// Return immediately if already in terminal state
	currentStatus, err := l.StatusAlways()
	if err != nil {
		return PermissionResultUnknown, err
	}
	if isTerminalStatus(currentStatus) {
		return currentStatus, nil
	}

	// Subscribe BEFORE triggering native request to avoid race condition
	resultChan := make(chan PermissionResult, 1)
	sub := l.changes.Listen(EventHandler{
		OnEvent: func(data any) {
			change, ok := parsePermissionChange(data)
			if ok && change.Permission == "location_always" {
				select {
				case resultChan <- change.Result:
				default:
				}
			}
		},
		OnError: func(err error) {
			errors.Report(&errors.DriftError{
				Op:      "permissions.requestAlways",
				Kind:    errors.KindPlatform,
				Channel: "drift/permissions/changes",
				Err:     err,
			})
		},
	})
	defer sub.Cancel()

	// Trigger native request
	_, err = l.alwaysChannel.Invoke("request", map[string]any{"permission": "location_always"})
	if err != nil {
		return PermissionResultUnknown, err
	}

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		// Re-check status in case we missed the event
		if finalStatus, err := l.StatusAlways(); err == nil && isTerminalStatus(finalStatus) {
			return finalStatus, nil
		}
		if ctx.Err() == context.DeadlineExceeded {
			return PermissionResultUnknown, ErrTimeout
		}
		return PermissionResultUnknown, ErrCanceled
	}
}

// StatusAlways returns the current status of the background location permission.
func (l *locationPermissionType) StatusAlways() (PermissionResult, error) {
	result, err := l.alwaysChannel.Invoke("check", map[string]any{
		"permission": "location_always",
	})
	if err != nil {
		return PermissionResultUnknown, err
	}
	return parsePermissionResult(result), nil
}

// IsAlwaysGranted returns true if background location permission is granted.
func (l *locationPermissionType) IsAlwaysGranted() bool {
	status, err := l.StatusAlways()
	if err != nil {
		return false
	}
	return status == PermissionGranted
}

// ChangesAlways returns a channel that receives background location permission status changes.
// The same channel is returned on each call. Only one goroutine should read from it;
// if multiple goroutines read, events will be split between them (not broadcast).
// The channel has a small buffer; slow consumers may miss events.
//
// Note: Use Changes() for when-in-use location updates and ChangesAlways() for background
// location updates. The platform emits separate events for each permission level.
func (l *locationPermissionType) ChangesAlways() <-chan PermissionResult {
	l.alwaysChangesOnce.Do(func() {
		l.alwaysChangesChan = make(chan PermissionResult, 4)
		l.changes.Listen(EventHandler{
			OnEvent: func(data any) {
				change, ok := parsePermissionChange(data)
				if !ok {
					return
				}
				// Only send changes for location_always
				if change.Permission == "location_always" {
					select {
					case l.alwaysChangesChan <- change.Result:
					default:
						// Channel full, skip
					}
				}
			},
			OnError: func(err error) {
				errors.Report(&errors.DriftError{
					Op:      "permissions.streamError",
					Kind:    errors.KindPlatform,
					Channel: "drift/permissions/changes",
					Err:     err,
				})
			},
		})
	})

	return l.alwaysChangesChan
}

// notificationPermissionType extends permissionType with notification-specific options.
type notificationPermissionType struct {
	*permissionType
}

func newNotificationPermission() *notificationPermissionType {
	return &notificationPermissionType{
		permissionType: newPermission("notifications"),
	}
}

// Request requests notification permission with optional configuration and blocks
// until the user responds or the default timeout is exceeded.
//
// If no options are provided, defaults to Alert, Sound, and Badge enabled.
// Options only affect iOS; on Android they are ignored.
//
// If the permission is already in a terminal state (granted, permanently_denied,
// restricted, or provisional), returns immediately without showing a dialog.
//
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the user did not respond within the timeout period
//   - Other errors: platform communication failures
func (n *notificationPermissionType) Request(opts ...NotificationOptions) (PermissionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPermissionTimeout)
	defer cancel()
	return n.RequestWithContext(ctx, opts...)
}

// RequestWithContext requests notification permission with optional configuration and blocks
// until the user responds, the context is canceled, or the deadline is exceeded.
//
// If no options are provided, defaults to Alert, Sound, and Badge enabled.
// Options only affect iOS; on Android they are ignored.
//
// If the permission is already in a terminal state, returns immediately.
// This method blocks and should be called from a goroutine, not the main/render thread.
//
// Errors:
//   - ErrTimeout: the context deadline was exceeded
//   - ErrCanceled: the context was canceled
//   - Other errors: platform communication failures
func (n *notificationPermissionType) RequestWithContext(ctx context.Context, opts ...NotificationOptions) (PermissionResult, error) {
	n.requestMu.Lock()
	defer n.requestMu.Unlock()

	options := NotificationOptions{Alert: true, Sound: true, Badge: true}
	if len(opts) > 0 {
		options = opts[0]
	}

	// Return immediately if already in terminal state
	currentStatus, err := n.Status()
	if err != nil {
		return PermissionResultUnknown, err
	}
	if isTerminalStatus(currentStatus) {
		return currentStatus, nil
	}

	// Subscribe BEFORE triggering native request to avoid race condition
	resultChan := make(chan PermissionResult, 1)
	sub := n.changes.Listen(EventHandler{
		OnEvent: func(data any) {
			change, ok := parsePermissionChange(data)
			if ok && change.Permission == n.name {
				select {
				case resultChan <- change.Result:
				default:
				}
			}
		},
		OnError: func(err error) {
			errors.Report(&errors.DriftError{
				Op:      "permissions.requestNotification",
				Kind:    errors.KindPlatform,
				Channel: "drift/permissions/changes",
				Err:     err,
			})
		},
	})
	defer sub.Cancel()

	// Trigger native request
	_, err = n.channel.Invoke("request", map[string]any{
		"permission":  n.name,
		"alert":       options.Alert,
		"sound":       options.Sound,
		"badge":       options.Badge,
		"provisional": options.Provisional,
	})
	if err != nil {
		return PermissionResultUnknown, err
	}

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		// Re-check status in case we missed the event
		if finalStatus, err := n.Status(); err == nil && isTerminalStatus(finalStatus) {
			return finalStatus, nil
		}
		if ctx.Err() == context.DeadlineExceeded {
			return PermissionResultUnknown, ErrTimeout
		}
		return PermissionResultUnknown, ErrCanceled
	}
}

// OpenAppSettings opens the system settings page for this app, where users can
// manage permissions manually. Use this when a permission is permanently denied
// and the app cannot request it again.
//
// On iOS, opens the Settings app to the app's settings page.
// On Android, opens the App Info screen in system settings.
func OpenAppSettings() error {
	channel := NewMethodChannel("drift/permissions")
	_, err := channel.Invoke("openSettings", nil)
	return err
}

// permissionChange represents a permission status change event (internal use).
type permissionChange struct {
	Permission string
	Result     PermissionResult
}

func parsePermissionResult(result any) PermissionResult {
	if m, ok := result.(map[string]any); ok {
		if status := parseString(m["status"]); status != "" {
			return PermissionResult(status)
		}
	}
	return PermissionResultUnknown
}

func parsePermissionChange(data any) (permissionChange, bool) {
	m, ok := data.(map[string]any)
	if !ok {
		return permissionChange{}, false
	}
	return permissionChange{
		Permission: parseString(m["permission"]),
		Result:     PermissionResult(parseString(m["status"])),
	}, true
}
