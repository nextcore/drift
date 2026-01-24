package layout

import (
	"github.com/go-drift/drift/pkg/gestures"
	"github.com/go-drift/drift/pkg/rendering"
)

// HitTestResult collects hit test entries in paint order.
type HitTestResult struct {
	Entries []RenderObject
}

// Add inserts a render object into the hit test result list.
func (h *HitTestResult) Add(target RenderObject) {
	h.Entries = append(h.Entries, target)
}

// TapTarget is a render object that responds to tap events.
type TapTarget interface {
	OnTap()
}

// PointerHandler receives pointer events routed from hit testing.
type PointerHandler interface {
	HandlePointer(event gestures.PointerEvent)
}

// PaintContext provides the canvas for painting render objects.
type PaintContext struct {
	Canvas rendering.Canvas
}

// PaintChild paints a child render box at the given offset.
func (p *PaintContext) PaintChild(child RenderBox, offset rendering.Offset) {
	if child == nil {
		return
	}
	p.Canvas.Save()
	p.Canvas.Translate(offset.X, offset.Y)
	child.Paint(p)
	p.Canvas.Restore()
}

// PaintChildWithLayer paints a child, using its cached layer if available.
func (p *PaintContext) PaintChildWithLayer(child RenderBox, offset rendering.Offset) {
	if child == nil {
		return
	}

	p.Canvas.Save()
	p.Canvas.Translate(offset.X, offset.Y)

	// Use cached layer if child is a repaint boundary with valid cache
	if boundary, ok := child.(interface {
		IsRepaintBoundary() bool
		Layer() *rendering.DisplayList
		NeedsPaint() bool
	}); ok && boundary.IsRepaintBoundary() {
		if layer := boundary.Layer(); layer != nil && !boundary.NeedsPaint() {
			layer.Paint(p.Canvas)
			p.Canvas.Restore()
			return
		}
	}

	child.Paint(p)
	p.Canvas.Restore()
}
