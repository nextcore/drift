package core

import (
	"sync"

	"github.com/go-drift/drift/pkg/errors"
)

// ErrorWidgetBuilder creates a fallback widget when a widget build fails.
// The builder receives the build error and should return a widget to display
// in place of the failed widget.
type ErrorWidgetBuilder func(err *errors.BuildError) Widget

var (
	errorWidgetBuilder ErrorWidgetBuilder = DefaultErrorWidgetBuilder
	errorBuilderMu     sync.RWMutex
)

// SetErrorWidgetBuilder configures the global error widget builder.
// Pass nil to restore the default builder.
func SetErrorWidgetBuilder(builder ErrorWidgetBuilder) {
	errorBuilderMu.Lock()
	defer errorBuilderMu.Unlock()
	if builder == nil {
		errorWidgetBuilder = DefaultErrorWidgetBuilder
	} else {
		errorWidgetBuilder = builder
	}
}

// GetErrorWidgetBuilder returns the current error widget builder.
func GetErrorWidgetBuilder() ErrorWidgetBuilder {
	errorBuilderMu.RLock()
	defer errorBuilderMu.RUnlock()
	return errorWidgetBuilder
}

// DefaultErrorWidgetBuilder returns a placeholder widget when build fails.
// The actual error widget implementation is in pkg/widgets to avoid
// circular dependencies. This default returns nil, which signals
// the framework to use the widgets.ErrorWidget if available.
func DefaultErrorWidgetBuilder(err *errors.BuildError) Widget {
	// Return nil to signal that the default ErrorWidget should be used.
	// The element.safeBuild() will check for nil and use a minimal
	// fallback if the widgets package hasn't registered a better default.
	return nil
}

// ErrorBoundaryCapture is implemented by error boundary elements to capture
// build errors from descendant widgets.
type ErrorBoundaryCapture interface {
	// CaptureError captures a build error from a descendant widget.
	// Returns true if the error was captured and handled.
	CaptureError(err *errors.BuildError) bool
}
