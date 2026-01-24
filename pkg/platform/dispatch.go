package platform

import "sync"

var (
	dispatchMu   sync.RWMutex
	dispatchFunc func(callback func())
)

// RegisterDispatch sets the dispatch function used to schedule callbacks on the UI thread.
// This should be called once by the engine during initialization.
func RegisterDispatch(fn func(callback func())) {
	dispatchMu.Lock()
	dispatchFunc = fn
	dispatchMu.Unlock()
}

// Dispatch schedules a callback to run on the UI thread.
// Returns true if the callback was successfully scheduled, false if no dispatch function
// is registered or the callback is nil.
func Dispatch(callback func()) bool {
	dispatchMu.RLock()
	fn := dispatchFunc
	dispatchMu.RUnlock()
	if fn == nil || callback == nil {
		return false
	}
	fn(callback)
	return true
}
