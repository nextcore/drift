package main

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/platform"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

// buildSharePage creates a stateful widget for share demos.
func buildSharePage(ctx core.BuildContext) core.Widget {
	return sharePage{}
}

type sharePage struct{}

func (p sharePage) CreateElement() core.Element {
	return core.NewStatefulElement(p, nil)
}

func (p sharePage) Key() any {
	return nil
}

func (p sharePage) CreateState() core.State {
	return &shareState{}
}

type shareState struct {
	core.StateBase
	statusText     *core.ManagedState[string]
	textController *platform.TextEditingController
}

func (s *shareState) InitState() {
	s.statusText = core.NewManagedState(&s.StateBase, "Tap a button to open the share sheet.")
	s.textController = platform.NewTextEditingController("Check out this amazing app!")
}

func (s *shareState) Build(ctx core.BuildContext) core.Widget {
	_, colors, _ := theme.UseTheme(ctx)

	return demoPage(ctx, "Share",
		sectionTitle("Share Text", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "Enter custom text to share:", Style: labelStyle(colors)},
		widgets.VSpace(8),
		widgets.TextField{
			Label:        "Text to share",
			Controller:   s.textController,
			Placeholder:  "Enter text...",
			KeyboardType: platform.KeyboardTypeText,
			BorderRadius: 8,
		},
		widgets.VSpace(16),

		widgets.Button{
			Label: "Share Text",
			OnTap: func() {
				s.shareText()
			},
			Color:     colors.Primary,
			TextColor: colors.OnPrimary,
			Haptic:    true,
		},
		widgets.VSpace(24),

		sectionTitle("Share URL", colors),
		widgets.VSpace(12),
		widgets.Text{Content: "Share the Drift GitHub repository:", Style: labelStyle(colors)},
		widgets.VSpace(8),

		widgets.Button{
			Label: "Share URL",
			OnTap: func() {
				s.shareURL()
			},
			Color:     colors.Secondary,
			TextColor: colors.OnSecondary,
			Haptic:    true,
		},
		widgets.VSpace(24),

		sectionTitle("Result", colors),
		widgets.VSpace(12),
		statusCard(s.statusText.Get(), colors),
		widgets.VSpace(40),
	)
}

func (s *shareState) shareText() {
	text := s.textController.Text()
	if text == "" {
		s.statusText.Set("Please enter some text to share")
		return
	}

	result, err := platform.Share.ShareText(text)
	if err != nil {
		s.statusText.Set("Error sharing: " + err.Error())
		return
	}

	s.statusText.Set("Share result: " + string(result))
}

func (s *shareState) shareURL() {
	result, err := platform.Share.ShareURL("https://github.com/go-drift/drift")
	if err != nil {
		s.statusText.Set("Error sharing: " + err.Error())
		return
	}

	s.statusText.Set("Share result: " + string(result))
}
