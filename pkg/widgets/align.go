package widgets

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// Align positions its child within itself according to the given alignment.
//
// Align expands to fill available space, then positions the child within
// that space according to the Alignment field. The child is given loose
// constraints, allowing it to size itself.
//
// Example:
//
//	Align{
//	    Alignment:   layout.AlignmentBottomRight,
//	    ChildWidget: Text{Content: "Bottom right"},
//	}
//
// See also:
//   - [Center] for centering (equivalent to Align with AlignmentCenter)
//   - [Container] for combined alignment, padding, and decoration
type Align struct {
	ChildWidget core.Widget
	Alignment   layout.Alignment
}

func (a Align) CreateElement() core.Element {
	return core.NewRenderObjectElement(a, nil)
}

func (a Align) Key() any {
	return nil
}

func (a Align) Child() core.Widget {
	return a.ChildWidget
}

func (a Align) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	r := &renderAlign{alignment: a.Alignment}
	r.SetSelf(r)
	return r
}

func (a Align) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if r, ok := renderObject.(*renderAlign); ok {
		r.alignment = a.Alignment
		r.MarkNeedsLayout()
	}
}

type renderAlign struct {
	layout.RenderBoxBase
	child     layout.RenderBox
	alignment layout.Alignment
}

func (r *renderAlign) SetChild(child layout.RenderObject) {
	setParentOnChild(r.child, nil)
	r.child = setChildFromRenderObject(child)
	setParentOnChild(r.child, r)
}

func (r *renderAlign) VisitChildren(visitor func(layout.RenderObject)) {
	if r.child != nil {
		visitor(r.child)
	}
}

func (r *renderAlign) PerformLayout() {
	constraints := r.Constraints()
	size := constraints.Constrain(graphics.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight})
	r.SetSize(size)
	if r.child != nil {
		r.child.Layout(layout.Loose(size), true)
		childSize := r.child.Size()
		offset := r.alignment.WithinRect(
			graphics.RectFromLTWH(0, 0, size.Width, size.Height),
			childSize,
		)
		r.child.SetParentData(&layout.BoxParentData{Offset: offset})
	}
}

func (r *renderAlign) Paint(ctx *layout.PaintContext) {
	if r.child != nil {
		ctx.PaintChild(r.child, getChildOffset(r.child))
	}
}

func (r *renderAlign) HitTest(position graphics.Offset, result *layout.HitTestResult) bool {
	if !withinBounds(position, r.Size()) {
		return false
	}
	offset := getChildOffset(r.child)
	local := graphics.Offset{X: position.X - offset.X, Y: position.Y - offset.Y}
	if r.child != nil && r.child.HitTest(local, result) {
		return true
	}
	result.Add(r)
	return true
}
