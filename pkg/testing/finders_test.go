package testing

import (
	"testing"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/testing/internal/testbed"
	"github.com/go-drift/drift/pkg/widgets"
)

func TestByType(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 0})

	result := tester.Find(ByType[widgets.Text]())
	if !result.Exists() {
		t.Fatal("expected to find Text widget")
	}
	text := result.Widget().(widgets.Text)
	if text.Content != "0" {
		t.Errorf("expected text '0', got %q", text.Content)
	}
}

func TestByText(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 42})

	if !tester.Find(ByText("42")).Exists() {
		t.Error("expected to find text '42'")
	}
	if tester.Find(ByText("99")).Exists() {
		t.Error("should not find text '99'")
	}
}

func TestByTextContaining(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 123})

	if !tester.Find(ByTextContaining("12")).Exists() {
		t.Error("expected to find text containing '12'")
	}
	if tester.Find(ByTextContaining("99")).Exists() {
		t.Error("should not find text containing '99'")
	}
}

func TestByType_Counter(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 5})

	result := tester.Find(ByType[testbed.Counter]())
	if !result.Exists() {
		t.Fatal("expected to find Counter widget")
	}
}

func TestByType_GestureDetector(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 0})

	result := tester.Find(ByType[widgets.GestureDetector]())
	if !result.Exists() {
		t.Fatal("expected to find GestureDetector widget inside Counter")
	}
}

func TestFinderResult_Count(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 0})

	result := tester.Find(ByType[widgets.Text]())
	if result.Count() != 1 {
		t.Errorf("expected 1 Text widget, got %d", result.Count())
	}
}

func TestFinderResult_FirstOrNil(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(widgets.Text{Content: "hello"})

	if tester.Find(ByText("hello")).FirstOrNil() == nil {
		t.Error("FirstOrNil should return element for existing text")
	}
	if tester.Find(ByText("missing")).FirstOrNil() != nil {
		t.Error("FirstOrNil should return nil for missing text")
	}
}

func TestFinderResult_First_PanicsOnEmpty(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(widgets.Text{Content: "hello"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected First() to panic on empty result")
		}
	}()
	tester.Find(ByText("missing")).First()
}

func TestByPredicate(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 7})

	result := tester.Find(ByPredicate(func(e core.Element) bool {
		if tw, ok := e.Widget().(widgets.Text); ok {
			return tw.Content == "7"
		}
		return false
	}))
	if !result.Exists() {
		t.Error("expected predicate to find text '7'")
	}
}

func TestDescendant(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(testbed.Counter{Initial: 0})

	result := tester.Find(Descendant(
		ByType[widgets.GestureDetector](),
		ByType[widgets.Text](),
	))
	if !result.Exists() {
		t.Error("expected to find Text as descendant of GestureDetector")
	}
}

func TestFinderResult_RenderObject(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.SetSize(graphics.Size{Width: 100, Height: 50})
	tester.PumpWidget(testbed.LayoutBox{Width: 100, Height: 50})

	result := tester.Find(ByType[testbed.LayoutBox]())
	if !result.Exists() {
		t.Fatal("expected to find LayoutBox")
	}
	ro := result.RenderObject()
	if ro == nil {
		t.Fatal("expected render object for LayoutBox")
	}
	size := ro.Size()
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("expected size 100x50, got %vx%v", size.Width, size.Height)
	}
}
