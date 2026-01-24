package core

import (
	"testing"

	"github.com/go-drift/drift/pkg/errors"
)

// testStatelessWidget is a simple stateless widget for testing.
type testStatelessWidget struct {
	buildFn func(BuildContext) Widget
}

func (w testStatelessWidget) CreateElement() Element {
	return NewStatelessElement(w, nil)
}

func (w testStatelessWidget) Key() any {
	return nil
}

func (w testStatelessWidget) Build(ctx BuildContext) Widget {
	if w.buildFn != nil {
		return w.buildFn(ctx)
	}
	return nil
}

// testStatefulWidget is a simple stateful widget for testing.
type testStatefulWidget struct {
	createStateFn func() State
}

func (w testStatefulWidget) CreateElement() Element {
	return NewStatefulElement(w, nil)
}

func (w testStatefulWidget) Key() any {
	return nil
}

func (w testStatefulWidget) CreateState() State {
	if w.createStateFn != nil {
		return w.createStateFn()
	}
	return &testState{}
}

type testState struct {
	StateBase
	buildFn func(BuildContext) Widget
}

func (s *testState) Build(ctx BuildContext) Widget {
	if s.buildFn != nil {
		return s.buildFn(ctx)
	}
	return nil
}

// testErrorHandler captures errors for testing.
type testErrorHandler struct {
	errors.LogHandler
	buildErrors []*errors.BuildError
}

func (h *testErrorHandler) HandleBuildError(err *errors.BuildError) {
	h.buildErrors = append(h.buildErrors, err)
}

func TestStatelessElement_BuildPanic_ReportsError(t *testing.T) {
	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	widget := testStatelessWidget{
		buildFn: func(ctx BuildContext) Widget {
			panic("test panic in stateless build")
		},
	}

	owner := NewBuildOwner()
	element := NewStatelessElement(widget, owner)
	element.Mount(nil, nil)

	if len(handler.buildErrors) != 1 {
		t.Fatalf("expected 1 build error, got %d", len(handler.buildErrors))
	}

	err := handler.buildErrors[0]
	if err.Recovered != "test panic in stateless build" {
		t.Errorf("expected panic value 'test panic in stateless build', got %v", err.Recovered)
	}
	if err.Widget == "" {
		t.Error("expected Widget type to be set")
	}
	if err.Element == "" {
		t.Error("expected Element type to be set")
	}
	if err.StackTrace == "" {
		t.Error("expected StackTrace to be captured")
	}
}

func TestStatefulElement_BuildPanic_ReportsError(t *testing.T) {
	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	widget := testStatefulWidget{
		createStateFn: func() State {
			return &testState{
				buildFn: func(ctx BuildContext) Widget {
					panic("test panic in stateful build")
				},
			}
		},
	}

	owner := NewBuildOwner()
	element := NewStatefulElement(widget, owner)
	element.Mount(nil, nil)

	if len(handler.buildErrors) != 1 {
		t.Fatalf("expected 1 build error, got %d", len(handler.buildErrors))
	}

	err := handler.buildErrors[0]
	if err.Recovered != "test panic in stateful build" {
		t.Errorf("expected panic value 'test panic in stateful build', got %v", err.Recovered)
	}
}

func TestSafeBuild_ReturnsErrorPlaceholder_WhenNoBuilder(t *testing.T) {
	// Temporarily clear the error widget builder
	oldBuilder := GetErrorWidgetBuilder()
	SetErrorWidgetBuilder(func(err *errors.BuildError) Widget {
		return nil // Force fallback to errorPlaceholder
	})
	defer SetErrorWidgetBuilder(oldBuilder)

	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	widget := testStatelessWidget{
		buildFn: func(ctx BuildContext) Widget {
			panic("test panic")
		},
	}

	owner := NewBuildOwner()
	element := NewStatelessElement(widget, owner)
	element.Mount(nil, nil)

	// The child should be an errorPlaceholder
	if element.child == nil {
		t.Fatal("expected child element to be set")
	}

	childWidget := element.child.Widget()
	if _, ok := childWidget.(errorPlaceholder); !ok {
		t.Errorf("expected errorPlaceholder widget, got %T", childWidget)
	}
}

func TestSafeBuild_UsesCustomBuilder(t *testing.T) {
	var capturedErr *errors.BuildError
	customWidget := testStatelessWidget{
		buildFn: func(ctx BuildContext) Widget {
			return nil
		},
	}

	SetErrorWidgetBuilder(func(err *errors.BuildError) Widget {
		capturedErr = err
		return customWidget
	})
	defer SetErrorWidgetBuilder(nil)

	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	widget := testStatelessWidget{
		buildFn: func(ctx BuildContext) Widget {
			panic("custom builder test")
		},
	}

	owner := NewBuildOwner()
	element := NewStatelessElement(widget, owner)
	element.Mount(nil, nil)

	if capturedErr == nil {
		t.Fatal("expected custom builder to be called")
	}
	if capturedErr.Recovered != "custom builder test" {
		t.Errorf("expected panic value 'custom builder test', got %v", capturedErr.Recovered)
	}
}

func TestErrorPlaceholder_BuildReturnsNil(t *testing.T) {
	placeholder := errorPlaceholder{
		err: &errors.BuildError{Widget: "test"},
	}

	built := placeholder.Build(nil)
	if built != nil {
		t.Errorf("expected errorPlaceholder.Build() to return nil, got %v", built)
	}
}

func TestSetErrorWidgetBuilder_NilRestoresDefault(t *testing.T) {
	SetErrorWidgetBuilder(func(err *errors.BuildError) Widget {
		return testStatelessWidget{}
	})

	// Restore default
	SetErrorWidgetBuilder(nil)

	builder := GetErrorWidgetBuilder()
	if builder == nil {
		t.Fatal("expected non-nil builder after SetErrorWidgetBuilder(nil)")
	}

	// Default builder returns nil
	result := builder(&errors.BuildError{})
	if result != nil {
		t.Errorf("expected default builder to return nil, got %v", result)
	}
}

func TestDebugMode_Default(t *testing.T) {
	if !DebugMode {
		t.Error("expected DebugMode to default to true")
	}
}

func TestSetDebugMode(t *testing.T) {
	original := DebugMode

	SetDebugMode(false)
	if DebugMode {
		t.Error("expected DebugMode to be false")
	}

	SetDebugMode(true)
	if !DebugMode {
		t.Error("expected DebugMode to be true")
	}

	// Restore original
	SetDebugMode(original)
}

func TestStatelessElement_NormalBuild_NoError(t *testing.T) {
	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	buildCalled := false
	widget := testStatelessWidget{
		buildFn: func(ctx BuildContext) Widget {
			buildCalled = true
			return nil
		},
	}

	owner := NewBuildOwner()
	element := NewStatelessElement(widget, owner)
	element.Mount(nil, nil)

	if !buildCalled {
		t.Error("expected build to be called")
	}
	if len(handler.buildErrors) != 0 {
		t.Errorf("expected no build errors, got %d", len(handler.buildErrors))
	}
}

func TestStatefulElement_NormalBuild_NoError(t *testing.T) {
	handler := &testErrorHandler{}
	errors.SetHandler(handler)
	defer errors.SetHandler(nil)

	buildCalled := false
	widget := testStatefulWidget{
		createStateFn: func() State {
			return &testState{
				buildFn: func(ctx BuildContext) Widget {
					buildCalled = true
					return nil
				},
			}
		},
	}

	owner := NewBuildOwner()
	element := NewStatefulElement(widget, owner)
	element.Mount(nil, nil)

	if !buildCalled {
		t.Error("expected build to be called")
	}
	if len(handler.buildErrors) != 0 {
		t.Errorf("expected no build errors, got %d", len(handler.buildErrors))
	}
}
