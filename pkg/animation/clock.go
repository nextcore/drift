package animation

import "time"

// Clock provides time for animations. The default implementation uses
// system time. Tests can inject a fake clock via SetClock to control
// animation timing deterministically.
type Clock interface {
	Now() time.Time
}

// realClock uses system time.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// clock is the package-level time source, replaceable for testing.
var clock Clock = realClock{}

// SetClock replaces the animation clock. Returns the previous clock
// so callers can restore it during cleanup.
func SetClock(c Clock) Clock {
	prev := clock
	clock = c
	return prev
}

// Now returns the current time from the active clock.
func Now() time.Time { return clock.Now() }
