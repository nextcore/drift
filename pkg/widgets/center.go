package widgets

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/layout"
)

// Center positions its child at the center of the available space.
//
// Center expands to fill available space (like [Expanded]), then centers
// the child within that space. The child is given loose constraints,
// allowing it to size itself.
//
// Center is equivalent to Align{Alignment: layout.AlignmentCenter}.
//
// Example:
//
//	Center{Child: Text{Content: "Hello, World!"}}
//
// For more control over alignment, use [Container] with an Alignment field,
// or wrap the child in an [Align] widget.
type Center struct {
	core.StatelessBase
	Child core.Widget
}

func (c Center) Build(ctx core.BuildContext) core.Widget {
	return Align{
		Alignment: layout.AlignmentCenter,
		Child:     c.Child,
	}
}
