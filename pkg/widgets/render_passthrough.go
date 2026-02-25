package widgets

import (
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// renderPassthrough provides shared passthrough render methods for single-child
// render objects that simply delegate layout, paint, and hit testing to their child.
type renderPassthrough struct {
	layout.RenderBoxBase
	child layout.RenderObject
}

func (r *renderPassthrough) SetChild(child layout.RenderObject) {
	layout.SetParentOnChild(r.child, nil)
	r.child = child
	layout.SetParentOnChild(r.child, r.Self())
}

func (r *renderPassthrough) VisitChildren(visitor func(layout.RenderObject)) {
	if r.child != nil {
		visitor(r.child)
	}
}

func (r *renderPassthrough) PerformLayout() {
	constraints := r.Constraints()
	if r.child != nil {
		r.child.Layout(constraints, true)
		r.SetSize(r.child.Size())
	} else {
		r.SetSize(constraints.Constrain(graphics.Size{}))
	}
}

func (r *renderPassthrough) Paint(ctx *layout.PaintContext) {
	if r.child != nil {
		ctx.PaintChildWithLayer(r.child.(layout.RenderBox), graphics.Offset{})
	}
}

func (r *renderPassthrough) HitTest(position graphics.Offset, result *layout.HitTestResult) bool {
	if r.child != nil {
		return r.child.HitTest(position, result)
	}
	return false
}
