package platform

import (
	"sync"
	"sync/atomic"

	"github.com/go-drift/drift/pkg/rendering"
)

// PlatformView represents a native view embedded in Drift UI.
type PlatformView interface {
	// ViewID returns the unique identifier for this view.
	ViewID() int64

	// ViewType returns the type identifier for this view (e.g., "native_webview").
	ViewType() string

	// Create initializes the native view with given parameters.
	Create(params map[string]any) error

	// Dispose cleans up the native view.
	Dispose()

	// SetSize updates the view size in logical pixels.
	SetSize(size rendering.Size)

	// SetOffset updates the view position in logical pixels.
	SetOffset(offset rendering.Offset)

	// SetVisible shows or hides the native view.
	SetVisible(visible bool)
}

// PlatformViewFactory creates platform views of a specific type.
type PlatformViewFactory interface {
	// Create creates a new platform view instance.
	Create(viewID int64, params map[string]any) (PlatformView, error)

	// ViewType returns the view type this factory creates.
	ViewType() string
}

// PlatformViewRegistry manages platform view types and instances.
type PlatformViewRegistry struct {
	factories map[string]PlatformViewFactory
	views     map[int64]PlatformView
	nextID    atomic.Int64
	mu        sync.RWMutex
	channel   *MethodChannel
}

var platformViewRegistry *PlatformViewRegistry

// GetPlatformViewRegistry returns the global platform view registry.
func GetPlatformViewRegistry() *PlatformViewRegistry {
	if platformViewRegistry == nil {
		platformViewRegistry = newPlatformViewRegistry()
	}
	return platformViewRegistry
}

func newPlatformViewRegistry() *PlatformViewRegistry {
	r := &PlatformViewRegistry{
		factories: make(map[string]PlatformViewFactory),
		views:     make(map[int64]PlatformView),
		channel:   NewMethodChannel("drift/platform_views"),
	}

	// Handle incoming calls from native
	r.channel.SetHandler(r.handleMethodCall)

	// Also listen for events from native (text changes, focus, etc.)
	eventChannel := NewEventChannel("drift/platform_views")
	eventChannel.Listen(EventHandler{
		OnEvent: func(data any) {
			r.handleEvent(data)
		},
	})

	return r
}

// handleEvent processes events from native platform views.
func (r *PlatformViewRegistry) handleEvent(data any) {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return
	}

	method, _ := dataMap["method"].(string)
	if method == "" {
		return
	}

	switch method {
	case "onTextChanged":
		r.handleTextChanged(dataMap)
	case "onAction":
		r.handleAction(dataMap)
	case "onFocusChanged":
		r.handleFocusChanged(dataMap)
	case "onSwitchChanged":
		r.handleSwitchChanged(dataMap)
	}
}

// RegisterFactory registers a factory for a platform view type.
func (r *PlatformViewRegistry) RegisterFactory(factory PlatformViewFactory) {
	r.mu.Lock()
	r.factories[factory.ViewType()] = factory
	r.mu.Unlock()
}

// Create creates a new platform view of the given type.
func (r *PlatformViewRegistry) Create(viewType string, params map[string]any) (PlatformView, error) {
	r.mu.RLock()
	factory, ok := r.factories[viewType]
	r.mu.RUnlock()

	if !ok {
		return nil, ErrViewTypeNotFound
	}

	viewID := r.nextID.Add(1)

	view, err := factory.Create(viewID, params)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.views[viewID] = view
	r.mu.Unlock()

	// Notify native to create the view
	_, err = r.channel.Invoke("create", map[string]any{
		"viewId":   viewID,
		"viewType": viewType,
		"params":   params,
	})
	if err != nil {
		r.mu.Lock()
		delete(r.views, viewID)
		r.mu.Unlock()
		return nil, err
	}

	return view, nil
}

// Dispose destroys a platform view.
func (r *PlatformViewRegistry) Dispose(viewID int64) {
	r.mu.Lock()
	view, ok := r.views[viewID]
	if ok {
		delete(r.views, viewID)
	}
	r.mu.Unlock()

	if ok {
		view.Dispose()
		// Notify native to destroy the view
		r.channel.Invoke("dispose", map[string]any{
			"viewId": viewID,
		})
	}
}

// GetView returns a platform view by ID.
func (r *PlatformViewRegistry) GetView(viewID int64) PlatformView {
	r.mu.RLock()
	view := r.views[viewID]
	r.mu.RUnlock()
	return view
}

// UpdateViewGeometry notifies native of a view's position and size change.
func (r *PlatformViewRegistry) UpdateViewGeometry(viewID int64, offset rendering.Offset, size rendering.Size) error {
	_, err := r.channel.Invoke("setGeometry", map[string]any{
		"viewId": viewID,
		"x":      offset.X,
		"y":      offset.Y,
		"width":  size.Width,
		"height": size.Height,
	})
	return err
}

// SetViewVisible notifies native to show or hide a view.
func (r *PlatformViewRegistry) SetViewVisible(viewID int64, visible bool) error {
	_, err := r.channel.Invoke("setVisible", map[string]any{
		"viewId":  viewID,
		"visible": visible,
	})
	return err
}

// SetViewEnabled notifies native to enable or disable a view.
func (r *PlatformViewRegistry) SetViewEnabled(viewID int64, enabled bool) error {
	_, err := r.channel.Invoke("setEnabled", map[string]any{
		"viewId":  viewID,
		"enabled": enabled,
	})
	return err
}

// InvokeViewMethod invokes a method on a specific platform view.
func (r *PlatformViewRegistry) InvokeViewMethod(viewID int64, method string, args map[string]any) (any, error) {
	// Clone the args map to avoid mutating the caller's map
	size := 2
	if args != nil {
		size += len(args)
	}
	invokeArgs := make(map[string]any, size)
	for k, v := range args { // safe: range over nil map is no-op
		invokeArgs[k] = v
	}
	invokeArgs["viewId"] = viewID
	invokeArgs["method"] = method
	return r.channel.Invoke("invokeViewMethod", invokeArgs)
}

// handleMethodCall processes incoming method calls from native code.
func (r *PlatformViewRegistry) handleMethodCall(method string, args any) (any, error) {
	argsMap, _ := args.(map[string]any)

	switch method {
	case "onViewCreated":
		// Native has finished creating the view
		return nil, nil

	case "onViewDisposed":
		// Native has finished disposing the view
		return nil, nil

	case "onTextChanged":
		return r.handleTextChanged(argsMap)

	case "onAction":
		return r.handleAction(argsMap)

	case "onFocusChanged":
		return r.handleFocusChanged(argsMap)

	case "onSwitchChanged":
		return r.handleSwitchChanged(argsMap)

	default:
		return nil, ErrMethodNotFound
	}
}

func (r *PlatformViewRegistry) handleTextChanged(args map[string]any) (any, error) {
	viewID, _ := toInt64(args["viewId"])
	text, _ := args["text"].(string)
	selBase, _ := toInt(args["selectionBase"])
	selExt, _ := toInt(args["selectionExtent"])

	r.mu.RLock()
	view := r.views[viewID]
	r.mu.RUnlock()

	if textInput, ok := view.(*TextInputView); ok {
		textInput.handleTextChanged(text, selBase, selExt)
	}
	return nil, nil
}

func (r *PlatformViewRegistry) handleAction(args map[string]any) (any, error) {
	viewID, _ := toInt64(args["viewId"])
	action, _ := toInt(args["action"])

	r.mu.RLock()
	view := r.views[viewID]
	r.mu.RUnlock()

	if textInput, ok := view.(*TextInputView); ok {
		textInput.handleAction(TextInputAction(action))
	}
	return nil, nil
}

func (r *PlatformViewRegistry) handleFocusChanged(args map[string]any) (any, error) {
	viewID, _ := toInt64(args["viewId"])
	focused, _ := args["focused"].(bool)

	r.mu.RLock()
	view := r.views[viewID]
	r.mu.RUnlock()

	if textInput, ok := view.(*TextInputView); ok {
		textInput.handleFocusChanged(focused)
	}
	return nil, nil
}

func (r *PlatformViewRegistry) handleSwitchChanged(args map[string]any) (any, error) {
	viewID, _ := toInt64(args["viewId"])
	value, _ := args["value"].(bool)

	r.mu.RLock()
	view := r.views[viewID]
	r.mu.RUnlock()

	if switchView, ok := view.(*SwitchView); ok {
		switchView.handleValueChanged(value)
	}
	return nil, nil
}

// basePlatformView provides common implementation for platform views.
type basePlatformView struct {
	viewID   int64
	viewType string
	offset   rendering.Offset
	size     rendering.Size
	visible  bool
}

func (v *basePlatformView) ViewID() int64 {
	return v.viewID
}

func (v *basePlatformView) ViewType() string {
	return v.viewType
}

func (v *basePlatformView) SetSize(size rendering.Size) {
	v.size = size
	GetPlatformViewRegistry().UpdateViewGeometry(v.viewID, v.offset, v.size)
}

func (v *basePlatformView) SetOffset(offset rendering.Offset) {
	v.offset = offset
	GetPlatformViewRegistry().UpdateViewGeometry(v.viewID, v.offset, v.size)
}

func (v *basePlatformView) SetVisible(visible bool) {
	v.visible = visible
	GetPlatformViewRegistry().SetViewVisible(v.viewID, visible)
}

func (v *basePlatformView) SetEnabled(enabled bool) {
	GetPlatformViewRegistry().SetViewEnabled(v.viewID, enabled)
}
