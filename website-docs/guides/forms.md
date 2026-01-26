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

        widgets.NewButton("Submit", func() {
            if form.Validate() {
                form.Save()
                // Use f.parent.email and f.parent.password
            }
        }),
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

### Checkbox

```go
widgets.Checkbox{
    Value: isChecked,
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
// Native switch
widgets.Switch{
    Value:       isEnabled,
    OnTintColor: colors.Primary,
    OnChanged: func(value bool) {
        s.SetState(func() {
            isEnabled = value
        })
    },
}

// Drift-rendered toggle
widgets.Toggle{
    Value: isEnabled,
    OnChanged: func(value bool) {
        s.SetState(func() {
            isEnabled = value
        })
    },
}
```

### Radio

```go
widgets.Radio[string]{
    Value:      "email",
    GroupValue: selectedOption,
    OnChanged: func(value string) {
        s.SetState(func() {
            selectedOption = value
        })
    },
}
```

### Dropdown

```go
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
    BorderRadius: 8,
}
```

## Date and Time Pickers

Native modal pickers for date and time selection.

### DatePicker

```go
widgets.DatePicker{
    Value: selectedDate, // *time.Time, nil shows placeholder
    OnChanged: func(date time.Time) {
        s.SetState(func() {
            selectedDate = &date
        })
    },
    Placeholder: "Select date",
    Decoration: &widgets.InputDecoration{
        LabelText:    "Birth Date",
        BorderRadius: 8,
    },
}
```

### TimePicker

```go
widgets.TimePicker{
    Hour:   selectedHour,
    Minute: selectedMinute,
    OnChanged: func(hour, minute int) {
        s.SetState(func() {
            selectedHour = hour
            selectedMinute = minute
        })
    },
    Decoration: &widgets.InputDecoration{
        LabelText:    "Appointment Time",
        BorderRadius: 8,
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
// Indeterminate (spinning)
widgets.CircularProgressIndicator{
    Value: nil,
    Size:  36,
}

// Determinate (0.0 to 1.0)
progress := 0.65
widgets.CircularProgressIndicator{
    Value:       &progress,
    Size:        48,
    StrokeWidth: 5,
    Color:       colors.Primary,
    TrackColor:  colors.SurfaceVariant,
}
```

### LinearProgressIndicator

Drift-rendered linear progress bar. Set `Value` to `nil` for indeterminate animation.

```go
// Indeterminate
widgets.LinearProgressIndicator{
    Value: nil,
}

// Determinate
progress := 0.35
widgets.LinearProgressIndicator{
    Value:        &progress,
    Height:       6,
    BorderRadius: 3,
    Color:        colors.Primary,
    TrackColor:   colors.SurfaceVariant,
}
```

## Next Steps

- [State Management](/docs/guides/state-management) - Managing widget state
- [Layout](/docs/guides/layout) - Arranging widgets
- [Widgets](/docs/guides/widgets) - Full widget reference
