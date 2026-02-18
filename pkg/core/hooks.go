package core

// UseController creates a controller and registers it for automatic disposal.
// The controller will be disposed when the state is disposed.
//
// Example:
//
//	func (s *myState) InitState() {
//	    s.animation = core.UseController(s, func() *animation.AnimationController {
//	        return animation.NewAnimationController(300 * time.Millisecond)
//	    })
//	}
func UseController[C Disposable](s stateBase, create func() C) C {
	base := s.state()
	controller := create()
	base.OnDispose(func() {
		controller.Dispose()
	})
	return controller
}

// UseListenable subscribes to a listenable and triggers rebuilds.
// The subscription is automatically cleaned up when the state is disposed.
//
// Example:
//
//	func (s *myState) InitState() {
//	    s.controller = core.UseController(s, func() *MyController {
//	        return NewMyController()
//	    })
//	    core.UseListenable(s, s.controller)
//	}
func UseListenable(s stateBase, listenable Listenable) {
	base := s.state()
	unsub := listenable.AddListener(func() {
		base.SetState(nil)
	})
	base.OnDispose(unsub)
}

// UseObservable subscribes to an observable and triggers rebuilds when it changes.
// Call this once in InitState(), not in Build(). The subscription is automatically
// cleaned up when the state is disposed.
//
// Example:
//
//	func (s *myState) InitState() {
//	    s.counter = core.NewObservable(0)
//	    core.UseObservable(s, s.counter)
//	}
//
//	func (s *myState) Build(ctx core.BuildContext) core.Widget {
//	    // Use .Value() in Build to read the current value
//	    return widgets.Text{Content: fmt.Sprintf("Count: %d", s.counter.Value()), ...}
//	}
func UseObservable[T any](s stateBase, obs *Observable[T]) {
	base := s.state()
	unsub := obs.AddListener(func(T) {
		base.SetState(nil)
	})
	base.OnDispose(unsub)
}

// Managed holds a value and triggers rebuilds when it changes.
// Unlike Observable, it is tied to a specific StateBase.
//
// Managed is NOT thread-safe. It must only be accessed from the UI thread.
// To update from a background goroutine, use drift.Dispatch:
//
//	go func() {
//	    result := doExpensiveWork()
//	    drift.Dispatch(func() {
//	        s.data.Set(result)  // Safe - runs on UI thread
//	    })
//	}()
//
// Example:
//
//	type myState struct {
//	    core.StateBase
//	    count *core.Managed[int]
//	}
//
//	func (s *myState) InitState() {
//	    s.count = core.NewManaged(s, 0)
//	}
//
//	func (s *myState) Build(ctx core.BuildContext) core.Widget {
//	    return widgets.GestureDetector{
//	        OnTap: func() { s.count.Set(s.count.Value() + 1) },
//	        Child: widgets.Text{Content: fmt.Sprintf("Count: %d", s.count.Value()), ...},
//	    }
//	}
type Managed[T any] struct {
	base  *StateBase
	value T
}

// NewManaged creates a new managed state value.
// Changes to this value will automatically trigger a rebuild.
func NewManaged[T any](s stateBase, initial T) *Managed[T] {
	return &Managed[T]{
		base:  s.state(),
		value: initial,
	}
}

// Value returns the current value.
func (m *Managed[T]) Value() T {
	return m.value
}

// Set updates the value and triggers a rebuild.
func (m *Managed[T]) Set(value T) {
	m.value = value
	m.base.SetState(nil)
}

// Update applies a transformation to the current value and triggers a rebuild.
func (m *Managed[T]) Update(transform func(T) T) {
	m.value = transform(m.value)
	m.base.SetState(nil)
}
