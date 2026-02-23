---
id: textfield
title: TextField
---

# TextField

Native text input with decoration, label, and helper text. `TextInput` is the base native control; `TextField` wraps it with label, placeholder, and styling.

## Basic Usage

```go
// Themed (recommended)
controller := platform.NewTextEditingController("")
theme.TextFieldOf(ctx, controller).
    WithLabel("Email").
    WithPlaceholder("you@example.com").
    WithOnChanged(func(value string) {
        s.SetState(func() { email = value })
    })

// Explicit (must provide all visual properties)
widgets.TextField{
    Controller:      controller,
    Label:           "Email",
    Placeholder:     "you@example.com",
    Height:          48,
    Padding:         layout.EdgeInsetsSymmetric(12, 8),
    BackgroundColor: colors.Surface,
    BorderColor:     colors.Outline,
    FocusColor:      colors.Primary,
    BorderWidth:     1,
    BorderRadius:    8,
    Style:           graphics.TextStyle{FontSize: 16, Color: colors.OnSurface},
    PlaceholderColor: colors.OnSurfaceVariant,
    OnChanged: func(value string) {
        s.SetState(func() { email = value })
    },
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `Controller` | `*platform.TextEditingController` | Manages text content and selection |
| `Label` | `string` | Label text above the field |
| `Placeholder` | `string` | Placeholder text when empty |
| `HelperText` | `string` | Helper text below the field (hidden when `ErrorText` is set) |
| `ErrorText` | `string` | Error text below the field (overrides `HelperText`) |
| `Height` | `float64` | Field height |
| `Padding` | `layout.EdgeInsets` | Inner padding |
| `BackgroundColor` | `graphics.Color` | Background color |
| `BorderColor` | `graphics.Color` | Border color |
| `FocusColor` | `graphics.Color` | Border color when focused |
| `BorderWidth` | `float64` | Border width |
| `BorderRadius` | `float64` | Corner radius |
| `Style` | `graphics.TextStyle` | Text style (font size, color) |
| `PlaceholderColor` | `graphics.Color` | Placeholder text color |
| `ErrorColor` | `graphics.Color` | Error text and border color when `ErrorText` is set |
| `OnChanged` | `func(string)` | Called when text changes |
| `OnSubmitted` | `func(string)` | Called when the user submits |
| `OnEditingComplete` | `func()` | Called when editing is complete |
| `KeyboardType` | `platform.KeyboardType` | Keyboard type (`KeyboardTypeEmail`, `KeyboardTypeNumber`, etc.) |
| `InputAction` | `platform.TextInputAction` | Action button (`TextInputActionNext`, `TextInputActionDone`, etc.) |
| `Obscure` | `bool` | Hide text (for passwords) |
| `Autocorrect` | `bool` | Enable auto-correction |
| `Disabled` | `bool` | Reject input when true |

## Explicit Styling Requirements

Explicit text fields only render what you set. If colors, sizes, or text styles are zero, the widget can be invisible or collapsed. You must set `Height`, `Padding`, `BackgroundColor`, `BorderColor`, `FocusColor`, `BorderWidth`, `Style` (FontSize + Color), and `PlaceholderColor`.

If you want defaults from the theme, prefer `theme.TextFieldOf(ctx, controller)`.

## Common Patterns

### Password Field

```go
theme.TextFieldOf(ctx, controller).
    WithLabel("Password").
    WithPlaceholder("Enter password").
    WithObscure(true).
    WithInputAction(platform.TextInputActionDone)
```

### Email Field

```go
theme.TextFieldOf(ctx, controller).
    WithLabel("Email").
    WithPlaceholder("you@example.com").
    WithKeyboardType(platform.KeyboardTypeEmail).
    WithInputAction(platform.TextInputActionNext)
```

## Related

- [Forms & Validation](/docs/guides/forms) for TextFormField with validation
- [Button](/docs/catalog/input/button) for form submission
