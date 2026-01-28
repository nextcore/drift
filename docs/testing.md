# Widget Testing

Drift ships a testing framework in `pkg/testing` that lets you mount widgets, drive the build/layout/paint pipeline, simulate gestures, and compare render tree snapshots -- all without a real window or GPU.

Since the package name collides with the standard library, import it with an alias:

```go
import drifttest "github.com/go-drift/drift/pkg/testing"
```

## Setup

Create a `WidgetTester` and mount a widget:

```go
func TestGreeting(t *testing.T) {
    tester := drifttest.NewWidgetTesterWithT(t)
    tester.PumpWidget(MyGreeting{Name: "World"})

    result := tester.Find(drifttest.ByText("Hello, World!"))
    if !result.Exists() {
        t.Error("expected greeting text")
    }
}
```

`NewWidgetTesterWithT` registers a cleanup function via `t.Cleanup()` so global state (the animation clock) is restored automatically. If you need manual control, use `NewWidgetTester()` and call `Cleanup()` yourself.

### What the tester provides

| Aspect | Default |
|--------|---------|
| Surface size | 800 x 600 logical pixels |
| Device scale | 1.0 |
| Theme | Material light (deep copy, isolated per test) |
| Clock | `FakeClock` starting at 2024-01-01T00:00:00Z |

Override any of these before calling `PumpWidget`:

```go
tester.SetSize(graphics.Size{Width: 375, Height: 812})
tester.SetScale(2.0)
tester.SetTheme(myCustomTheme)
```

### What PumpWidget does

`PumpWidget` wraps the widget in a test scaffold (`DeviceScale` -> `AppTheme` -> your widget), mounts the element tree, and runs one full frame:

1. Drain the dispatch queue
2. Step ballistics and tickers
3. Flush build (rebuild dirty elements)
4. Flush layout (from root with tight constraints matching surface size)
5. Flush paint

Call `Pump()` to run additional frames after state changes. Call `PumpAndSettle(timeout)` to loop frames until the framework is idle (no pending builds, active tickers, or queued dispatches).

```go
tester.Tap(drifttest.ByText("+"))
tester.Pump() // process the rebuild triggered by setState
```

## Finders

Finders locate elements in the widget tree using depth-first pre-order traversal.

| Finder | Matches |
|--------|---------|
| `ByType[T]()` | Elements whose widget is exactly type `T` |
| `ByText("hello")` | `widgets.Text` with exact content |
| `ByTextContaining("hel")` | `widgets.Text` containing substring |
| `ByKey(myKey)` | Elements whose widget key equals `myKey` |
| `ByPredicate(fn)` | Elements satisfying a custom function |
| `Descendant(of, matching)` | Elements matching `matching` that are descendants of `of` |
| `Ancestor(of, matching)` | Elements matching `matching` that are ancestors of `of` |

### FinderResult

`tester.Find(finder)` returns a `FinderResult` with these accessors:

```go
result := tester.Find(drifttest.ByType[widgets.Text]())

result.Exists()          // bool
result.Count()           // int
result.First()           // core.Element (panics if empty)
result.FirstOrNil()      // core.Element or nil
result.At(2)             // core.Element at index
result.All()             // []core.Element
result.Widget()          // first match's widget
result.RenderObject()    // first match's render object
```

### Example: verifying a counter

```go
func TestCounter(t *testing.T) {
    tester := drifttest.NewWidgetTesterWithT(t)
    tester.PumpWidget(Counter{Initial: 0})

    // Find the text displaying the count
    text := tester.Find(drifttest.ByType[widgets.Text]())
    if text.Widget().(widgets.Text).Content != "0" {
        t.Error("expected initial count 0")
    }

    // Tap and verify increment
    tester.Tap(drifttest.ByType[widgets.GestureDetector]())
    tester.Pump()

    text = tester.Find(drifttest.ByType[widgets.Text]())
    if text.Widget().(widgets.Text).Content != "1" {
        t.Error("expected count 1 after tap")
    }
}
```

## Gesture Simulation

The tester routes synthetic pointer events through the render tree's hit testing, matching the production engine's dispatch path.

| Method | Behavior |
|--------|----------|
| `Tap(finder)` | Pointer down + up at center of first match |
| `TapAt(offset)` | Pointer down + up at logical position |
| `Drag(finder, delta)` | Down at center, move by delta, up |
| `DragFrom(start, delta)` | Down at start, move by delta, up |
| `Fling(finder, delta, velocity)` | Down, intermediate moves, up |

Low-level methods are available for multi-touch or custom sequences:

```go
tester.SendPointerDown(pos, pointerID)
tester.SendPointerMove(pos, pointerID)
tester.SendPointerUp(pos, pointerID)
tester.SendPointerCancel(pointerID)
```

Each pointer down performs a hit test, collects `PointerHandler` implementations from the results, and closes the gesture arena. Pointer up sweeps the arena, matching production behavior.

## Animation Testing

The tester injects a `FakeClock` into the animation package, replacing `time.Now()` for all tickers. This gives tests deterministic control over time.

```go
func TestFadeIn(t *testing.T) {
    tester := drifttest.NewWidgetTesterWithT(t)
    tester.PumpWidget(FadeIn{Duration: 300 * time.Millisecond})

    // Advance to midpoint
    tester.Clock().Advance(150 * time.Millisecond)
    tester.Pump()

    // Check intermediate state...

    // Complete the animation
    tester.Clock().Advance(150 * time.Millisecond)
    tester.PumpAndSettle(time.Second)
}
```

`PumpAndSettle` returns `drifttest.ErrSettleTimeout` if the framework doesn't reach idle within the timeout. The settle condition requires all of:

- No dirty elements (`BuildOwner.NeedsWork() == false`)
- No active tickers
- No active ballistics
- Empty dispatch queue

## Snapshot Testing

Snapshots serialize the render tree structure and display list operations to JSON. They catch unintended layout or paint regressions without pixel comparison.

### What snapshots capture

A snapshot contains two sections:

**Render tree** -- every render object with:
- Stable ID (`RenderFlex#0`, `renderLayoutBox#1`, ...)
- Type name
- Size `[width, height]` (rounded to 2 decimals)
- Offset `[x, y]` from parent
- Whitelisted properties (varies by render type)
- Children (recursive)

**Display operations** -- the paint commands recorded through the canvas:
- `save`, `restore`, `translate`
- `drawRect`, `drawRRect`, `drawCircle`, `drawPath`
- `clipRect`, `clipRRect`
- `saveLayer`, `drawPicture`

Floats are rounded to 2 decimal places. Colors are hex strings (`0xFFFF0000`). Map keys are sorted alphabetically. This ensures deterministic output across runs.

### Writing a snapshot test

```go
func TestLoginForm_Layout(t *testing.T) {
    tester := drifttest.NewWidgetTesterWithT(t)
    tester.SetSize(graphics.Size{Width: 375, Height: 667})
    tester.PumpWidget(LoginForm{})

    snapshot := tester.CaptureSnapshot()
    snapshot.MatchesFile(t, "testdata/login_form.snapshot.json")
}
```

### Creating and updating snapshots

On first run, the test fails because the golden file doesn't exist:

```
snapshot file missing: testdata/login_form.snapshot.json

To create: DRIFT_UPDATE_SNAPSHOTS=1 go test -run TestLoginForm_Layout
```

Create or update snapshots by setting the environment variable:

```bash
DRIFT_UPDATE_SNAPSHOTS=1 go test ./...
```

This writes the current snapshot to disk. Subsequent runs compare against the file and report a diff on mismatch:

```
snapshot mismatch: testdata/login_form.snapshot.json
--- expected
+++ actual
-  "size": [375.00, 48.00],
+  "size": [375.00, 56.00],

To update: DRIFT_UPDATE_SNAPSHOTS=1 go test -run TestLoginForm_Layout
```

### Snapshot file location

There is no enforced directory. Pass any path to `MatchesFile`. The convention is `testdata/` relative to the test file:

```
mypackage/
    login_form.go
    login_form_test.go
    testdata/
        login_form.snapshot.json
```

`UpdateFile` creates intermediate directories automatically.

### Programmatic comparison

Use `Diff` to compare two snapshots directly without golden files:

```go
before := tester.CaptureSnapshot()

// ... modify state ...
tester.Pump()

after := tester.CaptureSnapshot()
if diff := before.Diff(after); diff != "" {
    t.Errorf("unexpected change:\n%s", diff)
}
```

## Layout Constraints

The tester applies **tight constraints** matching the surface size to the root render object. This means the root widget is forced to exactly fill the surface, just like a real window.

If your widget under test requests a smaller size, the render object will still be constrained to the surface size. To test a specific size, either:

1. Set the surface size to match: `tester.SetSize(graphics.Size{Width: 100, Height: 50})`
2. Check the widget's properties instead of the render object's constrained size: `result.Widget().(MyWidget).Width`

## Quick Reference

```go
import drifttest "github.com/go-drift/drift/pkg/testing"

// Create
tester := drifttest.NewWidgetTesterWithT(t)

// Configure (before PumpWidget)
tester.SetSize(graphics.Size{Width: 375, Height: 812})
tester.SetScale(2.0)
tester.SetTheme(myTheme)

// Mount and drive frames
tester.PumpWidget(myWidget)
tester.Pump()
tester.PumpAndSettle(5 * time.Second)

// Find elements
tester.Find(drifttest.ByType[widgets.Text]())
tester.Find(drifttest.ByText("Submit"))
tester.Find(drifttest.ByTextContaining("Sub"))
tester.Find(drifttest.ByKey("submit-btn"))
tester.Find(drifttest.ByPredicate(func(e core.Element) bool { ... }))
tester.Find(drifttest.Descendant(parent, child))

// Simulate gestures
tester.Tap(finder)
tester.TapAt(graphics.Offset{X: 100, Y: 200})
tester.Drag(finder, graphics.Offset{X: 0, Y: -300})

// Control time
tester.Clock().Advance(100 * time.Millisecond)

// Snapshots
snap := tester.CaptureSnapshot()
snap.MatchesFile(t, "testdata/my_widget.snapshot.json")
snap.UpdateFile("testdata/my_widget.snapshot.json")
diff := snapA.Diff(snapB)

// Tree access
tester.RootElement()
tester.RootRenderObject()

// Dispatch (runs on next Pump)
tester.Dispatch(func() { ... })
```
