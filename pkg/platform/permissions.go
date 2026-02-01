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
// from showing a permission dialog.
func isTerminalStatus(status PermissionResult) bool {
	switch status {
	case PermissionGranted, PermissionPermanentlyDenied, PermissionRestricted,
		PermissionLimited, PermissionProvisional:
		return true
	default:
		return false
	}
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
func (p *permissionType) Request() (PermissionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPermissionTimeout)
	defer cancel()
	return p.RequestWithContext(ctx)
}

// RequestWithContext requests the permission from the user and blocks until the user
// responds, the context is canceled, or the context deadline is exceeded.
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

// locationPermissionType extends permissionType with location-specific methods.
type locationPermissionType struct {
	*permissionType
	alwaysChannel *MethodChannel

	// Mutex to serialize always permission requests
	alwaysRequestMu sync.Mutex
}

func newLocationPermission() *locationPermissionType {
	return &locationPermissionType{
		permissionType: newPermission("location"),
		alwaysChannel:  NewMethodChannel("drift/permissions"),
	}
}

// RequestAlwaysWithContext requests permission to access location in the background
// and blocks until the user responds, the context is canceled, or the deadline is exceeded.
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

// notificationPermissionType extends permissionType with notification-specific options.
type notificationPermissionType struct {
	*permissionType
}

func newNotificationPermission() *notificationPermissionType {
	return &notificationPermissionType{
		permissionType: newPermission("notifications"),
	}
}

// RequestWithContext requests notification permission with optional configuration and blocks
// until the user responds, the context is canceled, or the deadline is exceeded.
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
// The ctx parameter is currently unused and reserved for future cancellation support.
func OpenAppSettings(ctx context.Context) error {
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

// basicPermission implements Permission by wrapping permissionType.
type basicPermission struct {
	inner *permissionType
}

func (p *basicPermission) Status(ctx context.Context) (PermissionStatus, error) {
	return p.inner.Status()
}

func (p *basicPermission) Request(ctx context.Context) (PermissionStatus, error) {
	return p.inner.RequestWithContext(ctx)
}

func (p *basicPermission) IsGranted(ctx context.Context) bool {
	return p.inner.IsGranted()
}

func (p *basicPermission) IsDenied(ctx context.Context) bool {
	return p.inner.IsDenied()
}

func (p *basicPermission) ShouldShowRationale(ctx context.Context) (bool, error) {
	return p.inner.ShouldShowRationale(), nil
}

func (p *basicPermission) Listen(handler func(PermissionStatus)) (unsubscribe func()) {
	sub := p.inner.changes.Listen(EventHandler{
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
			if change.Permission == p.inner.name {
				handler(change.Result)
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
	return sub.Cancel
}

// locationAlwaysPermission implements Permission for background location.
// Preserves iOS behavior: when-in-use must be granted before requesting always.
type locationAlwaysPermission struct {
	inner *locationPermissionType
}

func (p *locationAlwaysPermission) Status(ctx context.Context) (PermissionStatus, error) {
	return p.inner.StatusAlways()
}

func (p *locationAlwaysPermission) Request(ctx context.Context) (PermissionStatus, error) {
	return p.inner.RequestAlwaysWithContext(ctx)
}

func (p *locationAlwaysPermission) IsGranted(ctx context.Context) bool {
	return p.inner.IsAlwaysGranted()
}

func (p *locationAlwaysPermission) IsDenied(ctx context.Context) bool {
	status, err := p.inner.StatusAlways()
	if err != nil {
		return false
	}
	return status == PermissionDenied || status == PermissionPermanentlyDenied
}

func (p *locationAlwaysPermission) ShouldShowRationale(ctx context.Context) (bool, error) {
	return p.inner.ShouldShowRationale(), nil
}

func (p *locationAlwaysPermission) Listen(handler func(PermissionStatus)) (unsubscribe func()) {
	sub := p.inner.changes.Listen(EventHandler{
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
			if change.Permission == "location_always" {
				handler(change.Result)
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
	return sub.Cancel
}

// notificationPermissionImpl implements NotificationPermission.
type notificationPermissionImpl struct {
	inner *notificationPermissionType
}

func (p *notificationPermissionImpl) Status(ctx context.Context) (PermissionStatus, error) {
	return p.inner.Status()
}

func (p *notificationPermissionImpl) Request(ctx context.Context) (PermissionStatus, error) {
	// Default options: Alert, Sound, Badge enabled
	return p.inner.RequestWithContext(ctx)
}

func (p *notificationPermissionImpl) RequestWithOptions(ctx context.Context, opts NotificationPermissionOptions) (PermissionStatus, error) {
	return p.inner.RequestWithContext(ctx, NotificationOptions{
		Alert:       opts.Alert,
		Sound:       opts.Sound,
		Badge:       opts.Badge,
		Provisional: opts.Provisional,
	})
}

func (p *notificationPermissionImpl) IsGranted(ctx context.Context) bool {
	return p.inner.IsGranted()
}

func (p *notificationPermissionImpl) IsDenied(ctx context.Context) bool {
	return p.inner.IsDenied()
}

func (p *notificationPermissionImpl) ShouldShowRationale(ctx context.Context) (bool, error) {
	return p.inner.ShouldShowRationale(), nil
}

func (p *notificationPermissionImpl) Listen(handler func(PermissionStatus)) (unsubscribe func()) {
	sub := p.inner.changes.Listen(EventHandler{
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
			if change.Permission == p.inner.name {
				handler(change.Result)
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
	return sub.Cancel
}
