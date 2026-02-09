package engine

import (
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// HitTestPlatformView checks whether a native platform view is the topmost
// hit target at the given pixel coordinates. Returns true if the first
// PointerHandler in the hit test result is a PlatformViewOwner with a
// matching viewID, meaning the platform view should receive the touch.
//
// Called synchronously from the native UI thread (via CGo) before each touch
// is dispatched to a platform view. Both this function and HandlePointer run
// on the same native thread, so they never execute concurrently despite both
// acquiring frameLock.
func HitTestPlatformView(viewID int64, x, y float64) bool {
	frameLock.Lock()
	defer frameLock.Unlock()

	rootRender := app.rootRender
	if rootRender == nil {
		return false
	}

	scale := app.deviceScale
	position := graphics.Offset{X: x / scale, Y: y / scale}

	result := &layout.HitTestResult{}
	if !rootRender.HitTest(position, result) || len(result.Entries) == 0 {
		return false
	}

	// Walk entries to find the first PointerHandler. Non-PointerHandler entries
	// (purely decorative widgets) are skipped, matching normal Drift hit testing
	// where decorations don't absorb touches. The first PointerHandler determines
	// whether the platform view is topmost.
	for _, entry := range result.Entries {
		handler, ok := entry.(layout.PointerHandler)
		if !ok {
			continue
		}
		// Check if this handler owns the platform view in question.
		if owner, ok := handler.(layout.PlatformViewOwner); ok {
			return owner.PlatformViewID() == viewID
		}
		// First interactive handler is not the platform view owner: view is obscured.
		return false
	}

	return false
}
