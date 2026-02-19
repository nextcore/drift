package widgets

import (
	"testing"

	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

type testLayoutCounterBox struct {
	layout.RenderBoxBase
	layoutCalls int
}

func (b *testLayoutCounterBox) PerformLayout() {
	b.layoutCalls++
	b.SetSize(graphics.Size{Width: 10, Height: 10})
}

func (b *testLayoutCounterBox) Paint(ctx *layout.PaintContext) {}

func (b *testLayoutCounterBox) HitTest(position graphics.Offset, result *layout.HitTestResult) bool {
	return false
}

func TestIndexedStack_ExpandLayoutsOnlyActiveChild(t *testing.T) {
	active := &testLayoutCounterBox{}
	active.SetSelf(active)
	inactive := &testLayoutCounterBox{}
	inactive.SetSelf(inactive)

	stack := &renderIndexedStack{
		alignment: layout.AlignmentTopLeft,
		fit:       StackFitExpand,
		index:     0,
	}
	stack.SetSelf(stack)
	stack.SetChildren([]layout.RenderObject{active, inactive})

	constraints := layout.Tight(graphics.Size{Width: 100, Height: 80})
	stack.Layout(constraints, true)

	if active.layoutCalls != 1 {
		t.Fatalf("expected active child to be laid out once, got %d", active.layoutCalls)
	}
	if inactive.layoutCalls != 0 {
		t.Fatalf("expected inactive child to skip layout in expand fit, got %d", inactive.layoutCalls)
	}
}

func TestIndexedStack_ExpandPositionsActiveChild(t *testing.T) {
	child := &testLayoutCounterBox{}
	child.SetSelf(child)

	pos := &renderPositioned{
		child: child,
		left:  ptrF64(8),
		top:   ptrF64(12),
	}
	pos.SetSelf(pos)

	stack := &renderIndexedStack{
		alignment: layout.AlignmentTopLeft,
		fit:       StackFitExpand,
		index:     0,
	}
	stack.SetSelf(stack)
	stack.SetChildren([]layout.RenderObject{pos})

	constraints := layout.Tight(graphics.Size{Width: 100, Height: 80})
	stack.Layout(constraints, true)

	data, ok := pos.ParentData().(*layout.BoxParentData)
	if !ok {
		t.Fatalf("expected positioned child to have BoxParentData")
	}
	if data.Offset.X != 8 || data.Offset.Y != 12 {
		t.Fatalf("expected positioned offset (8,12), got (%.1f,%.1f)", data.Offset.X, data.Offset.Y)
	}
	size := pos.Size()
	if size.Width != 10 || size.Height != 10 {
		t.Fatalf("expected positioned size 10x10 from child layout, got %.1fx%.1f", size.Width, size.Height)
	}
}

func ptrF64(v float64) *float64 { return &v }
