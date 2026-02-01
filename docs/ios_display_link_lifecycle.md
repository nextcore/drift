# iOS Display Link Lifecycle Bug

## Issue

When a modal view controller (camera picker, share sheet, date picker, etc.) is presented over the DriftViewController, the render loop stops and never restarts.

## Root Cause

`DriftViewController` manages a `CADisplayLink` for the render loop:

- `viewDidLoad` → starts display link
- `viewDidDisappear` → stops display link
- `viewDidAppear` → **did NOT restart display link** (bug)

When a modal is presented:
1. `viewDidDisappear` is called → display link stops
2. User interacts with modal
3. Modal is dismissed
4. `viewDidAppear` is called → display link was NOT restarted
5. Render loop never runs again → dispatched callbacks never execute

## Symptoms

- UI updates queued via `drift.Dispatch()` never execute
- App appears frozen after dismissing any modal
- Platform channel events (camera results, etc.) are received but UI doesn't update

## Fix

Added `viewWillAppear` to restart the display link if it was stopped:

```swift
override func viewWillAppear(_ animated: Bool) {
    super.viewWillAppear(animated)
    // Restart the render loop when view becomes visible again
    // (e.g., after dismissing a modal like camera picker)
    if displayLink == nil {
        startDisplayLink()
    }
}
```

## Files Changed

- `cmd/drift/internal/templates/ios/DriftViewController.swift`

## Affected Features

Any feature that presents a modal:
- Camera capture (UIImagePickerController)
- Photo gallery picker (PHPickerViewController)
- Share sheets (UIActivityViewController)
- Date/time pickers
- Any custom modal presentations
