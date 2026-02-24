package widgets

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// getChildOffset extracts the offset from a child's parent data.
func getChildOffset(child layout.RenderBox) graphics.Offset {
	if child == nil {
		return graphics.Offset{}
	}
	if data, ok := child.ParentData().(*layout.BoxParentData); ok {
		return data.Offset
	}
	return graphics.Offset{}
}

// Root creates a top-level view widget with the given child.
func Root(child core.Widget) View {
	return View{Child: child}
}

// Centered wraps a child in a Center widget.
func Centered(child core.Widget) Center {
	return Center{Child: child}
}

// Padded wraps a child with the specified padding.
func Padded(padding layout.EdgeInsets, child core.Widget) Padding {
	return Padding{Padding: padding, Child: child}
}

// VSpace creates a fixed-height vertical spacer.
func VSpace(height float64) SizedBox {
	return SizedBox{Height: height}
}

// HSpace creates a fixed-width horizontal spacer.
func HSpace(width float64) SizedBox {
	return SizedBox{Width: width}
}

// Tap wraps a child with a tap handler.
func Tap(onTap func(), child core.Widget) GestureDetector {
	return GestureDetector{OnTap: onTap, Child: child}
}

// Drag wraps a child with pan (omnidirectional) drag handlers.
func Drag(onUpdate func(DragUpdateDetails), child core.Widget) GestureDetector {
	return GestureDetector{OnPanUpdate: onUpdate, Child: child}
}

// HorizontalDrag wraps a child with horizontal-only drag handlers.
func HorizontalDrag(onUpdate func(DragUpdateDetails), child core.Widget) GestureDetector {
	return GestureDetector{OnHorizontalDragUpdate: onUpdate, Child: child}
}

// VerticalDrag wraps a child with vertical-only drag handlers.
func VerticalDrag(onUpdate func(DragUpdateDetails), child core.Widget) GestureDetector {
	return GestureDetector{OnVerticalDragUpdate: onUpdate, Child: child}
}

// Clamp constrains a value between min and max bounds.
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// PaddingAll wraps a child with uniform padding on all sides.
func PaddingAll(value float64, child core.Widget) Padding {
	return Padding{Padding: layout.EdgeInsetsAll(value), Child: child}
}

// PaddingSym wraps a child with symmetric horizontal and vertical padding.
func PaddingSym(horizontal, vertical float64, child core.Widget) Padding {
	return Padding{Padding: layout.EdgeInsetsSymmetric(horizontal, vertical), Child: child}
}

// PaddingOnly wraps a child with specific padding on each side.
func PaddingOnly(left, top, right, bottom float64, child core.Widget) Padding {
	return Padding{Padding: layout.EdgeInsetsOnly(left, top, right, bottom), Child: child}
}

// Spacer fills remaining space along the main axis of a [Row] or [Column].
// It is equivalent to Expanded{Child: SizedBox{}}.
func Spacer() Expanded {
	return Expanded{Child: SizedBox{}}
}
