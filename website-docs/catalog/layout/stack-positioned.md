---
id: stack-positioned
title: Stack & Positioned
---

# Stack & Positioned

`Stack` overlays children on top of each other. `Positioned` places a child at absolute coordinates within a Stack.

## Basic Usage

```go
widgets.Stack{
    Alignment: layout.AlignmentCenter,
    Fit:       widgets.StackFitLoose,
    Children: []core.Widget{
        backgroundImage,
        gradientOverlay,
        widgets.Positioned(titleText).Left(16).Right(16).Bottom(16),
    },
}
```

## Stack Properties

| Property | Type | Description |
|----------|------|-------------|
| `Alignment` | `layout.Alignment` | Default alignment for non-positioned children |
| `Fit` | `StackFit` | How non-positioned children are sized |
| `Children` | `[]core.Widget` | Child widgets (can include `Positioned` children) |

## Positioned

Use `Positioned` within a Stack for absolute positioning:

```go
widgets.Stack{
    Children: []core.Widget{
        mainContent,
        // Badge in top-right corner
        widgets.Positioned(badge).Top(8).Right(8),
    },
}
```

### Positioned Methods

| Method | Description |
|--------|-------------|
| `.Left(v)` | Distance from the left edge |
| `.Top(v)` | Distance from the top edge |
| `.Right(v)` | Distance from the right edge |
| `.Bottom(v)` | Distance from the bottom edge |
| `.Width(v)` | Override child width |
| `.Height(v)` | Override child height |
| `.Size(w, h)` | Set both width and height |
| `.Align(a)` | Relative positioning via alignment coordinates |
| `.Fill(inset)` | Set all four edges to the same inset |
| `.At(left, top)` | Set left and top position |

## Positioned with Alignment

`Positioned` also supports relative positioning via `Align`. When set, the child is centered on the alignment point within the Stack bounds.

The `Alignment` type uses coordinates from -1 to 1, where (-1, -1) is top-left, (0, 0) is center, and (1, 1) is bottom-right. Use the named constants like `graphics.AlignCenter`, `graphics.AlignBottomRight`, etc.

When `Align` is set, `Left`/`Top`/`Right`/`Bottom` become pixel offsets from that centered position:
- `Left`/`Top` shift the child in the positive direction (right/down)
- `Right`/`Bottom` shift the child in the negative direction (left/up, i.e., inward from edges)

```go
widgets.Stack{
    Children: []core.Widget{
        background,
        // Centered dialog (no offsets needed)
        widgets.Positioned(dialog).Align(graphics.AlignCenter),
        // Floating action button: starts at bottom-right corner,
        // then shifts 16px left and 16px up (inward from corner)
        widgets.Positioned(fab).Align(graphics.AlignBottomRight).Right(16).Bottom(16),
    },
}
```

## Stack Fit

| Fit | Effect |
|-----|--------|
| `StackFitLoose` | Children can be smaller than stack |
| `StackFitExpand` | Non-positioned children expand to fill |

## Related

- [Center & Align](/docs/catalog/layout/center-align) for simpler alignment without overlapping
- [Layout System](/docs/guides/layout) for how constraints flow through the tree
