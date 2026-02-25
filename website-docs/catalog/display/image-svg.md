---
id: image-svg
title: Image & SVG
---

# Image & SVG

Display raster images from assets or files, and render SVG content with flexible sizing.

## Image

Display a decoded image:

```go
widgets.Image{
    Source: myImage,  // image.Image
    Width:  200,
    Height: 150,
}
```

### Image Properties

| Property | Type | Description |
|----------|------|-------------|
| `Source` | `image.Image` | Decoded image to render |
| `Width` | `float64` | Display width |
| `Height` | `float64` | Display height |

## SvgImage

Renders an SVG with flexible sizing:

```go
widgets.SvgImage{
    Source:    myIcon,
    Width:     120,
    Height:    80,
    TintColor: colors.Primary,  // Optional tint
}
```

### SvgImage Properties

| Property | Type | Description |
|----------|------|-------------|
| `Source` | `*svg.Icon` | Loaded SVG icon |
| `Width` | `float64` | Display width |
| `Height` | `float64` | Display height |
| `TintColor` | `graphics.Color` | Optional tint color |

## SvgIcon

A convenience wrapper around `SvgImage` for square icons:

```go
widgets.SvgIcon{
    Source:    myIcon,
    Size:      24,
    TintColor: colors.OnSurface,
}
```

### SvgIcon Properties

| Property | Type | Description |
|----------|------|-------------|
| `Source` | `*svg.Icon` | Loaded SVG icon |
| `Size` | `float64` | Width and height (square) |
| `TintColor` | `graphics.Color` | Optional tint color |

## Caching Static SVGs

For static SVG assets (logos, icons), cache loaded icons so rebuilds reuse the same underlying SVG DOM:

```go
var svgCache = svg.NewIconCache()

func loadIcon(name string) *svg.Icon {
    icon, err := svgCache.Get(name, func() (*svg.Icon, error) {
        f, err := assetFS.Open("assets/" + name)
        if err != nil {
            return nil, err
        }
        defer f.Close()
        return svg.Load(f)
    })
    if err != nil {
        return nil
    }
    return icon
}
```

## Related

- [Icon](/docs/catalog/display/icon) for rendering text glyphs as icons
- [Widget Architecture](/docs/guides/widgets) for how widgets work
