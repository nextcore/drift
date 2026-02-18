package core

import "testing"

// MockDisposable for testing UseController
type mockDisposable struct {
	disposed bool
}

func (m *mockDisposable) Dispose() {
	m.disposed = true
}

func TestUseController(t *testing.T) {
	base := &StateBase{}

	controller := UseController(base, func() *mockDisposable {
		return &mockDisposable{}
	})

	if controller.disposed {
		t.Error("Controller should not be disposed initially")
	}

	base.Dispose()

	if !controller.disposed {
		t.Error("Controller should be disposed when StateBase is disposed")
	}
}

func TestUseListenable(t *testing.T) {
	base := &StateBase{}
	notifier := NewNotifier()

	UseListenable(base, notifier)

	// We can't easily test SetState being called without a real element,
	// but we can verify the subscription is set up
	if notifier.ListenerCount() != 1 {
		t.Errorf("Expected 1 listener, got %d", notifier.ListenerCount())
	}

	base.Dispose()

	if notifier.ListenerCount() != 0 {
		t.Errorf("Expected 0 listeners after dispose, got %d", notifier.ListenerCount())
	}
}

func TestUseObservable(t *testing.T) {
	base := &StateBase{}
	obs := NewObservable(42)

	UseObservable(base, obs)

	// Verify we can read the value
	if obs.Value() != 42 {
		t.Errorf("Expected 42, got %d", obs.Value())
	}

	// Verify listener was registered (observable should trigger rebuild on change)
	obs.Set(100)
	// SetState was called (can't easily verify without element, but no panic = good)
}

func TestUseObservable_Cleanup(t *testing.T) {
	base := &StateBase{}
	obs := NewObservable(0)

	UseObservable(base, obs)

	base.Dispose()

	// After dispose, setting the observable should not panic
	obs.Set(999)
}

func TestManaged_Value(t *testing.T) {
	base := &StateBase{}
	state := NewManaged(base, 42)

	if state.Value() != 42 {
		t.Errorf("Expected 42, got %d", state.Value())
	}
}

func TestManaged_Set(t *testing.T) {
	base := &StateBase{}
	state := NewManaged(base, 0)

	state.Set(100)

	if state.Value() != 100 {
		t.Errorf("Expected 100, got %d", state.Value())
	}
}

func TestManaged_Update(t *testing.T) {
	base := &StateBase{}
	state := NewManaged(base, 10)

	state.Update(func(v int) int { return v * 2 })

	if state.Value() != 20 {
		t.Errorf("Expected 20, got %d", state.Value())
	}
}

func TestManaged_StringType(t *testing.T) {
	base := &StateBase{}
	state := NewManaged(base, "hello")

	if state.Value() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", state.Value())
	}

	state.Set("world")

	if state.Value() != "world" {
		t.Errorf("Expected 'world', got '%s'", state.Value())
	}
}

func TestManaged_StructType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	base := &StateBase{}
	state := NewManaged(base, Person{Name: "Alice", Age: 30})

	if state.Value().Name != "Alice" || state.Value().Age != 30 {
		t.Errorf("Unexpected struct value: %+v", state.Value())
	}

	state.Update(func(p Person) Person {
		p.Age++
		return p
	})

	if state.Value().Age != 31 {
		t.Errorf("Expected age 31, got %d", state.Value().Age)
	}
}
