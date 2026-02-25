//go:build !android && !darwin && !ios

package widgets

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/semantics"
)

// Semantics is a no-op widget on non-mobile platforms.
// It just passes through to its child.
type Semantics struct {
	core.RenderObjectBase
	Child                core.Widget
	Label                string
	Value                string
	Hint                 string
	Tooltip              string
	Role                 semantics.SemanticsRole
	Flags                semantics.SemanticsFlag
	Container            bool
	MergeDescendants     bool
	ExplicitChildNodes   bool
	CurrentValue         *float64
	MinValue             *float64
	MaxValue             *float64
	HeadingLevel         int
	OnTap                func()
	OnLongPress          func()
	OnScrollLeft         func()
	OnScrollRight        func()
	OnScrollUp           func()
	OnScrollDown         func()
	OnIncrease           func()
	OnDecrease           func()
	OnDismiss            func()
	CustomActions        []semantics.CustomSemanticsAction
	CustomActionHandlers map[int64]func()
}

func (s Semantics) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	r := &renderSemanticsStub{}
	r.SetSelf(r)
	return r
}

func (s Semantics) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
}

func (s Semantics) ChildWidget() core.Widget {
	return s.Child
}

type renderSemanticsStub struct {
	renderPassthrough
}

// MergeSemantics is a no-op widget on non-mobile platforms.
type MergeSemantics struct {
	core.RenderObjectBase
	Child core.Widget
}

func (m MergeSemantics) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	r := &renderMergeSemanticsStub{}
	r.SetSelf(r)
	return r
}

func (m MergeSemantics) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
}

func (m MergeSemantics) ChildWidget() core.Widget {
	return m.Child
}

type renderMergeSemanticsStub struct {
	renderPassthrough
}

// ExcludeSemantics is a no-op on non-mobile platforms.
type ExcludeSemantics struct {
	core.RenderObjectBase
	Child     core.Widget
	Excluding bool
}

// NewExcludeSemantics creates an ExcludeSemantics widget (no-op on non-mobile platforms).
func NewExcludeSemantics(child core.Widget) ExcludeSemantics {
	return ExcludeSemantics{Excluding: true, Child: child}
}

func (e ExcludeSemantics) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	r := &renderExcludeSemanticsStub{}
	r.SetSelf(r)
	return r
}

func (e ExcludeSemantics) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
}

func (e ExcludeSemantics) ChildWidget() core.Widget {
	return e.Child
}

type renderExcludeSemanticsStub struct {
	renderPassthrough
}
