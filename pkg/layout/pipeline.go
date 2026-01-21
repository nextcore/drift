package layout

// PipelineOwner tracks render objects that need layout or paint.
type PipelineOwner struct {
	dirtyLayout map[RenderObject]struct{}
	dirtyPaint  map[RenderObject]struct{}
	needsLayout bool
	needsPaint  bool
}

// ScheduleLayout marks a render object as needing layout.
func (p *PipelineOwner) ScheduleLayout(object RenderObject) {
	if p.dirtyLayout == nil {
		p.dirtyLayout = make(map[RenderObject]struct{})
	}
	if _, exists := p.dirtyLayout[object]; exists {
		return
	}
	p.dirtyLayout[object] = struct{}{}
	p.needsLayout = true
	p.needsPaint = true
}

// SchedulePaint marks a render object as needing paint.
func (p *PipelineOwner) SchedulePaint(object RenderObject) {
	if p.dirtyPaint == nil {
		p.dirtyPaint = make(map[RenderObject]struct{})
	}
	if _, exists := p.dirtyPaint[object]; exists {
		return
	}
	p.dirtyPaint[object] = struct{}{}
	p.needsPaint = true
}

// NeedsLayout reports if any render objects need layout.
func (p *PipelineOwner) NeedsLayout() bool {
	return p.needsLayout
}

// NeedsPaint reports if any render objects need paint.
func (p *PipelineOwner) NeedsPaint() bool {
	return p.needsPaint
}

// FlushLayoutForRoot runs layout from the root when any object is dirty.
func (p *PipelineOwner) FlushLayoutForRoot(root RenderObject, constraints Constraints) {
	if !p.needsLayout || root == nil {
		return
	}
	root.Layout(constraints)
	p.dirtyLayout = nil
	p.needsLayout = false
}

// FlushPaint clears the dirty paint list.
func (p *PipelineOwner) FlushPaint() {
	p.dirtyPaint = nil
	p.needsPaint = false
}

// FlushLayout clears the dirty layout list without performing layout.
func (p *PipelineOwner) FlushLayout() {
	p.dirtyLayout = nil
	p.needsLayout = false
}
