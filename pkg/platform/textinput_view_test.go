package platform

import (
	"testing"
)

func TestTextInputView_HandleTextChangedUpdatesState(t *testing.T) {
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{Obscure: true}, nil)

	view.handleTextChanged("hello", 5, 5)

	if view.Text() != "hello" {
		t.Errorf("Text() = %q, want %q", view.Text(), "hello")
	}
	base, ext := view.Selection()
	if base != 5 || ext != 5 {
		t.Errorf("Selection() = (%d, %d), want (5, 5)", base, ext)
	}
}

func TestTextInputView_HandleTextChangedRapid(t *testing.T) {
	// Simulate rapid typing: after each handleTextChanged the view's
	// stored state should reflect the latest native values.
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{Obscure: true}, nil)

	view.handleTextChanged("p", 1, 1)
	view.handleTextChanged("pa", 2, 2)
	view.handleTextChanged("pas", 3, 3)
	view.handleTextChanged("pass", 4, 4)

	if view.Text() != "pass" {
		t.Errorf("Text() = %q, want %q", view.Text(), "pass")
	}
	base, ext := view.Selection()
	if base != 4 || ext != 4 {
		t.Errorf("Selection() = (%d, %d), want (4, 4)", base, ext)
	}
}

func TestTextInputView_StateMatchesControllerAfterNativeChange(t *testing.T) {
	// The widget layer skips redundant SetValue calls by comparing the
	// controller's value against the platform view's stored state.
	// After handleTextChanged, both should hold the same values.
	setupTestBridge(t)

	var receivedText string
	var receivedBase, receivedExt int

	client := &testTextInputClient{
		onTextChanged: func(text string, selBase, selExt int) {
			receivedText = text
			receivedBase = selBase
			receivedExt = selExt
		},
	}

	view := NewTextInputView(1, TextInputViewConfig{Obscure: true}, client)

	view.handleTextChanged("password", 8, 8)

	// The client receives the same values that a controller.SetValue would store.
	// The view's stored state should match.
	if view.Text() != receivedText {
		t.Errorf("view.Text() = %q, client got %q", view.Text(), receivedText)
	}
	base, ext := view.Selection()
	if base != receivedBase || ext != receivedExt {
		t.Errorf("view.Selection() = (%d, %d), client got (%d, %d)", base, ext, receivedBase, receivedExt)
	}
}

func TestTextInputView_SetValueSendsToNative(t *testing.T) {
	bridge := setupTestBridge(t)

	r := GetPlatformViewRegistry()
	view, err := r.Create("textinput", map[string]any{"obscure": true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	textView := view.(*TextInputView)
	defer r.Dispose(textView.ViewID())

	bridge.reset()

	textView.SetValue(TextEditingValue{
		Text:      "test",
		Selection: TextSelection{BaseOffset: 4, ExtentOffset: 4},
	})

	bridge.mu.Lock()
	var found bool
	for _, c := range bridge.calls {
		if c.method == "invokeViewMethod" {
			found = true
			break
		}
	}
	bridge.mu.Unlock()

	if !found {
		t.Error("expected invokeViewMethod call for SetValue")
	}
}

func TestTextInputView_SetValueUpdatesStoredState(t *testing.T) {
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{}, nil)

	view.SetValue(TextEditingValue{
		Text:      "updated",
		Selection: TextSelection{BaseOffset: 3, ExtentOffset: 7},
	})

	if view.Text() != "updated" {
		t.Errorf("Text() = %q, want %q", view.Text(), "updated")
	}
	base, ext := view.Selection()
	if base != 3 || ext != 7 {
		t.Errorf("Selection() = (%d, %d), want (3, 7)", base, ext)
	}
}

func TestTextInputView_SnapshotConsistentRead(t *testing.T) {
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{}, nil)
	view.handleTextChanged("hello", 3, 5)

	text, base, ext := view.Snapshot()
	if text != "hello" || base != 3 || ext != 5 {
		t.Errorf("Snapshot() = (%q, %d, %d), want (\"hello\", 3, 5)", text, base, ext)
	}
}

func TestTextInputView_SnapshotDetectsProgrammaticChange(t *testing.T) {
	// After a native text change, the view's Snapshot matches the values
	// the controller would hold, so the widget layer skips SetValue.
	// But if the controller is updated programmatically (e.g., validation
	// reformats text), the snapshot will differ, and SetValue should be sent.
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{}, nil)

	// Native sends "user@"
	view.handleTextChanged("user@", 5, 5)

	// Programmatic change would set a different value on the controller.
	// The comparison in DidUpdateWidget checks controller value against snapshot.
	programmaticText := "user@example.com"
	programmaticSel := 16

	text, base, ext := view.Snapshot()
	if text == programmaticText && base == programmaticSel && ext == programmaticSel {
		t.Error("snapshot should differ from programmatic value, so SetValue gets sent")
	}
}

func TestTextInputView_NilClientDoesNotPanic(t *testing.T) {
	setupTestBridge(t)

	view := NewTextInputView(1, TextInputViewConfig{}, nil)

	// Should not panic with nil client.
	view.handleTextChanged("text", 4, 4)
	view.handleAction(TextInputActionDone)
	view.handleFocusChanged(true)
}

func TestTextInputViewConfig_EqualityForIdenticalValues(t *testing.T) {
	// The widget layer uses struct equality to skip redundant config updates.
	// Verify that identical configs compare equal.
	a := TextInputViewConfig{
		FontFamily:       "Roboto",
		FontSize:         16,
		FontWeight:       400,
		TextColor:        0xFF000000,
		PlaceholderColor: 0xFF999999,
		Obscure:          true,
		Autocorrect:      false,
		KeyboardType:     KeyboardTypePassword,
		InputAction:      TextInputActionDone,
		Capitalization:   TextCapitalizationNone,
		PaddingLeft:      12,
		PaddingTop:       8,
		PaddingRight:     12,
		PaddingBottom:    8,
		Placeholder:      "Password",
	}
	b := a // copy

	if a != b {
		t.Error("identical TextInputViewConfig values should be equal")
	}
}

func TestTextInputViewConfig_InequalityForDifferentValues(t *testing.T) {
	base := TextInputViewConfig{
		FontSize: 16,
		Obscure:  true,
	}

	tests := []struct {
		name   string
		modify func(c *TextInputViewConfig)
	}{
		{"Obscure", func(c *TextInputViewConfig) { c.Obscure = false }},
		{"FontSize", func(c *TextInputViewConfig) { c.FontSize = 14 }},
		{"KeyboardType", func(c *TextInputViewConfig) { c.KeyboardType = KeyboardTypeEmail }},
		{"Placeholder", func(c *TextInputViewConfig) { c.Placeholder = "different" }},
		{"PaddingLeft", func(c *TextInputViewConfig) { c.PaddingLeft = 10 }},
		{"TextColor", func(c *TextInputViewConfig) { c.TextColor = 0xFFFF0000 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modified := base
			tc.modify(&modified)
			if base == modified {
				t.Errorf("configs differing in %s should not be equal", tc.name)
			}
		})
	}
}

// testTextInputClient implements TextInputViewClient for testing.
type testTextInputClient struct {
	onTextChanged  func(string, int, int)
	onAction       func(TextInputAction)
	onFocusChanged func(bool)
}

func (c *testTextInputClient) OnTextChanged(text string, selBase, selExt int) {
	if c.onTextChanged != nil {
		c.onTextChanged(text, selBase, selExt)
	}
}

func (c *testTextInputClient) OnAction(action TextInputAction) {
	if c.onAction != nil {
		c.onAction(action)
	}
}

func (c *testTextInputClient) OnFocusChanged(focused bool) {
	if c.onFocusChanged != nil {
		c.onFocusChanged(focused)
	}
}
