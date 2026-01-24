---
id: platform
title: Platform Services
sidebar_position: 10
---

# Platform Services

Drift provides access to native platform capabilities through the `platform` package.

## Clipboard

Copy and paste text:

```go
import "github.com/go-drift/drift/pkg/platform"

// Copy text to clipboard
err := platform.Clipboard.SetText("Hello, World!")

// Get text from clipboard
text, err := platform.Clipboard.GetText()

// Check if clipboard has text
hasText, err := platform.Clipboard.HasText()

// Clear clipboard
err := platform.Clipboard.Clear()
```

### Example: Copy Button

```go
widgets.NewButton("Copy Link", func() {
    platform.Clipboard.SetText(shareURL)
    platform.Haptics.LightImpact()
    showToast("Link copied!")
})
```

## Haptic Feedback

Provide tactile feedback to users:

```go
// Light feedback for selections
platform.Haptics.LightImpact()

// Medium feedback for toggles and confirmations
platform.Haptics.MediumImpact()

// Heavy feedback for errors or warnings
platform.Haptics.HeavyImpact()

// Selection change feedback
platform.Haptics.SelectionClick()

// Custom vibration duration (milliseconds)
platform.Haptics.Vibrate(100)
```

### When to Use Haptics

| Feedback Type | Use Case |
|--------------|----------|
| `LightImpact` | List item selection, minor interactions |
| `MediumImpact` | Toggle switches, button taps |
| `HeavyImpact` | Errors, deletions, significant actions |
| `SelectionClick` | Picker value changes, slider movements |

## App Lifecycle

Respond to app lifecycle state changes:

```go
func (s *myState) InitState() {
    removeHandler := platform.Lifecycle.AddHandler(func(state platform.LifecycleState) {
        switch state {
        case platform.LifecycleStateResumed:
            // App came to foreground
            s.refreshData()
        case platform.LifecycleStatePaused:
            // App went to background
            s.saveState()
        case platform.LifecycleStateInactive:
            // App is transitioning (e.g., receiving a phone call)
        case platform.LifecycleStateDetached:
            // App is detached from any view
        }
    })

    // Clean up when widget is disposed
    s.OnDispose(removeHandler)
}
```

### Lifecycle States

| State | Description |
|-------|-------------|
| `LifecycleStateResumed` | App is visible and responding to user input |
| `LifecycleStateInactive` | App is transitioning (system dialog, app switcher) |
| `LifecycleStatePaused` | App is not visible but still running |
| `LifecycleStateDetached` | App is detached from any view |

### Checking Current State

```go
// Get current state
currentState := platform.Lifecycle.State()

// Convenience methods
if platform.Lifecycle.IsResumed() {
    // App is active
}

if platform.Lifecycle.IsPaused() {
    // App is in background
}
```

## System UI

Customize the status bar and system chrome:

```go
// Set system UI style
platform.SetSystemUI(platform.SystemUIStyle{
    StatusBarHidden: false,
    StatusBarStyle:  platform.StatusBarStyleLight, // or StatusBarStyleDark
    TitleBarHidden:  false,        // Android only
    Transparent:     false,        // Android only
    BackgroundColor: &colors.Surface, // Android only
})
```

## Permissions

Request runtime permissions:

```go
// Request a single permission
result, err := platform.RequestPermission(platform.PermissionCamera)
if result == platform.PermissionGranted {
    openCamera()
}

// Check current permission status
status, err := platform.CheckPermission(platform.PermissionLocation)

// Request multiple permissions
results, err := platform.RequestPermissions([]platform.Permission{
    platform.PermissionCamera,
    platform.PermissionMicrophone,
})

// Open app settings for manual permission management
platform.OpenAppSettings()
```

### Available Permissions

| Permission | Use |
|------------|-----|
| `PermissionCamera` | Camera access |
| `PermissionMicrophone` | Microphone access |
| `PermissionLocation` | Location services |
| `PermissionLocationAlways` | Background location |
| `PermissionStorage` | File storage |
| `PermissionContacts` | Contacts access |
| `PermissionPhotos` | Photo library |
| `PermissionCalendar` | Calendar access |
| `PermissionNotifications` | Push notifications |

### Permission Results

| Result | Meaning |
|--------|---------|
| `PermissionGranted` | Permission was granted |
| `PermissionDenied` | Permission was denied |
| `PermissionPermanentlyDenied` | User selected "Don't ask again" |
| `PermissionRestricted` | Restricted by device policy |
| `PermissionLimited` | Limited access granted (iOS photos) |

## Notifications

Schedule local notifications:

```go
// Schedule a notification
err := platform.ScheduleLocalNotification(platform.NotificationRequest{
    ID:    "reminder-1",
    Title: "Reminder",
    Body:  "Meeting in 5 minutes",
    At:    time.Now().Add(5 * time.Minute),
    Data:  map[string]any{"meetingId": "123"},
})

// Cancel a notification
platform.CancelLocalNotification("reminder-1")

// Cancel all notifications
platform.CancelAllLocalNotifications()

// Set app badge count
platform.SetNotificationBadge(3)
```

### Listening for Notifications

```go
// Listen for delivered notifications
go func() {
    for event := range platform.Notifications() {
        drift.Dispatch(func() {
            handleNotification(event)
        })
    }
}()

// Listen for notification opens (user tapped notification)
go func() {
    for open := range platform.NotificationOpens() {
        drift.Dispatch(func() {
            navigateToContent(open.Data)
        })
    }
}()
```

## Share

Open the native share sheet:

```go
// Share text
result, err := platform.Share.ShareText("Check out this app!")

// Share text with subject (for email)
result, err := platform.Share.ShareTextWithSubject("Check out this!", "App Recommendation")

// Share a URL
result, err := platform.Share.ShareURL("https://example.com")

// Share URL with text
result, err := platform.Share.ShareURLWithText("https://example.com", "Check out this link!")

// Share a file
result, err := platform.Share.ShareFile("/path/to/file.pdf", "application/pdf")
```

## Location

Access device location:

```go
// Get current location
location, err := platform.GetCurrentLocation(platform.LocationOptions{
    HighAccuracy: true,
})
if err == nil {
    fmt.Printf("Lat: %f, Lng: %f\n", location.Latitude, location.Longitude)
}

// Start continuous location updates
platform.StartLocationUpdates(platform.LocationOptions{
    HighAccuracy:   true,
    DistanceFilter: 10, // meters
})

// Listen for updates
go func() {
    for update := range platform.LocationUpdates() {
        drift.Dispatch(func() {
            s.SetState(func() {
                s.userLocation = update
            })
        })
    }
}()

// Stop updates when done
platform.StopLocationUpdates()
```

### Location Data

| Field | Type | Description |
|-------|------|-------------|
| `Latitude` | `float64` | Latitude in degrees |
| `Longitude` | `float64` | Longitude in degrees |
| `Altitude` | `float64` | Altitude in meters |
| `Accuracy` | `float64` | Accuracy in meters |
| `Heading` | `float64` | Direction in degrees |
| `Speed` | `float64` | Speed in m/s |
| `Timestamp` | `time.Time` | When reading was taken |

## File Storage

Access files and directories:

```go
// Read a file
data, err := platform.ReadFile("/path/to/file.txt")

// Write a file
err := platform.WriteFile("/path/to/file.txt", []byte("content"))

// Delete a file
err := platform.DeleteFile("/path/to/file.txt")

// Get file info
info, err := platform.GetFileInfo("/path/to/file.txt")

// Get app directory
docsPath, err := platform.GetAppDirectory(platform.AppDirectoryDocuments)
cachePath, err := platform.GetAppDirectory(platform.AppDirectoryCache)
```

### File Picker

```go
// Open file picker
platform.PickFile(platform.PickFileOptions{
    AllowMultiple: false,
    AllowedTypes:  []string{"image/*", "application/pdf"},
})

// Listen for results
go func() {
    for result := range platform.StorageResults() {
        if result.Cancelled {
            continue
        }
        for _, file := range result.Files {
            handleSelectedFile(file)
        }
    }
}()
```

## Thread Safety

Platform services are safe to call from any goroutine. However, when updating UI state from platform callbacks, use `drift.Dispatch`:

```go
go func() {
    for update := range platform.LocationUpdates() {
        // Called from background goroutine
        drift.Dispatch(func() {
            // Now safe to update UI
            s.SetState(func() {
                s.location = update
            })
        })
    }
}()
```

## Next Steps

- [Accessibility](/docs/guides/accessibility) - Make your app accessible
- [API Reference](/docs/api/platform) - Platform API documentation
