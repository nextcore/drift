// Package widgets provides UI components for building widget trees.
//
// This package contains the concrete widget implementations that developers
// use to build user interfaces, including layout widgets (Row, Column, Stack),
// display widgets (Text, Icon, Image), input widgets (Button, TextField, Checkbox),
// and container widgets (Container, Padding, SizedBox).
//
// # Widget Construction
//
// Drift widgets use a three-tier construction pattern:
//
// ## Tier 1: Struct Literal (canonical, full control)
//
//	btn := Button{
//	    Label:     "Submit",
//	    OnTap:     handleSubmit,
//	    Color:     colors.Primary,
//	    Disabled:  !isValid,
//	}
//
// This is the PRIMARY way to create widgets. All fields are accessible.
//
// ## Tier 2: XxxOf Helper (convenience for required params)
//
//	btn := ButtonOf("Submit", handleSubmit)
//	txt := TextOf("Hello", style)
//
// XxxOf helpers exist only for widgets with clear required parameters.
// They match struct literal defaults unless documented otherwise.
//
// ## Tier 3: WithX Chaining (optional configuration)
//
//	btn := ButtonOf("Submit", handleSubmit).
//	    WithColor(colors.Primary, colors.OnPrimary).
//	    WithDisabled(!isValid)
//
// WithX methods return COPIES; they never mutate the receiver.
//
// # API Rules
//
//   - Canonical = struct literal. Always works, always documented.
//   - XxxOf only when required params exist. No helpers for zero-param widgets.
//   - Exception: Variadic children helpers (StackOf, ColumnOf, RowOf) are allowed
//     for ergonomics even though empty collections are technically valid.
//   - XxxOf must match zero-value defaults (document exceptions).
//   - WithX returns copies. Doc comment must state "returns a copy".
//   - No helper overloads. Use XxxOf + WithX, not multiple XxxOfWithY variants.
//
// # Helper Defaults
//
// Some helpers set non-zero defaults for ergonomics:
//
//   - ButtonOf sets Haptic: true (provides tactile feedback by default)
//   - TextOf sets Wrap: true (text wraps to multiple lines by default)
//
// These defaults are documented in each helper's doc comment.
//
// # Layout Widgets
//
// Use Row and Column for horizontal and vertical layouts:
//
//	widgets.Row{ChildrenWidgets: []core.Widget{...}}
//	widgets.Column{ChildrenWidgets: []core.Widget{...}}
//
// Helper functions provide a more concise syntax:
//
//	widgets.RowOf(alignment, crossAlignment, size, child1, child2, child3)
//	widgets.ColumnOf(alignment, crossAlignment, size, child1, child2, child3)
//
// # Input Widgets
//
// Button, TextField, Checkbox, Radio, and Switch handle user input.
// Use the three-tier pattern for customization:
//
//	// Struct literal
//	widgets.Button{Label: "Submit", OnTap: onTap, Disabled: true}
//
//	// Helper with chaining
//	widgets.ButtonOf("Submit", onTap).WithPadding(padding).WithColor(bg, text)
//
// # Scrolling
//
// ScrollView provides scrollable content with customizable physics:
//
//	widgets.ScrollView{Child: content, Physics: widgets.BouncingScrollPhysics{}}
//
// # Style Guide for Widget Authors
//
// When adding WithX methods:
//   - Use value receiver, not pointer
//   - Return the modified copy, never mutate
//   - Doc comment: "returns a copy of X with..."
package widgets
