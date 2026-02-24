package navigation

import (
	"github.com/go-drift/drift/pkg/animation"
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/semantics"
)

// backgroundParallaxFactor controls how far the background page shifts left
// during a foreground push/pop transition (fraction of page width).
const backgroundParallaxFactor = 0.33

// SlideDirection determines the direction of a slide transition.
type SlideDirection int

const (
	// SlideFromRight slides content in from the right.
	SlideFromRight SlideDirection = iota
	// SlideFromLeft slides content in from the left.
	SlideFromLeft
	// SlideFromBottom slides content in from the bottom.
	SlideFromBottom
	// SlideFromTop slides content in from the top.
	SlideFromTop
)

// SlideTransition animates a child sliding from a direction.
type SlideTransition struct {
	core.RenderObjectBase
	Animation *animation.AnimationController
	Direction SlideDirection
	Child     core.Widget
}

// ChildWidget returns the child widget.
func (s SlideTransition) ChildWidget() core.Widget {
	return s.Child
}

// CreateRenderObject creates the RenderSlideTransition.
func (s SlideTransition) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	slide := &renderSlideTransition{
		transitionRenderBase: transitionRenderBase{animation: s.Animation},
		direction:            s.Direction,
	}
	slide.SetSelf(slide)
	slide.subscribeAnimation()
	return slide
}

// UpdateRenderObject updates the RenderSlideTransition.
func (s SlideTransition) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if slide, ok := renderObject.(*renderSlideTransition); ok {
		if slide.animation != s.Animation {
			slide.unsubscribeAnimation()
			slide.animation = s.Animation
			slide.subscribeAnimation()
		}
		slide.direction = s.Direction
		slide.MarkNeedsPaint()
	}
}

// transitionRenderBase provides shared functionality for all transition render types.
type transitionRenderBase struct {
	layout.RenderBoxBase
	child       layout.RenderBox
	animation   *animation.AnimationController
	unsubscribe func()
}

func (r *transitionRenderBase) subscribeAnimation() {
	if r.animation != nil {
		r.unsubscribe = r.animation.AddListener(func() {
			r.MarkNeedsPaint()
		})
	}
}

func (r *transitionRenderBase) unsubscribeAnimation() {
	if r.unsubscribe != nil {
		r.unsubscribe()
		r.unsubscribe = nil
	}
}

func (r *transitionRenderBase) SetChild(child layout.RenderObject) {
	layout.SetParentOnChild(r.child, nil)
	if child == nil {
		r.child = nil
		return
	}
	if box, ok := child.(layout.RenderBox); ok {
		r.child = box
		layout.SetParentOnChild(r.child, r.Self())
	}
}

func (r *transitionRenderBase) VisitChildren(visitor func(layout.RenderObject)) {
	if r.child != nil {
		visitor(r.child)
	}
}

// DescribeSemanticsConfiguration makes the transition act as a semantic container.
// This ensures all page content is grouped under one node for accessibility navigation.
func (r *transitionRenderBase) DescribeSemanticsConfiguration(config *semantics.SemanticsConfiguration) bool {
	config.IsSemanticBoundary = true
	return true
}

func (r *transitionRenderBase) PerformLayout() {
	constraints := r.Constraints()
	if r.child != nil {
		r.child.Layout(constraints, true) // true: we read child.Size()
		r.SetSize(r.child.Size())
		r.child.SetParentData(&layout.BoxParentData{})
	} else {
		r.SetSize(constraints.Constrain(graphics.Size{}))
	}
}

func (r *transitionRenderBase) HitTest(position graphics.Offset, result *layout.HitTestResult) bool {
	if r.child == nil {
		return false
	}
	return r.child.HitTest(position, result)
}

func (r *transitionRenderBase) Dispose() {
	r.unsubscribeAnimation()
	r.RenderBoxBase.Dispose()
}

type renderSlideTransition struct {
	transitionRenderBase
	direction SlideDirection
}

func (r *renderSlideTransition) slideOffset() graphics.Offset {
	offset := graphics.Offset{}
	if r.animation != nil {
		// Calculate offset based on animation value and direction
		// value 0 = off screen, value 1 = on screen
		t := 1.0 - r.animation.Value // Invert so 0 = visible, 1 = off screen
		size := r.Size()

		switch r.direction {
		case SlideFromRight:
			offset.X = size.Width * t
		case SlideFromLeft:
			offset.X = -size.Width * t
		case SlideFromBottom:
			offset.Y = size.Height * t
		case SlideFromTop:
			offset.Y = -size.Height * t
		}
	}
	return offset
}

func (r *renderSlideTransition) ScrollOffset() graphics.Offset {
	return r.slideOffset()
}

func (r *renderSlideTransition) Paint(ctx *layout.PaintContext) {
	if r.child == nil {
		return
	}

	offset := r.slideOffset()
	ctx.PaintChildWithLayer(r.child, offset)
}

// BackgroundSlideTransition slides its child to the left as a foreground page
// enters. At animation value 0 the child is at its normal position; at value 1
// the child is shifted left by 33% of the width.
type BackgroundSlideTransition struct {
	core.RenderObjectBase
	Animation *animation.AnimationController
	Child     core.Widget
}

// ChildWidget returns the child widget.
func (b BackgroundSlideTransition) ChildWidget() core.Widget {
	return b.Child
}

// CreateRenderObject creates the renderBackgroundSlideTransition.
func (b BackgroundSlideTransition) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	r := &renderBackgroundSlideTransition{
		transitionRenderBase: transitionRenderBase{animation: b.Animation},
	}
	r.SetSelf(r)
	r.subscribeAnimation()
	return r
}

// UpdateRenderObject updates the renderBackgroundSlideTransition.
func (b BackgroundSlideTransition) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if r, ok := renderObject.(*renderBackgroundSlideTransition); ok {
		if r.animation != b.Animation {
			r.unsubscribeAnimation()
			r.animation = b.Animation
			r.subscribeAnimation()
		}
		r.MarkNeedsPaint()
	}
}

type renderBackgroundSlideTransition struct {
	transitionRenderBase
}

func (r *renderBackgroundSlideTransition) slideOffset() graphics.Offset {
	if r.animation == nil {
		return graphics.Offset{}
	}
	return graphics.Offset{
		X: -r.Size().Width * backgroundParallaxFactor * r.animation.Value,
	}
}

func (r *renderBackgroundSlideTransition) ScrollOffset() graphics.Offset {
	return r.slideOffset()
}

func (r *renderBackgroundSlideTransition) Paint(ctx *layout.PaintContext) {
	if r.child == nil {
		return
	}
	offset := r.slideOffset()
	ctx.PaintChildWithLayer(r.child, offset)
}

// FadeTransition animates the opacity of its child.
type FadeTransition struct {
	core.RenderObjectBase
	Animation *animation.AnimationController
	Child     core.Widget
}

// ChildWidget returns the child widget.
func (f FadeTransition) ChildWidget() core.Widget {
	return f.Child
}

// CreateRenderObject creates the RenderFadeTransition.
func (f FadeTransition) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	fade := &renderFadeTransition{
		transitionRenderBase: transitionRenderBase{animation: f.Animation},
	}
	fade.SetSelf(fade)
	fade.subscribeAnimation()
	return fade
}

// UpdateRenderObject updates the RenderFadeTransition.
func (f FadeTransition) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if fade, ok := renderObject.(*renderFadeTransition); ok {
		if fade.animation != f.Animation {
			fade.unsubscribeAnimation()
			fade.animation = f.Animation
			fade.subscribeAnimation()
		}
		fade.MarkNeedsPaint()
	}
}

type renderFadeTransition struct {
	transitionRenderBase
}

func (r *renderFadeTransition) Paint(ctx *layout.PaintContext) {
	if r.child == nil {
		return
	}
	opacity := 1.0
	if r.animation != nil {
		opacity = r.animation.Value
	}
	if opacity <= 0 {
		return
	}
	if opacity >= 1 {
		ctx.PaintChildWithLayer(r.child, graphics.Offset{})
		return
	}
	size := r.Size()
	bounds := graphics.RectFromLTWH(0, 0, size.Width, size.Height)
	ctx.Canvas.SaveLayerAlpha(bounds, opacity)
	ctx.PaintChildWithLayer(r.child, graphics.Offset{})
	ctx.Canvas.Restore()
}
