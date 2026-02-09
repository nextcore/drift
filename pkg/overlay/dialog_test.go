package overlay

import (
	"testing"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"

	dtesting "github.com/go-drift/drift/pkg/testing"
)

// TestDialog_Build verifies that Dialog builds a Container with themed styling.
func TestDialog_Build(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(Dialog{
		Child: widgets.Text{Content: "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should render without panic and contain the text.
	if !tester.Find(dtesting.ByText("hello")).Exists() {
		t.Error("expected dialog child text to be found")
	}

	// Should have a Container in the tree.
	if !tester.Find(dtesting.ByType[widgets.Container]()).Exists() {
		t.Error("expected Container in dialog tree")
	}
}

// TestDialog_ExplicitWidth verifies that an explicit Width is applied.
func TestDialog_ExplicitWidth(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(Dialog{
		Width: 400,
		Child: widgets.Text{Content: "wide"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Find the Container and check its Width.
	result := tester.Find(dtesting.ByType[widgets.Container]())
	if !result.Exists() {
		t.Fatal("expected Container")
	}
	container := result.First().Widget().(widgets.Container)
	if container.Width != 400 {
		t.Errorf("expected width 400, got %f", container.Width)
	}
}

// TestDialog_ThemeStyling verifies that Dialog reads from DialogThemeData.
func TestDialog_ThemeStyling(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	// Use default theme; verify Container picks up the theme defaults.
	err := tester.PumpWidget(Dialog{
		Child: widgets.SizedBox{Width: 10, Height: 10},
	})
	if err != nil {
		t.Fatal(err)
	}

	result := tester.Find(dtesting.ByType[widgets.Container]())
	if !result.Exists() {
		t.Fatal("expected Container")
	}
	container := result.First().Widget().(widgets.Container)

	// Default dialog theme: border radius 28, non-zero color, shadow present.
	defaults := theme.DefaultDialogTheme(theme.LightColorScheme())
	if container.BorderRadius != defaults.BorderRadius {
		t.Errorf("expected border radius %f, got %f", defaults.BorderRadius, container.BorderRadius)
	}
	if container.Color != defaults.BackgroundColor {
		t.Errorf("expected background color %v, got %v", defaults.BackgroundColor, container.Color)
	}
	if container.Shadow == nil {
		t.Error("expected shadow to be set")
	}
}

// TestAlertDialog_Build verifies that AlertDialog builds title, content, and actions.
func TestAlertDialog_Build(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Title:   widgets.Text{Content: "Title"},
		Content: widgets.Text{Content: "Body"},
		Actions: []core.Widget{
			widgets.Text{Content: "OK"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !tester.Find(dtesting.ByText("Title")).Exists() {
		t.Error("expected title text")
	}
	if !tester.Find(dtesting.ByText("Body")).Exists() {
		t.Error("expected content text")
	}
	if !tester.Find(dtesting.ByText("OK")).Exists() {
		t.Error("expected action text")
	}
}

// TestAlertDialog_DefaultWidth verifies that AlertDialog defaults to 280 width.
func TestAlertDialog_DefaultWidth(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Title: widgets.Text{Content: "T"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// The AlertDialog wraps in a Dialog; find the Dialog's Container.
	result := tester.Find(dtesting.ByType[widgets.Container]())
	if !result.Exists() {
		t.Fatal("expected Container")
	}
	container := result.First().Widget().(widgets.Container)
	expectedWidth := theme.DefaultDialogTheme(theme.LightColorScheme()).AlertDialogWidth
	if container.Width != expectedWidth {
		t.Errorf("expected default width %v, got %f", expectedWidth, container.Width)
	}
}

// TestAlertDialog_CustomWidth verifies that a custom Width overrides the default.
func TestAlertDialog_CustomWidth(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Title: widgets.Text{Content: "T"},
		Width: 350,
	})
	if err != nil {
		t.Fatal(err)
	}

	result := tester.Find(dtesting.ByType[widgets.Container]())
	if !result.Exists() {
		t.Fatal("expected Container")
	}
	container := result.First().Widget().(widgets.Container)
	if container.Width != 350 {
		t.Errorf("expected width 350, got %f", container.Width)
	}
}

// TestAlertDialog_TitleOnly verifies AlertDialog works with only a title.
func TestAlertDialog_TitleOnly(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Title: widgets.Text{Content: "Only Title"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !tester.Find(dtesting.ByText("Only Title")).Exists() {
		t.Error("expected title text")
	}
}

// TestAlertDialog_ActionsRow verifies that actions are placed in a Row.
func TestAlertDialog_ActionsRow(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Actions: []core.Widget{
			widgets.Text{Content: "Cancel"},
			widgets.Text{Content: "OK"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should have a Row containing the actions.
	result := tester.Find(dtesting.ByType[widgets.Row]())
	if !result.Exists() {
		t.Error("expected Row for actions")
	}
}

// TestAlertDialog_ActionsGap verifies that AlertDialog inserts 8px gaps between actions.
func TestAlertDialog_ActionsGap(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{
		Actions: []core.Widget{
			widgets.Text{Content: "A"},
			widgets.Text{Content: "B"},
			widgets.Text{Content: "C"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	result := tester.Find(dtesting.ByType[widgets.Row]())
	if !result.Exists() {
		t.Fatal("expected Row for actions")
	}
	row := result.First().Widget().(widgets.Row)
	// 3 actions + 2 gaps = 5 children.
	if len(row.Children) != 5 {
		t.Errorf("expected 5 row children (3 actions + 2 gaps), got %d", len(row.Children))
	}
}

// TestAlertDialog_Empty verifies that an empty AlertDialog builds without panic.
func TestAlertDialog_Empty(t *testing.T) {
	tester := dtesting.NewWidgetTesterWithT(t)

	err := tester.PumpWidget(AlertDialog{})
	if err != nil {
		t.Fatal(err)
	}

	// Should have a Dialog (Container) with an empty Column.
	if !tester.Find(dtesting.ByType[widgets.Container]()).Exists() {
		t.Error("expected Container from Dialog wrapper")
	}
}
