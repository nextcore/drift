package testing

import (
	"fmt"

	"github.com/go-drift/drift/pkg/gestures"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// pointerState tracks active pointer handlers for gesture simulation.
type pointerState struct {
	handlers []layout.PointerHandler
	position graphics.Offset
}

// nextPointerID is incremented for each new pointer to avoid collisions.
var nextPointerID int64

func allocPointerID() int64 {
	nextPointerID++
	return nextPointerID
}

// Tap simulates a tap at the center of the first element matched by finder.
func (t *WidgetTester) Tap(finder Finder) error {
	result := t.Find(finder)
	if !result.Exists() {
		return fmt.Errorf("Tap: finder matched no elements: %s", finder.Description())
	}

	ro := extractRenderObject(result.First())
	if ro == nil {
		return fmt.Errorf("Tap: element has no render object: %s", finder.Description())
	}

	center := renderCenter(ro)
	return t.TapAt(center)
}

// TapAt simulates a tap at the given logical position.
func (t *WidgetTester) TapAt(pos graphics.Offset) error {
	id := allocPointerID()
	if err := t.SendPointerDown(pos, int(id)); err != nil {
		return err
	}
	return t.SendPointerUp(pos, int(id))
}

// Drag simulates a drag gesture on the first element matched by finder.
func (t *WidgetTester) Drag(finder Finder, delta graphics.Offset) error {
	result := t.Find(finder)
	if !result.Exists() {
		return fmt.Errorf("Drag: finder matched no elements: %s", finder.Description())
	}
	ro := extractRenderObject(result.First())
	if ro == nil {
		return fmt.Errorf("Drag: element has no render object: %s", finder.Description())
	}
	start := renderCenter(ro)
	return t.DragFrom(start, delta)
}

// DragFrom simulates a drag from start by delta.
func (t *WidgetTester) DragFrom(start, delta graphics.Offset) error {
	id := allocPointerID()
	if err := t.SendPointerDown(start, int(id)); err != nil {
		return err
	}
	end := graphics.Offset{X: start.X + delta.X, Y: start.Y + delta.Y}
	if err := t.SendPointerMove(end, int(id)); err != nil {
		return err
	}
	return t.SendPointerUp(end, int(id))
}

// Fling simulates a fling gesture with velocity. The velocity parameter
// determines the final delta speed reported to gesture recognizers.
func (t *WidgetTester) Fling(finder Finder, delta, velocity graphics.Offset) error {
	result := t.Find(finder)
	if !result.Exists() {
		return fmt.Errorf("Fling: finder matched no elements: %s", finder.Description())
	}
	ro := extractRenderObject(result.First())
	if ro == nil {
		return fmt.Errorf("Fling: element has no render object: %s", finder.Description())
	}
	start := renderCenter(ro)

	id := allocPointerID()
	if err := t.SendPointerDown(start, int(id)); err != nil {
		return err
	}

	// Emit intermediate move events to build velocity
	steps := 10
	for i := 1; i <= steps; i++ {
		frac := float64(i) / float64(steps)
		pos := graphics.Offset{
			X: start.X + delta.X*frac,
			Y: start.Y + delta.Y*frac,
		}
		if err := t.SendPointerMove(pos, int(id)); err != nil {
			return err
		}
	}

	end := graphics.Offset{X: start.X + delta.X, Y: start.Y + delta.Y}
	return t.SendPointerUp(end, int(id))
}

// SendPointerDown sends a pointer-down event at pos with the given pointer ID.
func (t *WidgetTester) SendPointerDown(pos graphics.Offset, pointerID int) error {
	return t.sendPointer(gestures.PointerEvent{
		PointerID: int64(pointerID),
		Position:  pos,
		Phase:     gestures.PointerPhaseDown,
	})
}

// SendPointerMove sends a pointer-move event at pos with the given pointer ID.
func (t *WidgetTester) SendPointerMove(pos graphics.Offset, pointerID int) error {
	state := t.pointers[pointerID]
	delta := graphics.Offset{}
	if state != nil {
		delta = graphics.Offset{X: pos.X - state.position.X, Y: pos.Y - state.position.Y}
	}
	return t.sendPointer(gestures.PointerEvent{
		PointerID: int64(pointerID),
		Position:  pos,
		Delta:     delta,
		Phase:     gestures.PointerPhaseMove,
	})
}

// SendPointerUp sends a pointer-up event at pos with the given pointer ID.
func (t *WidgetTester) SendPointerUp(pos graphics.Offset, pointerID int) error {
	state := t.pointers[pointerID]
	delta := graphics.Offset{}
	if state != nil {
		delta = graphics.Offset{X: pos.X - state.position.X, Y: pos.Y - state.position.Y}
	}
	return t.sendPointer(gestures.PointerEvent{
		PointerID: int64(pointerID),
		Position:  pos,
		Delta:     delta,
		Phase:     gestures.PointerPhaseUp,
	})
}

// SendPointerCancel sends a pointer-cancel event for the given pointer ID.
func (t *WidgetTester) SendPointerCancel(pointerID int) error {
	state := t.pointers[pointerID]
	pos := graphics.Offset{}
	if state != nil {
		pos = state.position
	}
	return t.sendPointer(gestures.PointerEvent{
		PointerID: int64(pointerID),
		Position:  pos,
		Phase:     gestures.PointerPhaseCancel,
	})
}

func (t *WidgetTester) sendPointer(event gestures.PointerEvent) error {
	if t.rootRender == nil {
		return fmt.Errorf("no widget mounted")
	}

	pointerID := int(event.PointerID)

	switch event.Phase {
	case gestures.PointerPhaseDown:
		// Hit test
		result := &layout.HitTestResult{}
		t.rootRender.HitTest(event.Position, result)

		// Collect handlers
		handlers := collectPointerHandlers(result.Entries)
		t.pointers[pointerID] = &pointerState{
			handlers: handlers,
			position: event.Position,
		}

		// Dispatch to handlers
		for _, h := range handlers {
			h.HandlePointer(event)
		}

		// Close arena
		gestures.DefaultArena.Close(event.PointerID)

	case gestures.PointerPhaseMove:
		state := t.pointers[pointerID]
		if state == nil {
			return nil
		}
		state.position = event.Position
		for _, h := range state.handlers {
			h.HandlePointer(event)
		}

	case gestures.PointerPhaseUp:
		state := t.pointers[pointerID]
		if state == nil {
			return nil
		}
		state.position = event.Position
		for _, h := range state.handlers {
			h.HandlePointer(event)
		}
		gestures.DefaultArena.Sweep(event.PointerID)
		delete(t.pointers, pointerID)

	case gestures.PointerPhaseCancel:
		state := t.pointers[pointerID]
		if state == nil {
			return nil
		}
		for _, h := range state.handlers {
			h.HandlePointer(event)
		}
		gestures.DefaultArena.Sweep(event.PointerID)
		delete(t.pointers, pointerID)
	}

	return nil
}

// collectPointerHandlers extracts unique PointerHandler instances from
// hit test entries, preserving paint order.
func collectPointerHandlers(entries []layout.RenderObject) []layout.PointerHandler {
	handlers := make([]layout.PointerHandler, 0, len(entries))
	seen := make(map[layout.PointerHandler]struct{})
	for _, entry := range entries {
		if h, ok := entry.(layout.PointerHandler); ok {
			if _, exists := seen[h]; !exists {
				seen[h] = struct{}{}
				handlers = append(handlers, h)
			}
		}
	}
	return handlers
}

// renderCenter returns the center of a render object in absolute (root-relative)
// coordinates by walking the full ancestor chain.
func renderCenter(ro layout.RenderObject) graphics.Offset {
	size := ro.Size()
	center := graphics.Offset{X: size.Width / 2, Y: size.Height / 2}

	abs := absoluteOffset(ro)
	return graphics.Offset{X: abs.X + center.X, Y: abs.Y + center.Y}
}

// absoluteOffset walks up the parent chain accumulating offsets from
// BoxParentData to compute the root-relative position of a render object.
func absoluteOffset(ro layout.RenderObject) graphics.Offset {
	offset := graphics.Offset{}
	cur := ro
	for cur != nil {
		if pd, ok := cur.ParentData().(*layout.BoxParentData); ok {
			offset.X += pd.Offset.X
			offset.Y += pd.Offset.Y
		}
		if parent, ok := cur.(interface{ Parent() layout.RenderObject }); ok {
			cur = parent.Parent()
		} else {
			break
		}
	}
	return offset
}
