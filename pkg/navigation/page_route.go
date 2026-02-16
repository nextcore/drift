package navigation

import (
	"time"

	"github.com/go-drift/drift/pkg/animation"
	"github.com/go-drift/drift/pkg/core"
)

// TransitionDuration is the default duration for page transitions.
const TransitionDuration = 450 * time.Millisecond

// AnimatedPageRoute provides a route with animated page transitions.
type AnimatedPageRoute struct {
	BaseRoute

	// Builder creates the page content.
	Builder func(ctx core.BuildContext) core.Widget

	// foregroundController drives this route's own slide-in/slide-out animation.
	foregroundController *animation.AnimationController

	// isInitialRoute tracks if this is the first route (no animation needed)
	isInitialRoute bool
}

// NewAnimatedPageRoute creates an AnimatedPageRoute with the given builder and settings.
func NewAnimatedPageRoute(builder func(core.BuildContext) core.Widget, settings RouteSettings) *AnimatedPageRoute {
	return &AnimatedPageRoute{
		BaseRoute: NewBaseRoute(settings),
		Builder:   builder,
	}
}

// ForegroundController returns this route's foreground animation controller.
// Satisfies the AnimatedRoute interface.
func (m *AnimatedPageRoute) ForegroundController() *animation.AnimationController {
	return m.foregroundController
}

// Build returns the page content wrapped in a foreground slide transition.
// Background slide animation is handled by the navigator.
func (m *AnimatedPageRoute) Build(ctx core.BuildContext) core.Widget {
	if m.Builder == nil {
		return nil
	}

	content := m.Builder(ctx)

	// Wrap in foreground slide transition if we have an animation
	if m.foregroundController != nil {
		content = SlideTransition{
			Animation: m.foregroundController,
			Direction: SlideFromRight,
			Child:     content,
		}
	}

	return content
}

// DidPush is called when the route is pushed.
func (m *AnimatedPageRoute) DidPush() {
	// Only animate if not the initial route
	if !m.isInitialRoute {
		m.foregroundController = animation.NewAnimationController(TransitionDuration)
		m.foregroundController.Curve = animation.IOSNavigationCurve
		m.foregroundController.Forward()
	}
}

// SetInitialRoute marks this as the initial route (no animation).
func (m *AnimatedPageRoute) SetInitialRoute() {
	m.isInitialRoute = true
}

// DidPop is called when the route is popped.
func (m *AnimatedPageRoute) DidPop(result any) {
	if m.foregroundController != nil {
		m.foregroundController.Reverse()
	}
}

// PageRoute is a simpler route without transitions.
type PageRoute struct {
	BaseRoute

	// Builder creates the page content.
	Builder func(ctx core.BuildContext) core.Widget
}

// NewPageRoute creates a PageRoute with the given builder and settings.
func NewPageRoute(builder func(core.BuildContext) core.Widget, settings RouteSettings) *PageRoute {
	return &PageRoute{
		BaseRoute: NewBaseRoute(settings),
		Builder:   builder,
	}
}

// Build returns the page content.
func (p *PageRoute) Build(ctx core.BuildContext) core.Widget {
	if p.Builder == nil {
		return nil
	}
	return p.Builder(ctx)
}
