package graphics

import (
	"math"
	"sync"
	"unsafe"
)

// Command buffer opcodes. Values must match the C++ replay interpreter.
const (
	cmdSave            float32 = 1
	cmdRestore         float32 = 2
	cmdSaveLayerAlpha  float32 = 3
	cmdSaveLayerBlur   float32 = 4
	cmdSaveLayer       float32 = 5
	cmdTranslate       float32 = 6
	cmdScale           float32 = 7
	cmdRotate          float32 = 8
	cmdClipRect        float32 = 9
	cmdClipRRect       float32 = 10
	cmdClear           float32 = 11
	cmdDrawRect        float32 = 12
	cmdDrawRRect       float32 = 13
	cmdDrawCircle      float32 = 14
	cmdDrawLine        float32 = 15
	cmdDrawRectShadow  float32 = 16
	cmdDrawRRectShadow float32 = 17
	cmdSVGTinted       float32 = 18
	cmdLottie          float32 = 19
)

// commandBuffer accumulates batchable ops as a flat float32 slice.
type commandBuffer struct {
	data []float32
}

var commandBufferPool = sync.Pool{
	New: func() any {
		return &commandBuffer{data: make([]float32, 0, 1024)}
	},
}

func getCommandBuffer() *commandBuffer {
	buf := commandBufferPool.Get().(*commandBuffer)
	buf.data = buf.data[:0]
	return buf
}

func putCommandBuffer(buf *commandBuffer) {
	if cap(buf.data) > 64*1024 {
		// Don't pool oversized buffers
		return
	}
	commandBufferPool.Put(buf)
}

func (b *commandBuffer) len() int {
	return len(b.data)
}

func (b *commandBuffer) write(v float32) {
	b.data = append(b.data, v)
}

func (b *commandBuffer) writeF64(v float64) {
	b.data = append(b.data, float32(v))
}

func (b *commandBuffer) writeColor(c Color) {
	b.data = append(b.data, math.Float32frombits(uint32(c)))
}

func (b *commandBuffer) writePtr(p unsafe.Pointer) {
	v := uintptr(p)
	b.data = append(b.data, math.Float32frombits(uint32(v)))
	b.data = append(b.data, math.Float32frombits(uint32(v>>32)))
}

func (b *commandBuffer) writeSave() {
	b.write(cmdSave)
}

func (b *commandBuffer) writeRestore() {
	b.write(cmdRestore)
}

func (b *commandBuffer) writeSaveLayerAlpha(bounds Rect, alpha float64) {
	b.write(cmdSaveLayerAlpha)
	b.writeF64(bounds.Left)
	b.writeF64(bounds.Top)
	b.writeF64(bounds.Right)
	b.writeF64(bounds.Bottom)
	b.writeF64(alpha)
}

func (b *commandBuffer) writeSaveLayerBlur(bounds Rect, sigmaX, sigmaY float64) {
	b.write(cmdSaveLayerBlur)
	b.writeF64(bounds.Left)
	b.writeF64(bounds.Top)
	b.writeF64(bounds.Right)
	b.writeF64(bounds.Bottom)
	b.writeF64(sigmaX)
	b.writeF64(sigmaY)
}

func (b *commandBuffer) writeSaveLayer(bounds Rect, paint *Paint) {
	b.write(cmdSaveLayer)
	b.writeF64(bounds.Left)
	b.writeF64(bounds.Top)
	b.writeF64(bounds.Right)
	b.writeF64(bounds.Bottom)

	// Blend mode with default
	blend := int32(paint.BlendMode)
	if blend < 0 || blend > int32(BlendModeLuminosity) {
		blend = int32(BlendModeSrcOver)
	}
	b.write(float32(blend))

	// Alpha with default
	alpha := float32(paint.Alpha)
	if !(alpha >= 0 && alpha <= 1) {
		alpha = 1.0
	}
	b.write(alpha)

	// Color filter encoding
	cfData := encodeColorFilter(paint.ColorFilter)
	b.write(float32(len(cfData)))
	b.data = append(b.data, cfData...)

	// Image filter encoding
	ifData := encodeImageFilter(paint.ImageFilter)
	b.write(float32(len(ifData)))
	b.data = append(b.data, ifData...)
}

func (b *commandBuffer) writeTranslate(dx, dy float64) {
	b.write(cmdTranslate)
	b.writeF64(dx)
	b.writeF64(dy)
}

func (b *commandBuffer) writeScale(sx, sy float64) {
	b.write(cmdScale)
	b.writeF64(sx)
	b.writeF64(sy)
}

func (b *commandBuffer) writeRotate(radians float64) {
	b.write(cmdRotate)
	b.writeF64(radians)
}

func (b *commandBuffer) writeClipRect(rect Rect) {
	b.write(cmdClipRect)
	b.writeF64(rect.Left)
	b.writeF64(rect.Top)
	b.writeF64(rect.Right)
	b.writeF64(rect.Bottom)
}

func (b *commandBuffer) writeClipRRect(rrect RRect) {
	b.write(cmdClipRRect)
	b.writeF64(rrect.Rect.Left)
	b.writeF64(rrect.Rect.Top)
	b.writeF64(rrect.Rect.Right)
	b.writeF64(rrect.Rect.Bottom)
	b.writeF64(rrect.TopLeft.X)
	b.writeF64(rrect.TopLeft.Y)
	b.writeF64(rrect.TopRight.X)
	b.writeF64(rrect.TopRight.Y)
	b.writeF64(rrect.BottomRight.X)
	b.writeF64(rrect.BottomRight.Y)
	b.writeF64(rrect.BottomLeft.X)
	b.writeF64(rrect.BottomLeft.Y)
}

func (b *commandBuffer) writeClear(color Color) {
	b.write(cmdClear)
	b.writeColor(color)
}

// writePaint encodes paint parameters used by draw ops.
// Format: color_bits, style, strokeWidth, cap, join, miter, blend, alpha,
//
//	dash_count, [dash_intervals..., dash_phase],
//	has_gradient(0/1), [gradient_type, x1,y1,x2,y2, cx,cy,radius, stop_count, colors..., positions...]
func (b *commandBuffer) writePaint(paint Paint, shapeBounds Rect) {
	b.writeColor(paint.Color)
	b.write(float32(paint.Style))
	b.writeF64(paint.StrokeWidth)

	cap, join, miter, dash, dashPhase, blend, alpha := paintParams(paint)
	b.write(float32(cap))
	b.write(float32(join))
	b.write(miter)
	b.write(float32(blend))
	b.write(alpha)

	// Dash encoding
	b.write(float32(len(dash)))
	if len(dash) > 0 {
		b.data = append(b.data, dash...)
		b.write(dashPhase)
	}

	// Gradient encoding
	gradientBounds := shapeBounds
	if paint.GradientBounds != nil {
		gradientBounds = *paint.GradientBounds
	}
	if payload, ok := buildGradientPayload(paint.Gradient, gradientBounds); ok {
		b.write(1) // has_gradient
		b.write(float32(payload.gradientType))
		b.writeF64(payload.start.X)
		b.writeF64(payload.start.Y)
		b.writeF64(payload.end.X)
		b.writeF64(payload.end.Y)
		b.writeF64(payload.center.X)
		b.writeF64(payload.center.Y)
		b.writeF64(payload.radius)
		b.write(float32(len(payload.colors)))
		for _, c := range payload.colors {
			b.write(math.Float32frombits(c))
		}
		b.data = append(b.data, payload.positions...)
	} else {
		b.write(0) // no gradient
	}
}

func (b *commandBuffer) writeDrawRect(rect Rect, paint Paint) {
	b.write(cmdDrawRect)
	b.writeF64(rect.Left)
	b.writeF64(rect.Top)
	b.writeF64(rect.Right)
	b.writeF64(rect.Bottom)
	b.writePaint(paint, rect)
}

func (b *commandBuffer) writeDrawRRect(rrect RRect, paint Paint) {
	b.write(cmdDrawRRect)
	b.writeF64(rrect.Rect.Left)
	b.writeF64(rrect.Rect.Top)
	b.writeF64(rrect.Rect.Right)
	b.writeF64(rrect.Rect.Bottom)
	b.writeF64(rrect.TopLeft.X)
	b.writeF64(rrect.TopLeft.Y)
	b.writeF64(rrect.TopRight.X)
	b.writeF64(rrect.TopRight.Y)
	b.writeF64(rrect.BottomRight.X)
	b.writeF64(rrect.BottomRight.Y)
	b.writeF64(rrect.BottomLeft.X)
	b.writeF64(rrect.BottomLeft.Y)
	b.writePaint(paint, rrect.Rect)
}

func (b *commandBuffer) writeDrawCircle(center Offset, radius float64, paint Paint) {
	b.write(cmdDrawCircle)
	b.writeF64(center.X)
	b.writeF64(center.Y)
	b.writeF64(radius)
	bounds := RectFromLTWH(center.X-radius, center.Y-radius, radius*2, radius*2)
	b.writePaint(paint, bounds)
}

func (b *commandBuffer) writeDrawLine(start, end Offset, paint Paint) {
	b.write(cmdDrawLine)
	b.writeF64(start.X)
	b.writeF64(start.Y)
	b.writeF64(end.X)
	b.writeF64(end.Y)
	bounds := Rect{
		Left:   math.Min(start.X, end.X),
		Top:    math.Min(start.Y, end.Y),
		Right:  math.Max(start.X, end.X),
		Bottom: math.Max(start.Y, end.Y),
	}
	b.writePaint(paint, bounds)
}

func (b *commandBuffer) writeDrawRectShadow(rect Rect, shadow BoxShadow) {
	b.write(cmdDrawRectShadow)
	b.writeF64(rect.Left)
	b.writeF64(rect.Top)
	b.writeF64(rect.Right)
	b.writeF64(rect.Bottom)
	b.writeColor(shadow.Color)
	b.writeF64(shadow.Sigma())
	b.writeF64(shadow.Offset.X)
	b.writeF64(shadow.Offset.Y)
	b.writeF64(shadow.Spread)
	b.write(float32(shadow.BlurStyle))
}

func (b *commandBuffer) writeDrawRRectShadow(rrect RRect, shadow BoxShadow) {
	b.write(cmdDrawRRectShadow)
	b.writeF64(rrect.Rect.Left)
	b.writeF64(rrect.Rect.Top)
	b.writeF64(rrect.Rect.Right)
	b.writeF64(rrect.Rect.Bottom)
	b.writeF64(rrect.TopLeft.X)
	b.writeF64(rrect.TopLeft.Y)
	b.writeF64(rrect.TopRight.X)
	b.writeF64(rrect.TopRight.Y)
	b.writeF64(rrect.BottomRight.X)
	b.writeF64(rrect.BottomRight.Y)
	b.writeF64(rrect.BottomLeft.X)
	b.writeF64(rrect.BottomLeft.Y)
	b.writeColor(shadow.Color)
	b.writeF64(shadow.Sigma())
	b.writeF64(shadow.Offset.X)
	b.writeF64(shadow.Offset.Y)
	b.writeF64(shadow.Spread)
	b.write(float32(shadow.BlurStyle))
}

func (b *commandBuffer) writeSVGTinted(svgPtr unsafe.Pointer, bounds Rect, tintColor Color) {
	b.write(cmdSVGTinted)
	b.writePtr(svgPtr)
	b.writeF64(bounds.Left)
	b.writeF64(bounds.Top)
	b.writeF64(bounds.Right)
	b.writeF64(bounds.Bottom)
	b.writeColor(tintColor)
}

func (b *commandBuffer) writeLottie(animPtr unsafe.Pointer, bounds Rect, t float64) {
	b.write(cmdLottie)
	b.writePtr(animPtr)
	b.writeF64(bounds.Left)
	b.writeF64(bounds.Top)
	b.writeF64(bounds.Right)
	b.writeF64(bounds.Bottom)
	b.writeF64(t)
}
