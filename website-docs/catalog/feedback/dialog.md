---
id: dialog
title: Dialog
---

# Dialog

Modal dialogs that appear above the main content with a semi-transparent barrier.

## ShowAlertDialog (Recommended)

The simplest way to present a dialog. Handles text styling, button layout, and dismiss logic automatically.

```go
overlay.ShowAlertDialog(ctx, overlay.AlertDialogOptions{
    Title:        "Delete item?",
    Content:      "This action cannot be undone.",
    ConfirmLabel: "Delete",
    OnConfirm:    func() { deleteItem() },
    CancelLabel:  "Cancel",
    Destructive:  true,
})
```

### AlertDialogOptions

| Property | Type | Description |
|----------|------|-------------|
| `Title` | `string` | Heading text (HeadlineSmall style). Empty omits it |
| `Content` | `string` | Body text (BodyMedium style). Empty omits it |
| `ConfirmLabel` | `string` | Confirm button label. Empty omits the button |
| `OnConfirm` | `func()` | Called before dismiss when confirm is tapped |
| `CancelLabel` | `string` | Cancel button label. Empty omits the button |
| `OnCancel` | `func()` | Called before dismiss when cancel is tapped |
| `Destructive` | `bool` | Style confirm button with Error color |
| `Persistent` | `bool` | Prevent barrier tap from dismissing |

## ShowDialog (Custom Content)

For full control over dialog content, use `ShowDialog` with a builder function. The builder receives the `BuildContext` and a `dismiss` callback.

```go
dismiss := overlay.ShowDialog(ctx, overlay.DialogOptions{
    BarrierColor: graphics.RGBA(0, 0, 0, 0.5),
    Builder: func(ctx core.BuildContext, dismiss func()) core.Widget {
        textTheme := theme.ThemeOf(ctx).TextTheme
        return overlay.Dialog{
            Child: widgets.Column{
                MainAxisSize: widgets.MainAxisSizeMin,
                Children: []core.Widget{
                    theme.TextOf(ctx, "Custom Dialog", textTheme.HeadlineSmall),
                    widgets.VSpace(16),
                    theme.TextOf(ctx, "Any content goes here.", textTheme.BodyMedium),
                    widgets.VSpace(24),
                    theme.ButtonOf(ctx, "Close", dismiss),
                },
            },
        }
    },
})
```

### DialogOptions

| Property | Type | Description |
|----------|------|-------------|
| `Builder` | `DialogBuilder` | Creates the dialog content widget (required) |
| `Persistent` | `bool` | Prevent barrier tap from dismissing |
| `BarrierColor` | `graphics.Color` | Barrier color. Zero is transparent; set explicitly for a visible scrim |

The returned `dismiss` function removes the dialog programmatically. It is idempotent.

## Dialog Widget (Card Chrome)

`Dialog` provides themed card chrome: surface color, border radius, elevation shadow, and padding. It reads from `DialogThemeData`.

```go
overlay.Dialog{
    Child: myContent,
    Width: 320, // optional, zero = shrink to content
}
```

### Dialog Properties

| Property | Type | Description |
|----------|------|-------------|
| `Child` | `core.Widget` | Dialog content |
| `Width` | `float64` | Explicit width in pixels. Zero shrinks to content |

## AlertDialog Widget (Layout)

`AlertDialog` arranges title, content, and actions inside a `Dialog`. Use this when you need themed card chrome with structured layout but want to provide your own widgets.

```go
overlay.AlertDialog{
    Title:   theme.TextOf(ctx, "Confirm", textTheme.HeadlineSmall),
    Content: theme.TextOf(ctx, "Proceed?", textTheme.BodyMedium),
    Actions: []core.Widget{
        theme.ButtonOf(ctx, "Cancel", dismiss),
        theme.ButtonOf(ctx, "OK", func() { confirm(); dismiss() }),
    },
    Width: 320,
}
```

### AlertDialog Properties

| Property | Type | Description |
|----------|------|-------------|
| `Title` | `core.Widget` | Heading widget. Nil omits it |
| `Content` | `core.Widget` | Body widget. Nil omits it |
| `Actions` | `[]core.Widget` | Buttons placed in a right-aligned row |
| `Width` | `float64` | Dialog width. Zero defaults to 280 |

## Custom Container (No Dialog Widget)

Skip the `Dialog` widget entirely for completely custom chrome:

```go
overlay.ShowDialog(ctx, overlay.DialogOptions{
    BarrierColor: graphics.RGBA(0, 0, 0, 0.5),
    Builder: func(ctx core.BuildContext, dismiss func()) core.Widget {
        return widgets.Container{
            Width: 400, Color: myColor, BorderRadius: 8,
            Child: myContent(dismiss),
        }
    },
})
```

## Theming

Dialog appearance is controlled by `DialogThemeData`:

| Property | Default | Description |
|----------|---------|-------------|
| `BackgroundColor` | `SurfaceContainerHigh` | Dialog surface color |
| `BorderRadius` | `28` | Corner radius (Material 3) |
| `Elevation` | `3` | Shadow level (1-5) |
| `Padding` | `24px all sides` | Inner padding |
| `TitleContentSpacing` | `16` | Vertical gap between title and content in AlertDialog |
| `ContentActionsSpacing` | `24` | Vertical gap above the actions row in AlertDialog |
| `ActionSpacing` | `8` | Horizontal gap between action buttons in AlertDialog |
| `AlertDialogWidth` | `280` | Default width for AlertDialog when Width is zero |

Override via `ThemeData.DialogTheme`:

```go
themeData.DialogTheme = &theme.DialogThemeData{
    BackgroundColor: colors.Surface,
    BorderRadius:    16,
    Elevation:       2,
    Padding:         layout.EdgeInsetsAll(16),
}
```

## Related

- [Overlays](/docs/guides/overlay) for the underlying overlay system
- [Button](/docs/catalog/input/button) for dialog action buttons
- [Theming](/docs/guides/theming) for customizing dialog appearance
