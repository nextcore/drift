package core

// ListenableBuilder is a convenience [StatefulWidget] that rebuilds whenever
// a [Listenable] notifies. It collapses the common "subscribe + SetState"
// pattern into a single struct literal:
//
//	core.ListenableBuilder{
//	    Listenable: counter,
//	    Builder: func(ctx core.BuildContext) core.Widget {
//	        return widgets.Text{Content: fmt.Sprint(counter.Value())}
//	    },
//	}
//
// For multiple listenables, merge them with [NewDerived] or use a full
// [StatefulWidget].
type ListenableBuilder struct {
	StatefulBase
	Listenable Listenable
	Builder    func(ctx BuildContext) Widget
}

func (ListenableBuilder) CreateState() State {
	return &listenableBuilderState{}
}

type listenableBuilderState struct {
	StateBase
	unsub func() // removes listener and unregisters disposer
}

func (s *listenableBuilderState) widget() *ListenableBuilder {
	return s.Element().Widget().(*ListenableBuilder)
}

// subscribe registers a listener on l that triggers a rebuild. The listener
// removal is also registered as a disposer so that StateBase.Dispose handles
// cleanup automatically. unsubscribe() is only needed for DidUpdateWidget
// swaps, where we must remove both the listener and the stale disposer entry.
func (s *listenableBuilderState) subscribe(l Listenable) {
	unsub := l.AddListener(func() {
		s.SetState(nil)
	})
	unregister := s.OnDispose(unsub)
	s.unsub = func() {
		unsub()
		unregister()
	}
}

func (s *listenableBuilderState) unsubscribe() {
	if s.unsub != nil {
		s.unsub()
		s.unsub = nil
	}
}

func (s *listenableBuilderState) InitState() {
	w := s.widget()
	if w.Listenable == nil {
		panic("ListenableBuilder: Listenable must not be nil")
	}
	if w.Builder == nil {
		panic("ListenableBuilder: Builder must not be nil")
	}
	s.subscribe(w.Listenable)
}

func (s *listenableBuilderState) DidUpdateWidget(old StatefulWidget) {
	oldW := old.(*ListenableBuilder)
	newW := s.widget()
	if newW.Builder == nil {
		panic("ListenableBuilder: Builder must not be nil")
	}
	if oldW.Listenable != newW.Listenable {
		s.unsubscribe()
		if newW.Listenable == nil {
			panic("ListenableBuilder: Listenable must not be nil")
		}
		s.subscribe(newW.Listenable)
	}
}

func (s *listenableBuilderState) Build(ctx BuildContext) Widget {
	return s.widget().Builder(ctx)
}
