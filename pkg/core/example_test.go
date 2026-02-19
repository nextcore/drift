package core_test

import (
	"fmt"

	"github.com/go-drift/drift/pkg/core"
)

// This example shows how to create an Observable for reactive state.
// Observable is thread-safe and can be shared across goroutines.
func ExampleObservable() {
	// Create an observable with an initial value
	counter := core.NewObservable(0)

	// Add a listener that fires when the value changes
	unsub := counter.AddListener(func(value int) {
		fmt.Printf("Counter changed to: %d\n", value)
	})

	// Update the value - this triggers all listeners
	counter.Set(5)

	// Read the current value
	current := counter.Value()
	fmt.Printf("Current value: %d\n", current)

	// Clean up when done
	unsub()

	// Output:
	// Counter changed to: 5
	// Current value: 5
}

// This example shows how to use Observable with a custom equality function.
// This is useful when you want to avoid unnecessary updates.
func ExampleNewObservableWithEquality() {
	type User struct {
		ID   int
		Name string
	}

	// Only notify listeners when the user ID changes
	user := core.NewObservableWithEquality(User{ID: 1, Name: "Alice"}, func(a, b User) bool {
		return a.ID == b.ID
	})

	user.AddListener(func(u User) {
		fmt.Printf("User changed: %s\n", u.Name)
	})

	// This won't trigger listeners because ID is the same
	user.Set(User{ID: 1, Name: "Alice Updated"})

	// This will trigger listeners because ID changed
	user.Set(User{ID: 2, Name: "Bob"})

	// Output:
	// User changed: Bob
}

// This example shows the Notifier type for event broadcasting.
// Unlike Observable, Notifier doesn't hold a value.
func ExampleNotifier() {
	refresh := core.NewNotifier()

	// Add a listener
	unsub := refresh.AddListener(func() {
		fmt.Println("Refresh triggered!")
	})

	// Trigger the notification
	refresh.Notify()

	// Clean up
	unsub()

	// Output:
	// Refresh triggered!
}

// This example shows how to use Managed for automatic rebuilds.
// Managed wraps a value and triggers rebuilds when it changes.
func ExampleManaged() {
	// Direct usage for demonstration:
	base := &core.StateBase{}
	count := core.NewManaged(base, 0)

	// Get the current value
	fmt.Printf("Initial: %d\n", count.Value())

	// Update using transform function
	count.Update(func(v int) int { return v + 10 })
	fmt.Printf("After update: %d\n", count.Value())

	// Output:
	// Initial: 0
	// After update: 10
}

// This example shows how to create a custom controller.
func ExampleControllerBase() {
	controller := &core.ControllerBase{}
	unsub := controller.AddListener(func() {
		fmt.Println("Controller notified")
	})
	controller.NotifyListeners()
	unsub()
	controller.Dispose()

	// Output:
	// Controller notified
}
