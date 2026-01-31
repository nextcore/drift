package testing

import (
	"testing"
	"time"

	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/testing/internal/testbed"
	"github.com/go-drift/drift/pkg/widgets"
)

func TestNewWidgetTester_Defaults(t *testing.T) {
	tester := NewWidgetTesterWithT(t)

	if tester.size.Width != DefaultTestWidth || tester.size.Height != DefaultTestHeight {
		t.Errorf("expected default size %dx%d, got %vx%v", DefaultTestWidth, DefaultTestHeight, tester.size.Width, tester.size.Height)
	}
	if tester.scale != DefaultScale {
		t.Errorf("expected default scale %v, got %v", DefaultScale, tester.scale)
	}
	if tester.clock == nil {
		t.Fatal("expected fake clock to be set")
	}
}

func TestPumpWidget_MountsTree(t *testing.T) {
	tester := NewWidgetTesterWithT(t)

	err := tester.PumpWidget(widgets.Text{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if tester.RootElement() == nil {
		t.Fatal("expected root element after PumpWidget")
	}
	if tester.RootRenderObject() == nil {
		t.Fatal("expected root render object after PumpWidget")
	}
}

func TestPumpWidget_Remount(t *testing.T) {
	tester := NewWidgetTesterWithT(t)

	tester.PumpWidget(widgets.Text{Content: "first"})
	first := tester.RootElement()

	tester.PumpWidget(widgets.Text{Content: "second"})
	second := tester.RootElement()

	if first == second {
		t.Error("expected new root element after remount")
	}
}

func TestSetSize(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.SetSize(graphics.Size{Width: 375, Height: 667})

	tester.PumpWidget(testbed.LayoutBox{Width: 375, Height: 667})

	ro := tester.RootRenderObject()
	if ro == nil {
		t.Fatal("no render object")
	}
	size := ro.Size()
	if size.Width != 375 || size.Height != 667 {
		t.Errorf("expected size 375x667, got %vx%v", size.Width, size.Height)
	}
}

func TestPumpAndSettle_IdleWidget(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(widgets.Text{Content: "static"})

	err := tester.PumpAndSettle(time.Second)
	if err != nil {
		t.Errorf("expected settle for static widget, got: %v", err)
	}
}

func TestDispatch(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.PumpWidget(widgets.Text{Content: "test"})

	called := false
	tester.Dispatch(func() { called = true })

	if called {
		t.Error("dispatch should not run until Pump")
	}

	tester.Pump()

	if !called {
		t.Error("dispatch should have run after Pump")
	}
}
