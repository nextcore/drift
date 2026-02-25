---
id: safearea
title: SafeArea
---

# SafeArea

Avoids system UI such as notches, status bars, and navigation bars by adding padding to keep content within the safe region.

## Basic Usage

```go
widgets.SafeArea{
    Child: content,
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `Top` | `bool` | Include top inset |
| `Bottom` | `bool` | Include bottom inset |
| `Left` | `bool` | Include left inset |
| `Right` | `bool` | Include right inset |
| `Child` | `core.Widget` | Child widget |

By default (all fields `false`), SafeArea applies padding on **all** sides. Setting one or more sides to `true` switches to selective mode, where only the specified sides receive padding. For example, setting `Top: true, Bottom: true` applies padding on top and bottom only, leaving left and right unpadded.

## Selective Sides

Apply safe area padding only on specific sides:

```go
widgets.SafeArea{
    Top:    true,
    Bottom: true,
    Child:  content,
}
```

## Common Patterns

### Full-Screen Layout with Safe Content

```go
func (s *myState) Build(ctx core.BuildContext) core.Widget {
    return widgets.SafeArea{
        Child: widgets.PaddingAll(20,
            widgets.Column{
                MainAxisSize: widgets.MainAxisSizeMin,
                Children: []core.Widget{
                    header,
                    widgets.VSpace(16),
                    content,
                },
            },
        ),
    }
}
```

## Related

- [Padding](/docs/catalog/layout/padding) for adding custom spacing
- [Layout System](/docs/guides/layout) for how constraints flow through the tree
