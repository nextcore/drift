---
id: widgets
title: Widgets
sidebar_position: 2
---

# Widgets

Widgets are the building blocks of Drift applications. Everything you see on screen is a widget.

## Widget Types

Drift has three types of widgets:

1. **StatelessWidget** - For UI that depends only on configuration
2. **StatefulWidget** - For UI that manages mutable state
3. **InheritedWidget** - For sharing data down the widget tree

## Creating Widgets

### Struct Literals

Use struct literals when fields are self-documenting:

```go
button := widgets.Button{
    Label: "Submit",
    OnTap: handleSubmit,
}

text := widgets.Text{
    Content: "Hello, Drift",
    Style:   rendering.TextStyle{Color: colors.OnSurface, FontSize: 16},
}
```

### Helper Functions

Use helpers when sensible defaults improve ergonomics:

```go
// NewButton applies defaults (haptic feedback, etc.)
button := widgets.NewButton("Submit", handleSubmit)

// TextOf is concise for styled text
title := widgets.TextOf("Welcome", textTheme.HeadlineLarge)

// Centered wraps a child in a Center widget
centered := widgets.Centered(child)

// ColumnOf/RowOf avoid verbose struct initialization
col := widgets.ColumnOf(
    widgets.MainAxisAlignmentStart,
    widgets.CrossAxisAlignmentStart,
    widgets.MainAxisSizeMin,
    child1,
    child2,
)
```

### Builder Pattern

Chain methods when you need to override defaults:

```go
button := widgets.NewButton("Submit", onSubmit).
    WithColor(colors.Primary, colors.OnPrimary).
    WithFontSize(18).
    WithPadding(layout.EdgeInsetsSymmetric(32, 16))

container := widgets.NewContainer(child).
    WithColor(colors.Surface).
    WithPaddingAll(20).
    WithAlignment(layout.AlignmentCenter).
    Build()
```

## Available Widgets

### Layout Widgets

| Widget | Purpose |
|--------|---------|
| `Row` | Horizontal arrangement |
| `Column` | Vertical arrangement |
| `Stack` | Overlay children |
| `IndexedStack` | Show one child at a time by index |
| `Center` | Center child in available space |
| `Padding` | Add spacing around child |
| `Container` | Decoration, sizing, alignment |
| `SizedBox` | Fixed dimensions |
| `Expanded` | Fill remaining flex space |
| `SafeArea` | Avoid system UI (notches, nav bars) |
| `Positioned` | Absolute positioning within Stack |

### Display Widgets

| Widget | Purpose |
|--------|---------|
| `Text` | Display text with styling |
| `Icon` | Material icons |
| `SVGIcon` | SVG vector icons |
| `Image` | Display images from assets/files |

### Input Widgets

| Widget | Purpose |
|--------|---------|
| `Button` | Tappable button with haptic feedback |
| `NativeTextField` | Platform-native text input |
| `Checkbox` | Boolean toggle |
| `Radio` | Single selection from group |
| `Switch` | On/off toggle |
| `Dropdown` | Selection menu |
| `Form` | Form container with validation |

### Scrolling Widgets

| Widget | Purpose |
|--------|---------|
| `ScrollView` | Scrollable content |
| `ListView` | Scrollable list of widgets |
| `ListViewBuilder` | Virtualized lazy-loading list |

### Decorative Widgets

| Widget | Purpose |
|--------|---------|
| `ClipRRect` | Rounded rectangle clipping |
| `DecoratedBox` | Background, border, gradient, shadow |
| `Opacity` | Static transparency |
| `AnimatedOpacity` | Animated transparency |
| `AnimatedContainer` | Animated layout/decoration changes |
| `RepaintBoundary` | Isolate paint for performance |

### Navigation Widgets

| Widget | Purpose |
|--------|---------|
| `Navigator` | Stack-based route management |
| `TabBar` | Tab navigation bar |
| `TabScaffold` | Tab layout with content |

### Error Handling

| Widget | Purpose |
|--------|---------|
| `ErrorBoundary` | Catch and recover from build errors |

## Custom Stateless Widgets

Create a stateless widget for UI that depends only on its configuration:

```go
type Greeting struct {
    Name string
}

func (g Greeting) CreateElement() core.Element {
    return core.NewStatelessElement(g, nil)
}

func (g Greeting) Key() any { return nil }

func (g Greeting) Build(ctx core.BuildContext) core.Widget {
    return widgets.TextOf("Hello, "+g.Name, theme.TextThemeOf(ctx).BodyLarge)
}
```

## Custom Stateful Widgets

Create a stateful widget when you need to manage mutable state:

```go
type Counter struct{}

func (c Counter) CreateElement() core.Element {
    return core.NewStatefulElement(c, nil)
}

func (c Counter) Key() any { return nil }

func (c Counter) CreateState() core.State {
    return &counterState{}
}

type counterState struct {
    core.StateBase
    count int
}

func (s *counterState) InitState() {
    s.count = 0
}

func (s *counterState) Build(ctx core.BuildContext) core.Widget {
    return widgets.NewButton(
        fmt.Sprintf("Count: %d", s.count),
        func() {
            s.SetState(func() {
                s.count++
            })
        },
    )
}
```

## Error Boundaries

Catch and recover from errors in a widget subtree:

```go
widgets.ErrorBoundary{
    ChildWidget: riskyWidget,
    FallbackBuilder: func(err *errors.BuildError) core.Widget {
        return widgets.Text{Content: "Something went wrong"}
    },
    OnError: func(err *errors.BuildError) {
        log.Printf("Widget error: %v", err)
    },
}
```

Error boundaries prevent a single widget's error from crashing your entire app. They catch panics during the build phase and display a fallback UI instead.

### When to Use Error Boundaries

- Around third-party widgets
- Around complex widget subtrees
- At screen boundaries
- Around widgets that depend on external data

```go
// Wrap each screen in an error boundary
func buildHomeScreen(ctx core.BuildContext) core.Widget {
    return widgets.ErrorBoundary{
        ChildWidget: HomeContent{},
        FallbackBuilder: func(err *errors.BuildError) core.Widget {
            return widgets.Column{
                ChildrenWidgets: []core.Widget{
                    widgets.Text{Content: "Failed to load home screen"},
                    widgets.NewButton("Retry", func() {
                        // Trigger rebuild
                    }),
                },
            }
        },
    }
}
```

## Lists and Scrolling

### Basic ListView

For small lists with all items in memory:

```go
widgets.ListView{
    ChildrenWidgets: []core.Widget{
        item1,
        item2,
        item3,
    },
    Padding: layout.EdgeInsetsAll(16),
}
```

### Virtualized Lists

For large lists, use `ListViewBuilder` which only builds visible items:

```go
widgets.ListViewBuilder{
    ItemCount:   1000,
    ItemExtent:  60,  // Fixed item height enables virtualization
    CacheExtent: 100, // Extra pixels to render beyond viewport
    ItemBuilder: func(ctx core.BuildContext, index int) core.Widget {
        item := items[index]
        return widgets.NewContainer(
            widgets.Text{Content: item.Title},
        ).WithPaddingAll(16).Build()
    },
}
```

### Scroll Physics

Control scroll behavior:

```go
widgets.ScrollView{
    Physics: widgets.BouncingScrollPhysics{}, // iOS-style bounce
    // or
    Physics: widgets.ClampingScrollPhysics{}, // Android-style clamp
    ChildWidget: content,
}
```

### Scroll Direction

```go
// Helper to get pointer to Axis
horizontal := widgets.AxisHorizontal

widgets.ListView{
    ScrollDirection: &horizontal, // Pointer to Axis (nil defaults to vertical)
    ChildrenWidgets: items,
}
```

## Next Steps

- [State Management](/docs/guides/state-management) - Managing widget state
- [Layout](/docs/guides/layout) - Arranging widgets
- [Animation](/docs/guides/animation) - Animating widgets
- [API Reference](/docs/api/widgets) - Full widgets API
