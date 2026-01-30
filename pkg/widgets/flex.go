package widgets

import (
	"fmt"
	"log"
	"math"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/layout"
)

// Axis represents the layout direction.
// AxisVertical is the zero value, making it the default for ScrollDirection fields.
type Axis int

const (
	AxisVertical Axis = iota
	AxisHorizontal
)

// String returns a human-readable representation of the axis.
func (a Axis) String() string {
	switch a {
	case AxisVertical:
		return "vertical"
	case AxisHorizontal:
		return "horizontal"
	default:
		return fmt.Sprintf("Axis(%d)", int(a))
	}
}

// MainAxisAlignment controls how children are positioned along the main axis
// (horizontal for [Row], vertical for [Column]).
type MainAxisAlignment int

const (
	// MainAxisAlignmentStart places children at the start (left for Row, top for Column).
	MainAxisAlignmentStart MainAxisAlignment = iota
	// MainAxisAlignmentEnd places children at the end (right for Row, bottom for Column).
	MainAxisAlignmentEnd
	// MainAxisAlignmentCenter centers children along the main axis.
	MainAxisAlignmentCenter
	// MainAxisAlignmentSpaceBetween distributes free space evenly between children.
	// No space before the first or after the last child.
	MainAxisAlignmentSpaceBetween
	// MainAxisAlignmentSpaceAround distributes free space evenly, with half-sized
	// spaces at the start and end.
	MainAxisAlignmentSpaceAround
	// MainAxisAlignmentSpaceEvenly distributes free space evenly, including
	// equal space before the first and after the last child.
	MainAxisAlignmentSpaceEvenly
)

// String returns a human-readable representation of the main axis alignment.
func (a MainAxisAlignment) String() string {
	switch a {
	case MainAxisAlignmentStart:
		return "start"
	case MainAxisAlignmentEnd:
		return "end"
	case MainAxisAlignmentCenter:
		return "center"
	case MainAxisAlignmentSpaceBetween:
		return "space_between"
	case MainAxisAlignmentSpaceAround:
		return "space_around"
	case MainAxisAlignmentSpaceEvenly:
		return "space_evenly"
	default:
		return fmt.Sprintf("MainAxisAlignment(%d)", int(a))
	}
}

// CrossAxisAlignment controls how children are positioned along the cross axis
// (vertical for [Row], horizontal for [Column]).
type CrossAxisAlignment int

const (
	// CrossAxisAlignmentStart places children at the start of the cross axis.
	CrossAxisAlignmentStart CrossAxisAlignment = iota
	// CrossAxisAlignmentEnd places children at the end of the cross axis.
	CrossAxisAlignmentEnd
	// CrossAxisAlignmentCenter centers children along the cross axis.
	CrossAxisAlignmentCenter
	// CrossAxisAlignmentStretch stretches children to fill the cross axis.
	CrossAxisAlignmentStretch
)

// String returns a human-readable representation of the cross axis alignment.
func (a CrossAxisAlignment) String() string {
	switch a {
	case CrossAxisAlignmentStart:
		return "start"
	case CrossAxisAlignmentEnd:
		return "end"
	case CrossAxisAlignmentCenter:
		return "center"
	case CrossAxisAlignmentStretch:
		return "stretch"
	default:
		return fmt.Sprintf("CrossAxisAlignment(%d)", int(a))
	}
}

// MainAxisSize controls how much space the flex container takes along its main axis.
type MainAxisSize int

const (
	// MainAxisSizeMin sizes the container to fit its children (shrink-wrap).
	MainAxisSizeMin MainAxisSize = iota
	// MainAxisSizeMax expands to fill all available space along the main axis.
	// This is required for [Expanded] children to receive space.
	MainAxisSizeMax
)

// String returns a human-readable representation of the main axis size.
func (s MainAxisSize) String() string {
	switch s {
	case MainAxisSizeMin:
		return "min"
	case MainAxisSizeMax:
		return "max"
	default:
		return fmt.Sprintf("MainAxisSize(%d)", int(s))
	}
}

// FlexFactor reports the flex value for a render box.
type FlexFactor interface {
	FlexFactor() int
}

// Row lays out children horizontally from left to right.
//
// Row is a flex container where the main axis is horizontal. Children are
// laid out in a single horizontal run and do not wrap.
//
// # Sizing Behavior
//
// By default (MainAxisSizeMin), Row shrinks to fit its children. Set
// MainAxisSizeMax to expand and fill available horizontal space - this is
// required when using [Expanded] children.
//
// # Alignment
//
// Use MainAxisAlignment to control horizontal spacing (Start, End, Center,
// SpaceBetween, SpaceAround, SpaceEvenly). Use CrossAxisAlignment to control
// vertical alignment (Start, End, Center, Stretch).
//
// # Flexible Children
//
// Wrap children in [Expanded] to make them share remaining space proportionally:
//
//	Row{
//	    MainAxisSize: MainAxisSizeMax,
//	    ChildrenWidgets: []core.Widget{
//	        Text{Content: "Label"},
//	        Expanded{ChildWidget: TextField{...}}, // Takes remaining space
//	        Button{...},
//	    },
//	}
//
// For vertical layout, use [Column].
type Row struct {
	ChildrenWidgets    []core.Widget
	MainAxisAlignment  MainAxisAlignment
	CrossAxisAlignment CrossAxisAlignment
	MainAxisSize       MainAxisSize
}

// RowOf creates a horizontal layout with the specified alignments and sizing behavior.
// This is a convenience helper for the common case of creating a Row with children.
func RowOf(alignment MainAxisAlignment, crossAlignment CrossAxisAlignment, size MainAxisSize, children ...core.Widget) Row {
	return Row{
		ChildrenWidgets:    children,
		MainAxisAlignment:  alignment,
		CrossAxisAlignment: crossAlignment,
		MainAxisSize:       size,
	}
}

func (r Row) CreateElement() core.Element {
	return core.NewRenderObjectElement(r, nil)
}

func (r Row) Key() any {
	return nil
}

func (r Row) Children() []core.Widget {
	return r.ChildrenWidgets
}

func (r Row) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	flex := &renderFlex{
		direction:      AxisHorizontal,
		alignment:      r.MainAxisAlignment,
		crossAlignment: r.CrossAxisAlignment,
		axisSize:       r.MainAxisSize,
	}
	flex.SetSelf(flex)
	return flex
}

func (r Row) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if flex, ok := renderObject.(*renderFlex); ok {
		if flex.direction != AxisHorizontal {
			flex.errorTextLayout = nil // invalidate cached error text
		}
		flex.direction = AxisHorizontal
		flex.alignment = r.MainAxisAlignment
		flex.crossAlignment = r.CrossAxisAlignment
		flex.axisSize = r.MainAxisSize
		flex.MarkNeedsLayout()
		flex.MarkNeedsPaint()
	}
}

// Column lays out children vertically from top to bottom.
//
// Column is a flex container where the main axis is vertical. Children are
// laid out in a single vertical run and do not wrap.
//
// # Sizing Behavior
//
// By default (MainAxisSizeMin), Column shrinks to fit its children. Set
// MainAxisSizeMax to expand and fill available vertical space - this is
// required when using [Expanded] children.
//
// # Alignment
//
// Use MainAxisAlignment to control vertical spacing (Start, End, Center,
// SpaceBetween, SpaceAround, SpaceEvenly). Use CrossAxisAlignment to control
// horizontal alignment (Start, End, Center, Stretch).
//
// # Flexible Children
//
// Wrap children in [Expanded] to make them share remaining space proportionally:
//
//	Column{
//	    MainAxisSize: MainAxisSizeMax,
//	    ChildrenWidgets: []core.Widget{
//	        Text{Content: "Header"},
//	        Expanded{ChildWidget: ListView{...}}, // Takes remaining space
//	        Text{Content: "Footer"},
//	    },
//	}
//
// For horizontal layout, use [Row].
type Column struct {
	ChildrenWidgets    []core.Widget
	MainAxisAlignment  MainAxisAlignment
	CrossAxisAlignment CrossAxisAlignment
	MainAxisSize       MainAxisSize
}

// ColumnOf creates a vertical layout with the specified alignments and sizing behavior.
// This is a convenience helper for the common case of creating a Column with children.
func ColumnOf(alignment MainAxisAlignment, crossAlignment CrossAxisAlignment, size MainAxisSize, children ...core.Widget) Column {
	return Column{
		ChildrenWidgets:    children,
		MainAxisAlignment:  alignment,
		CrossAxisAlignment: crossAlignment,
		MainAxisSize:       size,
	}
}

func (c Column) CreateElement() core.Element {
	return core.NewRenderObjectElement(c, nil)
}

func (c Column) Key() any {
	return nil
}

func (c Column) Children() []core.Widget {
	return c.ChildrenWidgets
}

func (c Column) CreateRenderObject(ctx core.BuildContext) layout.RenderObject {
	flex := &renderFlex{
		direction:      AxisVertical,
		alignment:      c.MainAxisAlignment,
		crossAlignment: c.CrossAxisAlignment,
		axisSize:       c.MainAxisSize,
	}
	flex.SetSelf(flex)
	return flex
}

func (c Column) UpdateRenderObject(ctx core.BuildContext, renderObject layout.RenderObject) {
	if flex, ok := renderObject.(*renderFlex); ok {
		if flex.direction != AxisVertical {
			flex.errorTextLayout = nil // invalidate cached error text
		}
		flex.direction = AxisVertical
		flex.alignment = c.MainAxisAlignment
		flex.crossAlignment = c.CrossAxisAlignment
		flex.axisSize = c.MainAxisSize
		flex.MarkNeedsLayout()
		flex.MarkNeedsPaint()
	}
}

type renderFlex struct {
	layout.RenderBoxBase
	children              []layout.RenderBox
	direction             Axis
	alignment             MainAxisAlignment
	crossAlignment        CrossAxisAlignment
	axisSize              MainAxisSize
	hasUnboundedFlexError bool
	unboundedFlexWarned   bool                 // one-shot flag to avoid log spam
	errorTextLayout       *graphics.TextLayout // cached error text layout
}

func (r *renderFlex) SetChildren(children []layout.RenderObject) {
	// Clear parent on old children
	for _, child := range r.children {
		setParentOnChild(child, nil)
	}
	r.children = r.children[:0]
	for _, child := range children {
		if box, ok := child.(layout.RenderBox); ok {
			r.children = append(r.children, box)
			setParentOnChild(box, r)
		}
	}
}

func (r *renderFlex) VisitChildren(visitor func(layout.RenderObject)) {
	for _, child := range r.children {
		visitor(child)
	}
}

func (r *renderFlex) mainAxis(size graphics.Size) float64 {
	if r.direction == AxisHorizontal {
		return size.Width
	}
	return size.Height
}

func (r *renderFlex) crossAxis(size graphics.Size) float64 {
	if r.direction == AxisHorizontal {
		return size.Height
	}
	return size.Width
}

func (r *renderFlex) makeSize(main, cross float64) graphics.Size {
	if r.direction == AxisHorizontal {
		return graphics.Size{Width: main, Height: cross}
	}
	return graphics.Size{Width: cross, Height: main}
}

func (r *renderFlex) makeOffset(main, cross float64) graphics.Offset {
	if r.direction == AxisHorizontal {
		return graphics.Offset{X: main, Y: cross}
	}
	return graphics.Offset{X: cross, Y: main}
}

func (r *renderFlex) PerformLayout() {
	constraints := r.Constraints()
	maxSize := graphics.Size{Width: constraints.MaxWidth, Height: constraints.MaxHeight}
	maxMain := r.mainAxis(maxSize)

	mainSize := 0.0
	crossSize := 0.0
	totalFlex := 0
	flexChildren := make([]layout.RenderBox, 0)
	flexFactors := make([]int, 0)

	for _, child := range r.children {
		if flex := r.flexFactor(child); flex > 0 {
			flexChildren = append(flexChildren, child)
			flexFactors = append(flexFactors, flex)
			totalFlex += flex
			continue
		}
		child.Layout(r.looseConstraints(maxSize), true) // true: we read child.Size()
		childSize := child.Size()
		mainSize += r.mainAxis(childSize)
		crossSize = math.Max(crossSize, r.crossAxis(childSize))
	}

	// Check for flex children with unbounded main axis
	r.hasUnboundedFlexError = false
	if totalFlex > 0 && maxMain == math.MaxFloat64 {
		if !r.unboundedFlexWarned {
			log.Printf("WARNING: Flex children (Expanded/Flexible) used with unbounded %s axis. "+
				"Flex children cannot flex in unbounded constraints. "+
				"Consider using fixed-size widgets instead.", r.direction.String())
			r.unboundedFlexWarned = true
		}
		r.hasUnboundedFlexError = true

		// Short-circuit: use safe fallback size and skip flex child layout
		// Use minimum constraints or a safe default for the error box
		fallbackMain := math.Max(constraints.MinWidth, 200)
		if r.direction == AxisVertical {
			fallbackMain = math.Max(constraints.MinHeight, 50)
		}
		fallbackCross := crossSize
		if fallbackCross == 0 {
			fallbackCross = 50
		}
		size := constraints.Constrain(r.makeSize(fallbackMain, fallbackCross))
		r.SetSize(size)
		return
	}

	remaining := max(maxMain-mainSize, 0)
	if r.axisSize != MainAxisSizeMax {
		remaining = 0
	}

	for i, child := range flexChildren {
		allocated := 0.0
		if totalFlex > 0 {
			allocated = remaining * float64(flexFactors[i]) / float64(totalFlex)
		}
		// Flex children get tight constraints in the main axis direction
		child.Layout(r.flexConstraints(constraints, allocated), true) // true: we read child.Size()
		childSize := child.Size()
		mainSize += r.mainAxis(childSize)
		crossSize = math.Max(crossSize, r.crossAxis(childSize))
	}

	finalMain := mainSize
	if r.axisSize == MainAxisSizeMax {
		finalMain = maxMain
	}

	size := constraints.Constrain(r.makeSize(finalMain, crossSize))
	r.SetSize(size)

	freeSpace := math.Max(0, r.mainAxis(size)-mainSize)
	spacing, startOffset := r.computeSpacing(freeSpace)

	cursor := startOffset
	for _, child := range r.children {
		crossOffset := r.crossAxisOffset(child.Size())
		child.SetParentData(&layout.BoxParentData{Offset: r.makeOffset(cursor, crossOffset)})
		cursor += r.mainAxis(child.Size()) + spacing
	}
}

func (r *renderFlex) flexFactor(child layout.RenderBox) int {
	if flexChild, ok := child.(FlexFactor); ok {
		return flexChild.FlexFactor()
	}
	return 0
}

func (r *renderFlex) looseConstraints(maxSize graphics.Size) layout.Constraints {
	if r.crossAlignment != CrossAxisAlignmentStretch {
		return layout.Loose(maxSize)
	}
	if r.direction == AxisHorizontal {
		return layout.Constraints{
			MinWidth:  0,
			MaxWidth:  maxSize.Width,
			MinHeight: maxSize.Height,
			MaxHeight: maxSize.Height,
		}
	}
	return layout.Constraints{
		MinWidth:  maxSize.Width,
		MaxWidth:  maxSize.Width,
		MinHeight: 0,
		MaxHeight: maxSize.Height,
	}
}

func (r *renderFlex) crossAxisOffset(childSize graphics.Size) float64 {
	freeSpace := r.crossAxis(r.Size()) - r.crossAxis(childSize)
	if freeSpace <= 0 {
		return 0
	}
	switch r.crossAlignment {
	case CrossAxisAlignmentEnd:
		return freeSpace
	case CrossAxisAlignmentCenter:
		return freeSpace * 0.5
	default:
		return 0
	}
}

func (r *renderFlex) flexConstraints(constraints layout.Constraints, mainSize float64) layout.Constraints {
	if r.direction == AxisHorizontal {
		minHeight := 0.0
		maxHeight := constraints.MaxHeight
		if r.crossAlignment == CrossAxisAlignmentStretch {
			minHeight = maxHeight
		}
		return layout.Constraints{
			MinWidth:  mainSize,
			MaxWidth:  mainSize,
			MinHeight: minHeight,
			MaxHeight: maxHeight,
		}
	}
	minWidth := 0.0
	maxWidth := constraints.MaxWidth
	if r.crossAlignment == CrossAxisAlignmentStretch {
		minWidth = maxWidth
	}
	return layout.Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: mainSize,
		MaxHeight: mainSize,
	}
}

func (r *renderFlex) computeSpacing(freeSpace float64) (spacing, offset float64) {
	n := len(r.children)
	switch r.alignment {
	case MainAxisAlignmentEnd:
		offset = freeSpace
	case MainAxisAlignmentCenter:
		offset = freeSpace * 0.5
	case MainAxisAlignmentSpaceBetween:
		if n > 1 {
			spacing = freeSpace / float64(n-1)
		}
	case MainAxisAlignmentSpaceAround:
		if n > 0 {
			spacing = freeSpace / float64(n)
			offset = spacing * 0.5
		}
	case MainAxisAlignmentSpaceEvenly:
		if n > 0 {
			spacing = freeSpace / float64(n+1)
			offset = spacing
		}
	}
	return
}

func (r *renderFlex) Paint(ctx *layout.PaintContext) {
	if r.hasUnboundedFlexError {
		r.paintErrorBox(ctx)
		return
	}
	for _, child := range r.children {
		ctx.PaintChild(child, getChildOffset(child))
	}
}

func (r *renderFlex) paintErrorBox(ctx *layout.PaintContext) {
	size := r.Size()

	// Draw bright pink background to indicate error
	ctx.Canvas.DrawRect(graphics.RectFromLTWH(0, 0, size.Width, size.Height), graphics.Paint{
		Color: graphics.RGBA(255, 0, 127, 255), // Bright pink #ff007f
		Style: graphics.PaintStyleFill,
	})

	// Draw cached black error text
	if r.errorTextLayout == nil {
		errorMsg := fmt.Sprintf("FLEX ERROR: Expanded in unbounded %s", r.direction.String())
		textStyle := graphics.TextStyle{
			FontSize: 14,
			Color:    graphics.ColorBlack,
		}
		r.errorTextLayout, _ = graphics.LayoutText(errorMsg, textStyle, graphics.DefaultFontManager())
	}
	if r.errorTextLayout != nil {
		ctx.Canvas.DrawText(r.errorTextLayout, graphics.Offset{X: 8, Y: 8})
	}
}

func (r *renderFlex) HitTest(position graphics.Offset, result *layout.HitTestResult) bool {
	if !withinBounds(position, r.Size()) {
		return false
	}
	// Skip children when in error state - they may have stale offsets
	if r.hasUnboundedFlexError {
		result.Add(r)
		return true
	}
	for i := len(r.children) - 1; i >= 0; i-- {
		child := r.children[i]
		offset := getChildOffset(child)
		local := graphics.Offset{X: position.X - offset.X, Y: position.Y - offset.Y}
		if child.HitTest(local, result) {
			return true
		}
	}
	result.Add(r)
	return true
}
