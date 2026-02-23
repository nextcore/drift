---
id: icon
title: Icon
---

# Icon

Renders a single text glyph (such as a Unicode symbol or emoji) with icon-friendly defaults.

Drift does not bundle an icon font. For icons, use [SvgIcon](/docs/catalog/display/image-svg#svgicon) to render SVG files from your own assets, or use [Image](/docs/catalog/display/image-svg#image) for PNG/raster icons.

## Basic Usage

```go
widgets.Icon{
    Glyph: "★",
    Size:  24,
    Color: colors.Primary,
}
```

For theme-styled icons with sensible defaults (size 24, `OnSurface` color):

```go
theme.IconOf(ctx, "✓")
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `Glyph` | `string` | The text glyph to render |
| `Size` | `float64` | Font size in pixels (zero means not rendered) |
| `Color` | `graphics.Color` | Glyph color (zero means transparent) |
| `Weight` | `graphics.FontWeight` | Font weight (optional) |

## Common Patterns

### Glyph Icon in a Row

```go
widgets.Row{
    CrossAxisAlignment: widgets.CrossAxisAlignmentCenter,
    MainAxisSize:       widgets.MainAxisSizeMin,
    Children: []core.Widget{
        widgets.Icon{Glyph: "✉", Size: 20, Color: colors.OnSurfaceVariant},
        widgets.HSpace(8),
        widgets.Text{Content: "user@example.com", Style: textTheme.BodyMedium},
    },
}
```

### SVG Icon from an Asset

For real icon assets, use `SvgIcon` with SVG files bundled in your app:

```go
var svgCache = svg.NewIconCache()

func loadIcon(name string) *svg.Icon {
    icon, _ := svgCache.Get(name, func() (*svg.Icon, error) {
        f, err := assetFS.Open("assets/" + name)
        if err != nil {
            return nil, err
        }
        defer f.Close()
        return svg.Load(f)
    })
    return icon
}

// Then use it in your widget tree:
widgets.SvgIcon{
    Source: loadIcon("settings.svg"),
    Size:   24,
    TintColor: colors.Primary,
}
```

## Related

- [Image & SVG](/docs/catalog/display/image-svg) for SVG and raster image display
- [Button](/docs/catalog/input/button) for tappable icon buttons
