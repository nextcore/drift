---
id: getting-started
title: Getting Started
sidebar_position: 1
---

# Getting Started

Get your first Drift app running in three steps.

## Prerequisites

- **Go 1.24** or later
- **Android**: Android SDK, NDK, Java 17+, and environment variables:
  ```bash
  export ANDROID_HOME=/path/to/android/sdk
  export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<version>
  ```
- **iOS**: macOS with Xcode (or Linux with [xtool](/docs/guides/xtool-setup))

## 1. Install the CLI

```bash
go install github.com/go-drift/drift/cmd/drift@latest
```

Make sure `$(go env GOPATH)/bin` or `GOBIN` is on your `PATH` so the `drift` command is available.

## 2. Create a Project

```bash
drift init hello-drift
cd hello-drift
```

This creates:
- `main.go` - Your app entry point
- `go.mod` - Go module file

The generated `main.go`:

```go
package main

import (
    "github.com/go-drift/drift/pkg/core"
    "github.com/go-drift/drift/pkg/drift"
    "github.com/go-drift/drift/pkg/widgets"
)

func main() {
    drift.NewApp(App()).Run()
}

func App() core.Widget {
    return widgets.Centered(
        widgets.Text{Content: "Hello, Drift!"},
    )
}
```

## 3. Run Your App

```bash
drift run android
# or
drift run ios --simulator "iPhone 17"
# or
drift run xtool
```

Skia binaries are downloaded automatically on first run.

## Configuration

Customize your app with `drift.yaml`:

```yaml
app:
  name: Hello Drift
  id: com.example.hellodrift

engine:
  version: latest
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `drift init <name>` | Create a new project |
| `drift run android` | Run on Android device/emulator |
| `drift run ios` | Run on iOS device |
| `drift run ios --simulator "iPhone 17"` | Run on iOS simulator |
| `drift run xtool` | Run iOS from Linux via xtool |
| `drift build android\|ios\|xtool` | Build without running |
| `drift clean` | Clear build cache |
| `drift fetch-skia` | Download Skia binaries manually |

## Next Steps

- [Widgets](/docs/guides/widgets) - Build UI with widgets
- [State Management](/docs/guides/state-management) - Handle app state
- [Layout](/docs/guides/layout) - Arrange widgets on screen
