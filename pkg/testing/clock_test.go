package testing

import (
	"testing"
	"time"

	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/testing/internal/testbed"
)

func TestFakeClock_Advance(t *testing.T) {
	clk := NewFakeClock()
	start := clk.Now()

	clk.Advance(100 * time.Millisecond)
	elapsed := clk.Now().Sub(start)

	if elapsed != 100*time.Millisecond {
		t.Errorf("expected 100ms elapsed, got %v", elapsed)
	}
}

func TestFakeClock_Set(t *testing.T) {
	clk := NewFakeClock()
	target := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	clk.Set(target)
	if !clk.Now().Equal(target) {
		t.Errorf("expected %v, got %v", target, clk.Now())
	}
}

func TestWidgetTester_Clock(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	clk := tester.Clock()

	if clk == nil {
		t.Fatal("expected non-nil clock")
	}

	start := clk.Now()
	clk.Advance(500 * time.Millisecond)
	if clk.Now().Sub(start) != 500*time.Millisecond {
		t.Error("clock advancement not reflected")
	}
}

func TestAnimatedBox_ClockAdvance(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.SetSize(graphics.Size{Width: 400, Height: 100})
	tester.PumpWidget(testbed.AnimatedBox{
		Duration: 1 * time.Second,
		From:     50,
		To:       200,
		Height:   100,
	})

	// Check the widget's requested width (not constrained render size)
	result := tester.Find(ByType[testbed.LayoutBox]())
	if !result.Exists() {
		t.Fatal("expected to find LayoutBox rendered by AnimatedBox")
	}
	initialWidth := result.Widget().(testbed.LayoutBox).Width

	// Advance to ~halfway
	tester.Clock().Advance(500 * time.Millisecond)
	tester.Pump()

	midResult := tester.Find(ByType[testbed.LayoutBox]())
	if !midResult.Exists() {
		t.Fatal("expected to find LayoutBox after clock advance")
	}
	midWidth := midResult.Widget().(testbed.LayoutBox).Width

	// Width should have changed from initial
	if midWidth == initialWidth {
		t.Errorf("expected width to change after advancing clock, still %v", midWidth)
	}

	// Advance past the end of the animation
	tester.Clock().Advance(600 * time.Millisecond)
	tester.Pump()

	finalResult := tester.Find(ByType[testbed.LayoutBox]())
	if !finalResult.Exists() {
		t.Fatal("expected to find LayoutBox after animation complete")
	}
	finalWidth := finalResult.Widget().(testbed.LayoutBox).Width

	// Final width should be at or near the target
	if finalWidth < 190 || finalWidth > 210 {
		t.Errorf("expected final width ~200, got %v", finalWidth)
	}
}

func TestPumpAndSettle_AnimatedBox(t *testing.T) {
	tester := NewWidgetTesterWithT(t)
	tester.SetSize(graphics.Size{Width: 400, Height: 100})
	tester.PumpWidget(testbed.AnimatedBox{
		Duration: 100 * time.Millisecond,
		From:     10,
		To:       100,
		Height:   50,
	})

	// Advance past the animation duration
	tester.Clock().Advance(200 * time.Millisecond)

	err := tester.PumpAndSettle(time.Second)
	if err != nil {
		t.Errorf("expected settle after animation completes, got: %v", err)
	}
}
