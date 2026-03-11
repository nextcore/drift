package core

import "testing"

// newListenableBuilderState wires up a listenableBuilderState with the given
// widget, mimicking what StatefulElement.Mount does. Returns the state, element,
// and build owner for assertions.
func newListenableBuilderState(w *ListenableBuilder) (*listenableBuilderState, *StatefulElement, *BuildOwner) {
	s := &listenableBuilderState{}
	owner := NewBuildOwner()
	elem := &StatefulElement{}
	elem.buildOwner = owner
	elem.self = elem
	elem.widget = w
	s.SetElement(elem)
	return s, elem, owner
}

func TestListenableBuilder_SubscribesOnInit(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, _, _ := newListenableBuilderState(w)

	s.InitState()

	if n.ListenerCount() != 1 {
		t.Errorf("expected 1 listener, got %d", n.ListenerCount())
	}
}

func TestListenableBuilder_UnsubscribesOnDispose(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, _, _ := newListenableBuilderState(w)
	s.InitState()

	s.Dispose()

	if n.ListenerCount() != 0 {
		t.Errorf("expected 0 listeners after dispose, got %d", n.ListenerCount())
	}
}

func TestListenableBuilder_MarksElementDirtyOnNotify(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, _, owner := newListenableBuilderState(w)
	s.InitState()

	n.Notify()

	if count := countDirty(owner); count != 1 {
		t.Errorf("expected 1 dirty element, got %d", count)
	}
}

func TestListenableBuilder_DidUpdateWidget_Resubscribes(t *testing.T) {
	old := &Notifier{}
	w := &ListenableBuilder{Listenable: old, Builder: func(ctx BuildContext) Widget { return nil }}
	s, elem, _ := newListenableBuilderState(w)
	s.InitState()

	// Swap to a new listenable.
	newN := &Notifier{}
	newW := &ListenableBuilder{Listenable: newN, Builder: w.Builder}
	elem.widget = newW
	s.DidUpdateWidget(w)

	if old.ListenerCount() != 0 {
		t.Errorf("old listenable: expected 0 listeners, got %d", old.ListenerCount())
	}
	if newN.ListenerCount() != 1 {
		t.Errorf("new listenable: expected 1 listener, got %d", newN.ListenerCount())
	}
}

func TestListenableBuilder_DidUpdateWidget_NoOpWhenUnchanged(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, elem, _ := newListenableBuilderState(w)
	s.InitState()

	// Update with same listenable pointer.
	newW := &ListenableBuilder{Listenable: n, Builder: w.Builder}
	elem.widget = newW
	s.DidUpdateWidget(w)

	if n.ListenerCount() != 1 {
		t.Errorf("expected listener count to stay 1, got %d", n.ListenerCount())
	}
}

func TestListenableBuilder_PanicsOnNilListenableAtInit(t *testing.T) {
	w := &ListenableBuilder{Listenable: nil, Builder: func(ctx BuildContext) Widget { return nil }}
	s, _, _ := newListenableBuilderState(w)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Listenable")
		}
	}()
	s.InitState()
}

func TestListenableBuilder_PanicsOnNilBuilderAtInit(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: nil}
	s, _, _ := newListenableBuilderState(w)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Builder")
		}
	}()
	s.InitState()
}

func TestListenableBuilder_PanicsOnNilBuilderInDidUpdateWidget(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, elem, _ := newListenableBuilderState(w)
	s.InitState()

	newW := &ListenableBuilder{Listenable: n, Builder: nil}
	elem.widget = newW

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Builder in DidUpdateWidget")
		}
	}()
	s.DidUpdateWidget(w)
}

func TestListenableBuilder_PanicsOnNilListenableInDidUpdateWidget(t *testing.T) {
	n := &Notifier{}
	w := &ListenableBuilder{Listenable: n, Builder: func(ctx BuildContext) Widget { return nil }}
	s, elem, _ := newListenableBuilderState(w)
	s.InitState()

	newW := &ListenableBuilder{Listenable: nil, Builder: w.Builder}
	elem.widget = newW

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Listenable in DidUpdateWidget")
		}
	}()
	s.DidUpdateWidget(w)
}

type sentinelWidget struct{ StatelessBase }

func (sentinelWidget) Build(BuildContext) Widget { return nil }

func TestListenableBuilder_BuildDelegatesToBuilder(t *testing.T) {
	sentinel := sentinelWidget{}
	n := &Notifier{}
	w := &ListenableBuilder{
		Listenable: n,
		Builder:    func(ctx BuildContext) Widget { return sentinel },
	}
	s, _, _ := newListenableBuilderState(w)
	s.InitState()

	got := s.Build(nil)
	if got != sentinel {
		t.Errorf("expected sentinel widget, got %v", got)
	}
}
