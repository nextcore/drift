---
id: forms
title: Forms
sidebar_position: 3
---

# Forms

Build forms with validation, selection controls, and native pickers.

## Form Validation

Use `Form` with `TextFormField` for validated text input. The form tracks all fields and provides `Validate()`, `Save()`, and `Reset()` methods.

```go
type loginState struct {
    core.StateBase
    email    string
    password string
}

func (s *loginState) Build(ctx core.BuildContext) core.Widget {
    return widgets.Form{
        Autovalidate: true,
        ChildWidget:  loginForm{parent: s},
    }
}

// Separate widget to access FormOf(ctx)
type loginForm struct {
    parent *loginState
}

func (f loginForm) Build(ctx core.BuildContext) core.Widget {
    form := widgets.FormOf(ctx)

    return widgets.ColumnOf(
        widgets.MainAxisAlignmentStart,
        widgets.CrossAxisAlignmentStretch,
        widgets.MainAxisSizeMin,

        widgets.TextFormField{
            Label:        "Email",
            Placeholder:  "you@example.com",
            KeyboardType: platform.KeyboardTypeEmail,
            InputAction:  platform.TextInputActionNext,
            BorderRadius: 8,
            Validator: func(value string) string {
                if value == "" {
                    return "Email is required"
                }
                if !strings.Contains(value, "@") {
                    return "Please enter a valid email"
                }
                return ""
            },
            OnSaved: func(value string) {
                f.parent.email = value
            },
        },
        widgets.VSpace(16),

        widgets.TextFormField{
            Label:        "Password",
            Placeholder:  "Enter password",
            InputAction:  platform.TextInputActionDone,
            Obscure:      true,
            BorderRadius: 8,
            Validator: func(value string) string {
                if len(value) < 8 {
                    return "Password must be at least 8 characters"
                }
                return ""
            },
            OnSaved: func(value string) {
                f.parent.password = value
            },
        },
        widgets.VSpace(24),

        widgets.Button{
            Label: "Submit",
            OnTap: func() {
                if form.Validate() {
                    form.Save()
                    // Use f.parent.email and f.parent.password
                }
            },
            Haptic: true,
        },
    )
}
```

### Form Methods

| Method | Description |
|--------|-------------|
| `Validate()` | Run validators on all fields, returns `bool` |
| `Save()` | Call `OnSaved` for all fields |
| `Reset()` | Reset all fields to initial values |

### TextFormField Options

| Field | Description |
|-------|-------------|
| `InitialValue` | Starting value when no Controller is provided |
| `Validator` | Returns error message or empty string if valid |
| `OnSaved` | Called when the form is saved |
| `OnChanged` | Called when the field value changes |
| `Autovalidate` | Validate on every change |
| `Obscure` | Hide text (for passwords) |
| `KeyboardType` | Keyboard type (`KeyboardTypeEmail`, `KeyboardTypeNumber`, etc.) |
| `InputAction` | Action button (`TextInputActionNext`, `TextInputActionDone`, etc.) |

## Selection Controls

Selection controls can be created explicitly or with themed constructors from `pkg/theme`.

### Checkbox

```go
// Themed checkbox (recommended)
theme.CheckboxOf(ctx, isChecked, func(value bool) {
    s.SetState(func() {
        isChecked = value
    })
})

// Explicit checkbox
widgets.Checkbox{
    Value:       isChecked,
    ActiveColor: colors.Primary,
    CheckColor:  colors.OnPrimary,
    OnChanged: func(value bool) {
        s.SetState(func() {
            isChecked = value
        })
    },
}
```

### Switch and Toggle

`Switch` uses native platform controls (UISwitch/SwitchCompat). `Toggle` is Drift-rendered.

```go
// Native switch (no themed constructor - use struct literal)
widgets.Switch{
    Value:       isEnabled,
    OnTintColor: colors.Primary,
    OnChanged: func(value bool) {
        s.SetState(func() {
            isEnabled = value
        })
    },
}

// Themed toggle (Drift-rendered)
theme.ToggleOf(ctx, isEnabled, func(value bool) {
    s.SetState(func() {
        isEnabled = value
    })
})

// Explicit toggle
widgets.Toggle{
    Value:         isEnabled,
    ActiveColor:   colors.Primary,
    InactiveColor: colors.SurfaceVariant,
    OnChanged: func(value bool) {
        s.SetState(func() {
            isEnabled = value
        })
    },
}
```

### Radio

```go
// Themed radio (recommended)
theme.RadioOf(ctx, "email", selectedOption, func(value string) {
    s.SetState(func() {
        selectedOption = value
    })
})

// Explicit radio
widgets.Radio[string]{
    Value:       "email",
    GroupValue:  selectedOption,
    ActiveColor: colors.Primary,
    OnChanged: func(value string) {
        s.SetState(func() {
            selectedOption = value
        })
    },
}
```

### Dropdown

```go
// Themed dropdown (recommended)
theme.DropdownOf(ctx, selectedPlan, []widgets.DropdownItem[string]{
    {Value: "starter", Label: "Starter"},
    {Value: "pro", Label: "Pro"},
    {Value: "enterprise", Label: "Enterprise"},
}, func(value string) {
    s.SetState(func() {
        selectedPlan = value
    })
}).WithBorderRadius(8)

// Explicit dropdown
widgets.Dropdown[string]{
    Value: selectedPlan,
    Hint:  "Select a plan",
    Items: []widgets.DropdownItem[string]{
        {Value: "starter", Label: "Starter"},
        {Value: "pro", Label: "Pro"},
        {Value: "enterprise", Label: "Enterprise"},
    },
    OnChanged: func(value string) {
        s.SetState(func() {
            selectedPlan = value
        })
    },
    BackgroundColor: colors.Surface,
    BorderColor:     colors.Outline,
    BorderRadius:    8,
    SelectedItemColor: colors.SurfaceVariant,
}
```

## Date and Time Pickers

Native modal pickers for date and time selection.

### DatePicker

```go
// Themed (recommended)
theme.DatePickerOf(ctx, selectedDate, func(date time.Time) {
    s.SetState(func() { selectedDate = &date })
})

// Explicit with full styling (no theme defaults)
widgets.DatePicker{
    Value: selectedDate, // *time.Time, nil shows placeholder
    OnChanged: func(date time.Time) {
        s.SetState(func() { selectedDate = &date })
    },
    Placeholder: "Select date",
    TextStyle:   graphics.TextStyle{FontSize: 16, Color: colors.OnSurface},
    Decoration: &widgets.InputDecoration{
        LabelText:       "Birth Date",
        BorderRadius:    8,
        BorderColor:     colors.Outline,
        BackgroundColor: colors.Surface,
        HintStyle:       graphics.TextStyle{FontSize: 16, Color: colors.OnSurfaceVariant},
        LabelStyle:      graphics.TextStyle{FontSize: 14, Color: colors.OnSurfaceVariant},
    },
}
```

### TimePicker

```go
// Themed (recommended)
theme.TimePickerOf(ctx, selectedHour, selectedMinute, func(h, m int) {
    s.SetState(func() { selectedHour, selectedMinute = h, m })
})

// Explicit with full styling (no theme defaults)
widgets.TimePicker{
    Hour:   selectedHour,
    Minute: selectedMinute,
    OnChanged: func(hour, minute int) {
        s.SetState(func() {
            selectedHour = hour
            selectedMinute = minute
        })
    },
    TextStyle: graphics.TextStyle{FontSize: 16, Color: colors.OnSurface},
    Decoration: &widgets.InputDecoration{
        LabelText:       "Appointment Time",
        BorderRadius:    8,
        BorderColor:     colors.Outline,
        BackgroundColor: colors.Surface,
        HintStyle:       graphics.TextStyle{FontSize: 16, Color: colors.OnSurfaceVariant},
        LabelStyle:      graphics.TextStyle{FontSize: 14, Color: colors.OnSurfaceVariant},
    },
}
```

## Progress Indicators

### ActivityIndicator

Native platform spinner (UIActivityIndicatorView on iOS, ProgressBar on Android).

```go
widgets.ActivityIndicator{
    Animating: true,
    Size:      widgets.ActivityIndicatorSizeMedium, // Small, Medium, Large
    Color:     colors.Primary, // Optional
}
```

### CircularProgressIndicator

Drift-rendered circular progress. Set `Value` to `nil` for indeterminate animation.

```go
// Themed (recommended)
theme.CircularProgressIndicatorOf(ctx, nil)  // indeterminate
theme.CircularProgressIndicatorOf(ctx, &progress)  // determinate

// Explicit (full control)
widgets.CircularProgressIndicator{
    Value:       nil,
    Size:        36,
    Color:       colors.Primary,
    TrackColor:  colors.SurfaceVariant,
    StrokeWidth: 4,
}
```

### LinearProgressIndicator

Drift-rendered linear progress bar. Set `Value` to `nil` for indeterminate animation.

```go
// Themed (recommended)
theme.LinearProgressIndicatorOf(ctx, nil)  // indeterminate
theme.LinearProgressIndicatorOf(ctx, &progress)  // determinate

// Explicit (full control)
widgets.LinearProgressIndicator{
    Value:        nil,
    Color:        colors.Primary,
    TrackColor:   colors.SurfaceVariant,
    Height:       4,
    BorderRadius: 2,
}
```

## Next Steps

- [State Management](/docs/guides/state-management) - Managing widget state
- [Layout](/docs/guides/layout) - Arranging widgets
- [Widgets](/docs/guides/widgets) - Full widget reference
