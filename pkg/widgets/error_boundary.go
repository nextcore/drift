package widgets

import (
	"reflect"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/errors"
)

// ErrorBoundary catches build errors from descendant widgets and displays
// a fallback widget instead of crashing the app. This provides scoped error
// handling for subtrees of the widget tree.
//
// Example:
//
//	ErrorBoundary{
//	    OnError: func(err *errors.BuildError) {
//	        log.Printf("Widget error: %v", err)
//	    },
//	    FallbackBuilder: func(err *errors.BuildError) core.Widget {
//	        return widgets.Text{Content: "Failed to load"}
//	    },
//	    ChildWidget: RiskyContent{},
//	}
type ErrorBoundary struct {
	// ChildWidget is the widget tree to wrap with error handling.
	ChildWidget core.Widget
	// FallbackBuilder creates a widget to show when an error is caught.
	// If nil, uses the default ErrorWidget.
	FallbackBuilder core.ErrorWidgetBuilder
	// OnError is called when an error is caught. Use for logging/analytics.
	OnError func(*errors.BuildError)
	// WidgetKey is an optional key for the widget. Changing the key forces
	// the ErrorBoundary to recreate its state, clearing any captured error.
	WidgetKey any
}

func (e ErrorBoundary) CreateElement() core.Element {
	return core.NewStatefulElement(e, nil)
}

func (e ErrorBoundary) Key() any {
	return e.WidgetKey
}

func (e ErrorBoundary) CreateState() core.State {
	return &errorBoundaryState{}
}

type errorBoundaryState struct {
	core.StateBase
	capturedError *errors.BuildError
}

func (s *errorBoundaryState) Build(ctx core.BuildContext) core.Widget {
	widget := ctx.Widget().(ErrorBoundary)

	// If we've captured an error, show the fallback
	if s.capturedError != nil {
		if widget.FallbackBuilder != nil {
			return widget.FallbackBuilder(s.capturedError)
		}
		return ErrorWidget{Error: s.capturedError}
	}

	// Wrap child in an inherited widget that marks this boundary
	return errorBoundaryScope{
		state:       s,
		childWidget: widget.ChildWidget,
	}
}

// Reset clears the captured error and rebuilds the child.
// Use this to retry rendering after an error.
func (s *errorBoundaryState) Reset() {
	s.SetState(func() {
		s.capturedError = nil
	})
}

// HasError returns true if this boundary has captured an error.
func (s *errorBoundaryState) HasError() bool {
	return s.capturedError != nil
}

// Error returns the captured error, or nil if none.
func (s *errorBoundaryState) Error() *errors.BuildError {
	return s.capturedError
}

// errorBoundaryScope is an InheritedWidget that marks the boundary in the tree.
// Child elements can find this to report errors to the boundary.
type errorBoundaryScope struct {
	state       *errorBoundaryState
	childWidget core.Widget
}

func (e errorBoundaryScope) CreateElement() core.Element {
	return newErrorBoundaryScopeElement(e)
}

func (e errorBoundaryScope) Key() any {
	return nil
}

func (e errorBoundaryScope) Child() core.Widget {
	return e.childWidget
}

func (e errorBoundaryScope) UpdateShouldNotify(oldWidget core.InheritedWidget) bool {
	return false // Scope widget never notifies - it's just for tree marking
}

func (e errorBoundaryScope) UpdateShouldNotifyDependent(oldWidget core.InheritedWidget, aspects map[any]struct{}) bool {
	return false
}

// errorBoundaryScopeElement wraps InheritedElement and implements ErrorBoundaryCapture.
// We need to override Mount and RebuildIfNeeded to ensure 'self' is passed as the parent
// so that findErrorBoundary can find this element.
type errorBoundaryScopeElement struct {
	*core.InheritedElement
	self core.Element // Keep reference to self for parent passing
}

func newErrorBoundaryScopeElement(widget errorBoundaryScope) *errorBoundaryScopeElement {
	inherited := core.NewInheritedElement(widget, nil)
	element := &errorBoundaryScopeElement{
		InheritedElement: inherited,
	}
	element.self = element // Set self reference
	return element
}

// Mount overrides InheritedElement.Mount to pass self as parent to children.
func (e *errorBoundaryScopeElement) Mount(parent core.Element, slot any) {
	e.InheritedElement.MountWithSelf(parent, slot, e.self)
}

// RebuildIfNeeded overrides to pass self as parent when updating children.
func (e *errorBoundaryScopeElement) RebuildIfNeeded() {
	e.InheritedElement.RebuildIfNeededWithSelf(e.self)
}

// CaptureError implements core.ErrorBoundaryCapture.
func (e *errorBoundaryScopeElement) CaptureError(err *errors.BuildError) bool {
	widget := e.Widget().(errorBoundaryScope)
	state := widget.state

	// Get the parent ErrorBoundary widget for OnError callback
	if state.Element() != nil {
		parentWidget := state.Element().Widget()
		if boundary, ok := parentWidget.(ErrorBoundary); ok && boundary.OnError != nil {
			boundary.OnError(err)
		}
	}

	// Capture the error and trigger rebuild
	state.SetState(func() {
		state.capturedError = err
	})

	return true
}

// ErrorBoundaryOf returns the nearest ErrorBoundary's state, or nil if none.
// This can be used to programmatically reset the boundary or check for errors.
func ErrorBoundaryOf(ctx core.BuildContext) *errorBoundaryState {
	inherited := ctx.DependOnInherited(reflect.TypeOf(errorBoundaryScope{}), nil)
	if inherited == nil {
		return nil
	}
	if scope, ok := inherited.(errorBoundaryScope); ok {
		return scope.state
	}
	return nil
}
