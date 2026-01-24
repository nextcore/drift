package widgets

import (
	"reflect"
	"sync"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/platform"
)

// SafeAreaData provides safe area insets to descendants via InheritedWidget.
type SafeAreaData struct {
	Insets      layout.EdgeInsets
	ChildWidget core.Widget
}

func (s SafeAreaData) CreateElement() core.Element {
	return core.NewInheritedElement(s, nil)
}

func (s SafeAreaData) Key() any {
	return nil
}

func (s SafeAreaData) Child() core.Widget {
	return s.ChildWidget
}

func (s SafeAreaData) UpdateShouldNotify(oldWidget core.InheritedWidget) bool {
	if old, ok := oldWidget.(SafeAreaData); ok {
		return s.Insets != old.Insets
	}
	return true
}

// SafeAreaProvider is a StatefulWidget that subscribes to platform safe area changes
// and provides SafeAreaData to descendants. This scopes rebuilds to only the provider
// and widgets that depend on safe area data, instead of rebuilding the entire tree.
type SafeAreaProvider struct {
	ChildWidget core.Widget
}

func (s SafeAreaProvider) CreateElement() core.Element {
	return core.NewStatefulElement(s, nil)
}

func (s SafeAreaProvider) Key() any {
	return nil
}

func (s SafeAreaProvider) CreateState() core.State {
	return &safeAreaProviderState{}
}

type safeAreaProviderState struct {
	core.StateBase
	insets      layout.EdgeInsets
	unsubscribe func()
	mu          sync.Mutex
	pending     layout.EdgeInsets
	hasPending  bool
}

func (s *safeAreaProviderState) InitState() {
	// Read initial insets
	platformInsets := platform.SafeArea.Insets()
	s.insets = layout.EdgeInsets{
		Top:    platformInsets.Top,
		Bottom: platformInsets.Bottom,
		Left:   platformInsets.Left,
		Right:  platformInsets.Right,
	}

	// Subscribe to changes
	s.unsubscribe = platform.SafeArea.AddHandler(s.onPlatformInsetsChanged)
	s.OnDispose(func() {
		if s.unsubscribe != nil {
			s.unsubscribe()
		}
	})
}

func (s *safeAreaProviderState) onPlatformInsetsChanged(insets platform.EdgeInsets) {
	newInsets := layout.EdgeInsets{
		Top:    insets.Top,
		Bottom: insets.Bottom,
		Left:   insets.Left,
		Right:  insets.Right,
	}

	// Batch rapid updates
	s.mu.Lock()
	s.pending = newInsets
	shouldSchedule := !s.hasPending
	s.hasPending = true
	s.mu.Unlock()

	if shouldSchedule {
		if !platform.Dispatch(s.applyPendingInsets) {
			// Dispatch not available - clear hasPending so future updates can retry
			s.mu.Lock()
			s.hasPending = false
			s.mu.Unlock()
		}
	}
}

func (s *safeAreaProviderState) applyPendingInsets() {
	s.mu.Lock()
	newInsets := s.pending
	s.hasPending = false
	s.mu.Unlock()

	if s.insets == newInsets {
		return
	}
	s.SetState(func() { s.insets = newInsets })
}

func (s *safeAreaProviderState) Build(ctx core.BuildContext) core.Widget {
	w := s.Element().Widget().(SafeAreaProvider)
	return SafeAreaData{
		Insets:      s.insets,
		ChildWidget: w.ChildWidget,
	}
}

var safeAreaDataType = reflect.TypeOf(SafeAreaData{})

// SafeAreaOf returns the current safe area insets from context.
func SafeAreaOf(ctx core.BuildContext) layout.EdgeInsets {
	inherited := ctx.DependOnInherited(safeAreaDataType)
	if sa, ok := inherited.(SafeAreaData); ok {
		return sa.Insets
	}
	return layout.EdgeInsets{}
}

// SafeAreaPadding returns the safe area insets as EdgeInsets for use with
// ScrollView.Padding or other widgets. The returned EdgeInsets can be modified
// using chainable methods:
//
//	ScrollView{
//	    Padding: widgets.SafeAreaPadding(ctx),              // just safe area
//	    ChildWidget: ...,
//	}
//	ScrollView{
//	    Padding: widgets.SafeAreaPadding(ctx).Add(24),      // safe area + 24px all sides
//	    ChildWidget: ...,
//	}
//	ScrollView{
//	    Padding: widgets.SafeAreaPadding(ctx).OnlyTop().Add(24), // only top safe area + 24px
//	    ChildWidget: ...,
//	}
func SafeAreaPadding(ctx core.BuildContext) layout.EdgeInsets {
	return SafeAreaOf(ctx)
}

// SafeArea is a convenience widget that applies safe area insets as padding.
type SafeArea struct {
	Top         bool
	Bottom      bool
	Left        bool
	Right       bool
	ChildWidget core.Widget
}

func (s SafeArea) CreateElement() core.Element {
	return core.NewStatelessElement(s, nil)
}

func (s SafeArea) Key() any {
	return nil
}

func (s SafeArea) Build(ctx core.BuildContext) core.Widget {
	insets := SafeAreaOf(ctx)

	// Default to applying all sides if none specified
	top, bottom, left, right := s.Top, s.Bottom, s.Left, s.Right
	if !top && !bottom && !left && !right {
		top, bottom, left, right = true, true, true, true
	}

	padding := layout.EdgeInsets{}
	if top {
		padding.Top = insets.Top
	}
	if bottom {
		padding.Bottom = insets.Bottom
	}
	if left {
		padding.Left = insets.Left
	}
	if right {
		padding.Right = insets.Right
	}

	return Padding{
		Padding:     padding,
		ChildWidget: s.ChildWidget,
	}
}
