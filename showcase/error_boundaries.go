package main

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/errors"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

// buildErrorBoundariesPage creates a stateful widget for the error boundaries demo.
func buildErrorBoundariesPage(ctx core.BuildContext) core.Widget {
	return errorBoundariesPage{}
}

type errorBoundariesPage struct{}

func (e errorBoundariesPage) CreateElement() core.Element {
	return core.NewStatefulElement(e, nil)
}

func (e errorBoundariesPage) Key() any {
	return nil
}

func (e errorBoundariesPage) CreateState() core.State {
	return &errorBoundariesState{}
}

type errorBoundariesState struct {
	core.StateBase
	triggerStatelessPanic bool
	triggerStatefulPanic  bool
	boundaryKey           int // Used to reset the boundary
}

func (s *errorBoundariesState) InitState() {}

func (s *errorBoundariesState) Build(ctx core.BuildContext) core.Widget {
	_, colors, _ := theme.UseTheme(ctx)

	return demoPage(ctx, "Error Boundaries",
		// Introduction
		sectionTitle("Graceful Error Handling", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "Error boundaries catch panics in widget builds and display " +
			"fallback UI instead of crashing the app.", Style: labelStyle(colors)},
		widgets.VSpace(24),

		// Section: Unhandled Panics
		sectionTitle("Unhandled Panics", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "When a widget panics without an ErrorBoundary, the global " +
			"ErrorWidget is displayed:", Style: labelStyle(colors)},
		widgets.VSpace(12),
		widgets.Row{
			MainAxisAlignment: widgets.MainAxisAlignmentStart,
			ChildrenWidgets: []core.Widget{
				widgets.Button{
					Label: "Trigger Stateless Panic",
					OnTap: func() {
						s.SetState(func() {
							s.triggerStatelessPanic = true
						})
					},
					Color:     colors.Error,
					TextColor: colors.OnError,
					Haptic:    true,
				},
				widgets.HSpace(12),
				widgets.Button{
					Label: "Reset",
					OnTap: func() {
						s.SetState(func() {
							s.triggerStatelessPanic = false
						})
					},
					Color:     colors.SurfaceVariant,
					TextColor: colors.OnSurfaceVariant,
					Haptic:    true,
				},
			},
		},
		widgets.VSpace(12),
		s.buildPanicDemo(colors),
		widgets.VSpace(24),

		// Section: ErrorBoundary Widget
		sectionTitle("ErrorBoundary Widget", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "Wrap risky widgets in an ErrorBoundary to catch errors locally " +
			"without affecting the rest of the app:", Style: labelStyle(colors)},
		widgets.VSpace(12),
		widgets.Row{
			MainAxisAlignment: widgets.MainAxisAlignmentStart,
			ChildrenWidgets: []core.Widget{
				widgets.Button{
					Label: "Trigger Bounded Panic",
					OnTap: func() {
						s.SetState(func() {
							s.triggerStatefulPanic = true
						})
					},
					Color:     colors.Error,
					TextColor: colors.OnError,
					Haptic:    true,
				},
				widgets.HSpace(12),
				widgets.Button{
					Label: "Reset Boundary",
					OnTap: func() {
						s.SetState(func() {
							s.triggerStatefulPanic = false
							s.boundaryKey++ // Force boundary rebuild
						})
					},
					Color:     colors.SurfaceVariant,
					TextColor: colors.OnSurfaceVariant,
					Haptic:    true,
				},
			},
		},
		widgets.VSpace(12),
		s.buildBoundaryDemo(colors),
		widgets.VSpace(24),

		// Section: Custom Fallback
		sectionTitle("Custom Fallback Builder", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "ErrorBoundary accepts a FallbackBuilder for custom error UI:", Style: labelStyle(colors)},
		widgets.VSpace(12),
		s.buildCustomFallbackDemo(colors),
		widgets.VSpace(24),

		// Section: Code Example
		sectionTitle("Usage Example", colors),
		widgets.VSpace(12),
		s.buildCodeExample(colors),
		widgets.VSpace(40),
	)
}

// buildPanicDemo shows a widget that may panic based on state.
func (s *errorBoundariesState) buildPanicDemo(colors theme.ColorScheme) core.Widget {
	if s.triggerStatelessPanic {
		return panicWidget{message: "Demo panic in stateless widget!"}
	}
	return widgets.Container{
		Color:   colors.SurfaceVariant,
		Padding: layout.EdgeInsetsAll(16),
		ChildWidget: widgets.Text{
			Content: "This widget is working normally.",
			Style: graphics.TextStyle{
				Color:    colors.OnSurfaceVariant,
				FontSize: 14,
			},
		},
	}
}

// buildBoundaryDemo shows an ErrorBoundary catching a panic.
func (s *errorBoundariesState) buildBoundaryDemo(colors theme.ColorScheme) core.Widget {
	var child core.Widget
	if s.triggerStatefulPanic {
		child = panicWidget{message: "Demo panic caught by boundary!"}
	} else {
		child = widgets.Container{
			Color:   colors.SurfaceVariant,
			Padding: layout.EdgeInsetsAll(16),
			ChildWidget: widgets.Text{
				Content: "Content inside ErrorBoundary - working normally.",
				Style: graphics.TextStyle{
					Color:    colors.OnSurfaceVariant,
					FontSize: 14,
				},
			},
		}
	}

	return widgets.ErrorBoundary{
		OnError: func(err *errors.BoundaryError) {
			// Log the error (in production, send to analytics)
			_ = err // Logged automatically by the framework
		},
		ChildWidget: child,
	}
}

// buildCustomFallbackDemo shows an ErrorBoundary with custom fallback UI.
func (s *errorBoundariesState) buildCustomFallbackDemo(colors theme.ColorScheme) core.Widget {
	return widgets.ErrorBoundary{
		FallbackBuilder: func(err *errors.BoundaryError) core.Widget {
			return widgets.Container{
				Color:   colors.ErrorContainer,
				Padding: layout.EdgeInsetsAll(16),
				ChildWidget: widgets.Column{
					CrossAxisAlignment: widgets.CrossAxisAlignmentStart,
					MainAxisSize:       widgets.MainAxisSizeMin,
					ChildrenWidgets: []core.Widget{
						widgets.Text{
							Content: "Custom Error Handler",
							Style: graphics.TextStyle{
								Color:      colors.OnErrorContainer,
								FontSize:   16,
								FontWeight: graphics.FontWeightBold,
							},
						},
						widgets.VSpace(8),
						widgets.Text{
							Content: "Widget: " + err.Widget,
							Style: graphics.TextStyle{
								Color:    colors.OnErrorContainer,
								FontSize: 12,
							},
						},
					},
				},
			}
		},
		ChildWidget: alwaysPanicWidget{},
	}
}

// buildCodeExample shows example code for using ErrorBoundary.
func (s *errorBoundariesState) buildCodeExample(colors theme.ColorScheme) core.Widget {
	code := `widgets.ErrorBoundary{
    OnError: func(err *errors.BuildError) {
        log.Printf("Error: %v", err)
    },
    FallbackBuilder: func(err *errors.BuildError) core.Widget {
        return widgets.Text{Content: "Something went wrong"}
    },
    ChildWidget: RiskyContent{},
}`

	return widgets.Container{
		Color:   colors.SurfaceVariant,
		Padding: layout.EdgeInsetsAll(16),
		ChildWidget: widgets.Text{
			Content: code,
			Style: graphics.TextStyle{
				Color:    colors.OnSurfaceVariant,
				FontSize: 12,
			},
		},
	}
}

// panicWidget is a widget that panics during build.
type panicWidget struct {
	message string
}

func (p panicWidget) CreateElement() core.Element {
	return core.NewStatelessElement(p, nil)
}

func (p panicWidget) Key() any {
	return nil
}

func (p panicWidget) Build(ctx core.BuildContext) core.Widget {
	panic(p.message)
}

// alwaysPanicWidget always panics - used to demo custom fallback.
type alwaysPanicWidget struct{}

func (a alwaysPanicWidget) CreateElement() core.Element {
	return core.NewStatelessElement(a, nil)
}

func (a alwaysPanicWidget) Key() any {
	return nil
}

func (a alwaysPanicWidget) Build(ctx core.BuildContext) core.Widget {
	panic("This widget always panics to demonstrate custom fallback")
}
