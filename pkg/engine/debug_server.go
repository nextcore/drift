package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/go-drift/drift/pkg/layout"
)

// debugServer manages the HTTP server for render tree inspection.
type debugServer struct {
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
}

var debugSrv debugServer

// RenderTreeNode represents a node in the serialized render tree.
// Uses SafeFloat for dimensions that may contain Inf/NaN from layout issues.
type RenderTreeNode struct {
	Type              string           `json:"type"`
	Size              SafeSize         `json:"size"`
	Constraints       *SafeConstraints `json:"constraints,omitempty"`
	Offset            SafeOffset       `json:"offset,omitempty"`
	Depth             int              `json:"depth"`
	NeedsLayout       bool             `json:"needsLayout"`
	NeedsPaint        bool             `json:"needsPaint"`
	IsRepaintBoundary bool             `json:"isRepaintBoundary"`
	Children          []RenderTreeNode `json:"children,omitempty"`
}

// SafeFloat wraps a float64 to handle Inf/NaN in JSON encoding.
type SafeFloat float64

func (f SafeFloat) MarshalJSON() ([]byte, error) {
	v := float64(f)
	if math.IsInf(v, 1) {
		return []byte(`"Infinity"`), nil
	}
	if math.IsInf(v, -1) {
		return []byte(`"-Infinity"`), nil
	}
	if math.IsNaN(v) {
		return []byte(`"NaN"`), nil
	}
	return json.Marshal(v)
}

// SafeSize is a JSON-safe version of graphics.Size.
type SafeSize struct {
	Width  SafeFloat `json:"width"`
	Height SafeFloat `json:"height"`
}

// SafeOffset is a JSON-safe version of graphics.Offset.
type SafeOffset struct {
	X SafeFloat `json:"x"`
	Y SafeFloat `json:"y"`
}

// SafeConstraints is a JSON-safe version of layout.Constraints.
type SafeConstraints struct {
	MinWidth  SafeFloat `json:"minWidth"`
	MaxWidth  SafeFloat `json:"maxWidth"`
	MinHeight SafeFloat `json:"minHeight"`
	MaxHeight SafeFloat `json:"maxHeight"`
}

// startDebugServer starts the HTTP debug server on the specified port.
// Returns the actual port (useful when port=0 for ephemeral allocation).
func startDebugServer(port int) (int, error) {
	debugSrv.mu.Lock()
	defer debugSrv.mu.Unlock()

	if debugSrv.server != nil {
		// Already running - return current port
		if debugSrv.listener != nil {
			return debugSrv.listener.Addr().(*net.TCPAddr).Port, nil
		}
		return port, nil
	}

	// Bind listener first to fail fast on port conflicts
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return 0, fmt.Errorf("debug server listen: %w", err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/tree", handleTree)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/debug", handleDebug)

	server := &http.Server{Handler: mux}
	debugSrv.server = server
	debugSrv.listener = listener

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			// Server failed - clear state so it can be restarted
			debugSrv.mu.Lock()
			debugSrv.server = nil
			debugSrv.listener = nil
			debugSrv.mu.Unlock()
			fmt.Printf("debug server error: %v\n", err)
		}
	}()

	return actualPort, nil
}

// stopDebugServer gracefully shuts down the debug server.
func stopDebugServer() {
	debugSrv.mu.Lock()
	server := debugSrv.server
	debugSrv.server = nil
	debugSrv.listener = nil
	debugSrv.mu.Unlock()

	if server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

// maxTreeDepth limits recursion depth to prevent stack overflow from malformed trees.
const maxTreeDepth = 500

// handleTree returns the render tree as JSON.
func handleTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Recover from panics during serialization
	defer func() {
		if rec := recover(); rec != nil {
			http.Error(w, fmt.Sprintf("panic: %v", rec), http.StatusInternalServerError)
		}
	}()

	frameLock.Lock()
	root := app.rootRender
	if root == nil {
		frameLock.Unlock()
		http.Error(w, "no render tree", http.StatusServiceUnavailable)
		return
	}
	tree := serializeRenderTreeWithDepth(root, 0)
	frameLock.Unlock()

	// Encode to buffer first so we can catch errors
	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("json encode error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// handleHealth returns a simple health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// handleDebug returns diagnostic info about the render tree state.
func handleDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	frameLock.Lock()
	root := app.rootRender
	var info struct {
		HasRoot  bool   `json:"hasRoot"`
		RootType string `json:"rootType,omitempty"`
		RootSize string `json:"rootSize,omitempty"`
	}
	info.HasRoot = root != nil
	if root != nil {
		info.RootType = reflect.TypeOf(root).String()
		size := root.Size()
		info.RootSize = fmt.Sprintf("%.2fx%.2f", size.Width, size.Height)
	}
	frameLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// serializeRenderTreeWithDepth recursively converts a render object tree to JSON-serializable form.
// The depth parameter limits recursion to prevent stack overflow.
func serializeRenderTreeWithDepth(obj layout.RenderObject, depth int) RenderTreeNode {
	size := obj.Size()
	node := RenderTreeNode{
		Type: reflect.TypeOf(obj).String(),
		Size: SafeSize{
			Width:  SafeFloat(size.Width),
			Height: SafeFloat(size.Height),
		},
		NeedsLayout:       getNeedsLayout(obj),
		NeedsPaint:        getNeedsPaint(obj),
		IsRepaintBoundary: obj.IsRepaintBoundary(),
	}

	// Get constraints if available
	if getter, ok := obj.(interface{ Constraints() layout.Constraints }); ok {
		c := getter.Constraints()
		node.Constraints = &SafeConstraints{
			MinWidth:  SafeFloat(c.MinWidth),
			MaxWidth:  SafeFloat(c.MaxWidth),
			MinHeight: SafeFloat(c.MinHeight),
			MaxHeight: SafeFloat(c.MaxHeight),
		}
	}

	// Get depth if available
	if getter, ok := obj.(interface{ Depth() int }); ok {
		node.Depth = getter.Depth()
	}

	// Get offset from parent data if available
	if pd, ok := obj.ParentData().(*layout.BoxParentData); ok {
		node.Offset = SafeOffset{
			X: SafeFloat(pd.Offset.X),
			Y: SafeFloat(pd.Offset.Y),
		}
	}

	// Recurse into children (with depth limit)
	if depth < maxTreeDepth {
		if cv, ok := obj.(layout.ChildVisitor); ok {
			cv.VisitChildren(func(child layout.RenderObject) {
				node.Children = append(node.Children, serializeRenderTreeWithDepth(child, depth+1))
			})
		}
	}

	return node
}

// getNeedsLayout safely retrieves the NeedsLayout flag.
func getNeedsLayout(obj layout.RenderObject) bool {
	if getter, ok := obj.(interface{ NeedsLayout() bool }); ok {
		return getter.NeedsLayout()
	}
	return false
}

// getNeedsPaint safely retrieves the NeedsPaint flag.
func getNeedsPaint(obj layout.RenderObject) bool {
	if getter, ok := obj.(interface{ NeedsPaint() bool }); ok {
		return getter.NeedsPaint()
	}
	return false
}
