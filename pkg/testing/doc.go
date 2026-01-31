// Package testing provides a widget testing framework for Drift.
//
// # Quick Start
//
// Create a tester, pump a widget, and make assertions:
//
//	func TestMyWidget(t *testing.T) {
//	    tester := drifttest.NewWidgetTesterWithT(t)
//	    tester.PumpWidget(MyWidget{})
//
//	    // Find elements
//	    button := tester.Find(drifttest.ByText("Submit")).First()
//
//	    // Simulate gestures
//	    tester.Tap(drifttest.ByText("Submit"))
//	    tester.Pump()
//
//	    // Assert state
//	    if !tester.Find(drifttest.ByText("Submitted")).Exists() {
//	        t.Error("expected 'Submitted' text")
//	    }
//	}
//
// # Snapshot Testing
//
// Capture and compare render tree snapshots:
//
//	snapshot := tester.CaptureSnapshot()
//	snapshot.MatchesFile(t, "testdata/my_widget.snapshot.json")
//
// Update snapshots with:
//
//	DRIFT_UPDATE_SNAPSHOTS=1 go test ./...
//
// # Animation Testing
//
// Control time for deterministic animation tests:
//
//	tester.Clock().Advance(100 * time.Millisecond)
//	tester.Pump()
//
// # Import Alias
//
// Since this package has the same name as the standard library testing
// package, import it with an alias:
//
//	import drifttest "github.com/go-drift/drift/pkg/testing"
package testing
