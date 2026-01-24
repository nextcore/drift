package core

// DebugMode controls whether debug information is displayed in error widgets.
// When true, error widgets show detailed error messages and stack traces.
// When false, error widgets show minimal information.
var DebugMode = true

// SetDebugMode enables or disables debug mode for the framework.
func SetDebugMode(debug bool) {
	DebugMode = debug
}
