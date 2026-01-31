package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/testing/internal/testbed"
)

func TestCaptureSnapshot_NotNil(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.LayoutBox{
		Width: 200, Height: 100,
		Color: graphics.RGB(255, 0, 0),
	})

	snap := tester.CaptureSnapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.RenderTree == nil {
		t.Fatal("expected non-nil render tree")
	}
}

func TestCaptureSnapshot_RenderTreeStructure(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.SetSize(graphics.Size{Width: 200, Height: 100})
	tester.PumpWidget(testbed.LayoutBox{
		Width: 200, Height: 100,
		Color: graphics.RGB(0, 255, 0),
	})

	snap := tester.CaptureSnapshot()
	root := snap.RenderTree
	if root == nil {
		t.Fatal("expected render tree root")
	}
	if root.Type == "" {
		t.Error("expected render tree root to have a type")
	}
	if root.ID == "" {
		t.Error("expected render tree root to have an ID")
	}
}

func TestSnapshot_Diff_Equal(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.LayoutBox{Width: 50, Height: 50})

	a := tester.CaptureSnapshot()
	b := tester.CaptureSnapshot()

	if diff := a.Diff(b); diff != "" {
		t.Errorf("expected no diff for identical snapshots, got:\n%s", diff)
	}
}

func TestSnapshot_Diff_Different(t *testing.T) {
	tester := NewWidgetTesterWithT(t)

	tester.PumpWidget(testbed.LayoutBox{Width: 50, Height: 50, Color: graphics.RGB(255, 0, 0)})
	a := tester.CaptureSnapshot()

	tester.PumpWidget(testbed.LayoutBox{Width: 100, Height: 50, Color: graphics.RGB(0, 255, 0)})
	b := tester.CaptureSnapshot()

	if diff := a.Diff(b); diff == "" {
		t.Error("expected diff for different snapshots")
	}
}

func TestSnapshot_UpdateAndMatch(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.LayoutBox{Width: 80, Height: 40})

	snap := tester.CaptureSnapshot()

	dir := t.TempDir()
	path := filepath.Join(dir, "testdata", "box.snapshot.json")

	if err := snap.UpdateFile(path); err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("snapshot file should exist after UpdateFile")
	}

	// MatchesFile should pass now
	snap.MatchesFile(t, path)
}

func TestSnapshot_MatchesFile_MissingFile(t *testing.T) {
	t.Setenv("DRIFT_UPDATE_SNAPSHOTS", "")
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.LayoutBox{Width: 50, Height: 50})
	snap := tester.CaptureSnapshot()

	// Use a recorder to intercept the Fatal
	failed := false
	sub := &fatalRecorder{name: t.Name(), onFatal: func() { failed = true }}
	snap.MatchesFile(sub, "/nonexistent/path/snap.json")

	if !failed {
		t.Error("expected MatchesFile to fail for missing file")
	}
}

func TestSnapshot_MatchesFile_Mismatch(t *testing.T) {
	t.Setenv("DRIFT_UPDATE_SNAPSHOTS", "")
	tester := NewWidgetTesterWithT(t)

	// Create snapshot for one widget
	tester.PumpWidget(testbed.LayoutBox{Width: 50, Height: 50, Color: graphics.RGB(255, 0, 0)})
	first := tester.CaptureSnapshot()

	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")
	first.UpdateFile(path)

	// Capture different widget (different color produces different display ops)
	tester.PumpWidget(testbed.LayoutBox{Width: 999, Height: 999, Color: graphics.RGB(0, 0, 255)})
	second := tester.CaptureSnapshot()

	errored := false
	sub := &errorRecorder{name: t.Name(), onError: func() { errored = true }}
	second.MatchesFile(sub, path)

	if !errored {
		t.Error("expected MatchesFile to report error for mismatch")
	}
}

func TestSnapshot_UpdateMode(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.LayoutBox{Width: 60, Height: 30})
	snap := tester.CaptureSnapshot()

	dir := t.TempDir()
	path := filepath.Join(dir, "update.snapshot.json")

	t.Setenv("DRIFT_UPDATE_SNAPSHOTS", "1")
	snap.MatchesFile(t, path)

	// File should now exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("snapshot file should be created in update mode")
	}
}

// fatalRecorder intercepts Fatalf calls for testing MatchesFile failures.
type fatalRecorder struct {
	name    string
	onFatal func()
}

func (r *fatalRecorder) Fatalf(format string, args ...any) { r.onFatal() }
func (r *fatalRecorder) Errorf(format string, args ...any) {}
func (r *fatalRecorder) Helper()                           {}
func (r *fatalRecorder) Name() string                      { return r.name }

// errorRecorder intercepts Errorf calls for testing MatchesFile mismatches.
type errorRecorder struct {
	name    string
	onError func()
}

func (r *errorRecorder) Fatalf(format string, args ...any) {}
func (r *errorRecorder) Errorf(format string, args ...any) { r.onError() }
func (r *errorRecorder) Helper()                           {}
func (r *errorRecorder) Name() string                      { return r.name }
