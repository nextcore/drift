package main

import (
	"time"

	"github.com/go-drift/drift/pkg/animation"
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

// buildAnimationsPage creates the animations demo page.
func buildAnimationsPage(ctx core.BuildContext) core.Widget {
	return pageScaffold(ctx, "Animations", AnimationsDemo{})
}

// AnimationsDemo showcases implicit animation widgets.
type AnimationsDemo struct{}

func (a AnimationsDemo) CreateElement() core.Element {
	return core.NewStatefulElement(a, nil)
}

func (a AnimationsDemo) Key() any {
	return nil
}

func (a AnimationsDemo) CreateState() core.State {
	return &animationsDemoState{}
}

type animationsDemoState struct {
	core.StateBase
	containerExpanded bool
	containerColorIdx int
	opacityVisible    bool
}

func (s *animationsDemoState) InitState() {
	s.containerExpanded = false
	s.containerColorIdx = 0
	s.opacityVisible = true
}

func (s *animationsDemoState) Build(ctx core.BuildContext) core.Widget {
	_, colors, textTheme := theme.UseTheme(ctx)

	// Colors to cycle through for AnimatedContainer demo
	containerColors := []graphics.Color{
		colors.Primary,
		colors.Secondary,
		colors.Error,
		colors.Outline,
	}
	currentColor := containerColors[s.containerColorIdx%len(containerColors)]

	// Container size based on expanded state
	containerWidth := 100.0
	containerHeight := 100.0
	if s.containerExpanded {
		containerWidth = 200.0
		containerHeight = 150.0
	}

	return widgets.ScrollView{
		ScrollDirection: widgets.AxisVertical,
		Physics:         widgets.BouncingScrollPhysics{},
		Padding:         layout.EdgeInsetsAll(20),
		ChildWidget: widgets.ColumnOf(
			widgets.MainAxisAlignmentStart,
			widgets.CrossAxisAlignmentStart,
			widgets.MainAxisSizeMin,

			// AnimatedContainer section
			widgets.Text{Content: "AnimatedContainer", Style: textTheme.TitleLarge},
			widgets.VSpace(8),
			widgets.Text{Content: "Automatically animates size, color, and padding changes", Style: graphics.TextStyle{
				Color:    colors.OnSurfaceVariant,
				FontSize: 14,
			}},
			widgets.VSpace(16),

			// AnimatedContainer demo
			widgets.AnimatedContainer{
				Duration:  300 * time.Millisecond,
				Curve:     animation.EaseInOut,
				Width:     containerWidth,
				Height:    containerHeight,
				Color:     currentColor,
				Alignment: layout.AlignmentCenter,
				ChildWidget: widgets.Text{Content: "Tap buttons", Style: graphics.TextStyle{
					Color:      colors.OnPrimary,
					FontSize:   14,
					FontWeight: graphics.FontWeightBold,
				}},
			},
			widgets.VSpace(16),

			// Controls for AnimatedContainer
			widgets.RowOf(
				widgets.MainAxisAlignmentCenter,
				widgets.CrossAxisAlignmentCenter,
				widgets.MainAxisSizeMax,
				widgets.Button{
					Label: "Toggle Size",
					OnTap: func() {
						s.SetState(func() {
							s.containerExpanded = !s.containerExpanded
						})
					},
					Color:     colors.SurfaceVariant,
					TextColor: colors.OnSurfaceVariant,
					Padding:   layout.EdgeInsetsSymmetric(16, 10),
					FontSize:  14,
					Haptic:    true,
				},
				widgets.HSpace(12),
				widgets.Button{
					Label: "Change Color",
					OnTap: func() {
						s.SetState(func() {
							s.containerColorIdx++
						})
					},
					Color:     colors.SurfaceVariant,
					TextColor: colors.OnSurfaceVariant,
					Padding:   layout.EdgeInsetsSymmetric(16, 10),
					FontSize:  14,
					Haptic:    true,
				},
			),

			widgets.VSpace(40),

			// AnimatedOpacity section
			widgets.Text{Content: "AnimatedOpacity", Style: textTheme.TitleLarge},
			widgets.VSpace(8),
			widgets.Text{Content: "Smoothly fades widgets in and out", Style: graphics.TextStyle{
				Color:    colors.OnSurfaceVariant,
				FontSize: 14,
			}},
			widgets.VSpace(16),

			// AnimatedOpacity demo
			widgets.AnimatedOpacity{
				Duration: 500 * time.Millisecond,
				Curve:    animation.EaseOut,
				Opacity:  boolToOpacity(s.opacityVisible),
				ChildWidget: widgets.Container{
					Width:     150,
					Height:    80,
					Color:     colors.Secondary,
					Alignment: layout.AlignmentCenter,
					ChildWidget: widgets.Text{Content: "Fade me!", Style: graphics.TextStyle{
						Color:      colors.OnSecondary,
						FontSize:   16,
						FontWeight: graphics.FontWeightBold,
					}},
				},
			},
			widgets.VSpace(16),

			// Control for AnimatedOpacity
			widgets.Button{
				Label: "Toggle Visibility",
				OnTap: func() {
					s.SetState(func() {
						s.opacityVisible = !s.opacityVisible
					})
				},
				Color:     colors.SurfaceVariant,
				TextColor: colors.OnSurfaceVariant,
				Padding:   layout.EdgeInsetsSymmetric(20, 12),
				FontSize:  14,
				Haptic:    true,
			},

			widgets.VSpace(40),
		),
	}
}

func boolToOpacity(visible bool) float64 {
	if visible {
		return 1.0
	}
	return 0.0
}
