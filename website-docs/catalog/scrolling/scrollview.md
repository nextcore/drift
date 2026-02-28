---
id: scrollview
title: ScrollView
---

# ScrollView

A scrollable container for content that isn't a list. Wraps a single child and provides scroll behavior.

## Basic Usage

```go
widgets.ScrollView{
    Child: widgets.Column{
        Children: []core.Widget{
            header,
            content,
            footer,
        },
    },
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `Child` | `core.Widget` | Scrollable content |
| `ScrollDirection` | `Axis` | Scroll axis (default vertical) |
| `Controller` | `*ScrollController` | Optional scroll controller |
| `Physics` | `ScrollPhysics` | Scroll behavior (bounce or clamp) |
| `Padding` | `layout.EdgeInsets` | Padding around scrollable content |

## Scroll Physics

Control scroll behavior when reaching the edges:

```go
widgets.ScrollView{
    Physics: widgets.BouncingScrollPhysics{}, // iOS-style bounce
    Child:   content,
}

widgets.ScrollView{
    Physics: widgets.ClampingScrollPhysics{}, // Android-style clamp
    Child:   content,
}
```

| Physics | Behavior |
|---------|----------|
| `BouncingScrollPhysics` | Bounces when reaching edges (iOS style) |
| `ClampingScrollPhysics` | Clamps at edges (Android style, default) |

### Custom Scroll Physics

Implement the `ScrollPhysics` interface to define custom scroll behavior. Return `true` from `AllowsOverscroll()` to enable overscroll with spring-back animation:

```go
type MyCustomPhysics struct{}

func (MyCustomPhysics) ApplyPhysicsToUserOffset(pos *widgets.ScrollPosition, offset float64) float64 {
    return offset
}

func (MyCustomPhysics) ApplyBoundaryConditions(pos *widgets.ScrollPosition, value float64) float64 {
    return 0
}

func (MyCustomPhysics) AllowsOverscroll() bool { return true }
```

## Related

- [ListView](/docs/catalog/scrolling/listview) for scrollable lists of items
- [SafeArea](/docs/catalog/layout/safearea) for avoiding system UI in scrollable content
