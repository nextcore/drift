---
id: intro
title: Introduction
sidebar_position: 1
---

# Drift

Drift is a cross-platform mobile UI framework in Go. It lets you write UI and application logic in Go, then build native Android and iOS apps via the Drift CLI, which generates platform scaffolding in a build cache and compiles your app with CGO + Skia.

## Why Drift?

- **Single codebase** - Write your app once in Go, deploy to Android and iOS
- **Go-native** - Use Go's tooling, testing, and ecosystem you already know
- **Skia rendering** - Hardware-accelerated graphics via the same engine Chrome and Flutter use
- **No bridge overhead** - Direct native compilation, no JavaScript or VM layer
- **iOS builds on Linux** - Build iOS apps without a Mac using [xtool](https://xtool.sh)

<video src="/showcase.mp4"  controls autoPlay muted playsInline width="300" />

## Prerequisites

- Go 1.24
- Android builds: Android SDK + NDK, Java 17+, and `ANDROID_HOME` + `ANDROID_NDK_HOME` env vars
- iOS builds: macOS with Xcode installed
- Skia: prebuilt binaries (downloaded automatically on first run)

## Quick Start

### 1. Install Drift

```bash
go install github.com/go-drift/drift/cmd/drift@latest
```

Make sure `$(go env GOPATH)/bin` or `GOBIN` is on your `PATH` so the `drift` command is available.

### 2. Create a New Project

**Option A: Using drift init (recommended)**

```bash
drift init hello-drift
cd hello-drift
```

**Option B: Manual setup**

```bash
mkdir hello-drift && cd hello-drift
go mod init example.com/hello-drift
go get github.com/go-drift/drift@latest
```

Then create `main.go`:

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

### 3. Run Your App

```bash
drift run android
# or
drift run ios --simulator "iPhone 17"
# or
drift run xtool
```

Skia binaries are downloaded automatically on first run.

### 4. (Optional) Add Configuration

Create `drift.yaml` to customize app metadata:

```yaml
app:
  name: Hello Drift
  id: com.example.hellodrift
engine:
  version: latest
```

## Build Commands

```bash
# Build
drift build android
drift build ios
drift build xtool

# Run on devices/simulators
drift run android
drift run ios --simulator "iPhone 17"
drift run xtool

# Fetch Skia binaries
drift fetch-skia              # all platforms
drift fetch-skia --android    # android only
drift fetch-skia --ios        # ios only

# Clean build cache
drift clean
```

## Next Steps

- [Getting Started Guide](/docs/guides/getting-started) - Detailed setup instructions
- [Widgets](/docs/guides/widgets) - Learn about available widgets
- [State Management](/docs/guides/state-management) - Managing state in your app
- [API Reference](/docs/api/core) - Full API documentation
