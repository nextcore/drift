//go:build android || darwin || ios

package graphics

import (
	"github.com/go-drift/drift/pkg/skia"
)

func init() {
	tryBatchReplay = func(d *DisplayList, canvas Canvas) bool {
		sc, ok := canvas.(*SkiaCanvas)
		if !ok {
			return false
		}
		d.replayBatched(sc)
		return true
	}
}

// replayBatched encodes batchable ops into a command buffer and replays them
// via a single CGO call. Non-batchable ops trigger a flush, execute directly
// through the SkiaCanvas methods, then resume encoding.
func (d *DisplayList) replayBatched(sc *SkiaCanvas) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	flush := func() {
		if buf.len() > 0 {
			skia.ReplayCommandBuffer(sc.canvas, buf.data)
			buf.data = buf.data[:0]
		}
	}

	for _, op := range d.ops {
		switch o := op.(type) {
		// State ops
		case opSave:
			buf.writeSave()
		case opRestore:
			buf.writeRestore()
		case opSaveLayerAlpha:
			buf.writeSaveLayerAlpha(o.bounds, o.alpha)
		case opSaveLayerBlur:
			buf.writeSaveLayerBlur(o.bounds, o.sigmaX, o.sigmaY)
		case opSaveLayer:
			if o.paint == nil {
				buf.writeSave()
			} else {
				buf.writeSaveLayer(o.bounds, o.paint)
			}

		// Transform ops
		case opTranslate:
			buf.writeTranslate(o.dx, o.dy)
		case opScale:
			buf.writeScale(o.sx, o.sy)
		case opRotate:
			buf.writeRotate(o.radians)

		// Clip ops
		case opClipRect:
			buf.writeClipRect(o.rect)
		case opClipRRect:
			buf.writeClipRRect(o.rrect)

		// Draw ops
		case opClear:
			buf.writeClear(o.color)
		case opRect:
			buf.writeDrawRect(o.rect, o.paint)
		case opRRect:
			buf.writeDrawRRect(o.rrect, o.paint)
		case opCircle:
			buf.writeDrawCircle(o.center, o.radius, o.paint)
		case opLine:
			buf.writeDrawLine(o.start, o.end, o.paint)
		case opRectShadow:
			buf.writeDrawRectShadow(o.rect, o.shadow)
		case opRRectShadow:
			buf.writeDrawRRectShadow(o.rrect, o.shadow)

		// SVG/Lottie ops (pointer-based, batchable)
		case opSVG:
			buf.writeSVGTinted(o.svgPtr, o.bounds, 0)
		case opSVGTinted:
			buf.writeSVGTinted(o.svgPtr, o.bounds, o.tintColor)
		case opLottie:
			buf.writeLottie(o.animPtr, o.bounds, o.t)

		// Non-batchable ops: flush buffer, execute directly, resume
		case opClipPath:
			flush()
			sc.ClipPath(o.path, o.op, o.antialias)
		case opDrawChildLayer:
			flush()
			o.execute(sc)
		case opText:
			flush()
			sc.DrawText(o.layout, o.position)
		case opImage:
			flush()
			sc.DrawImage(o.image, o.position)
		case opImageRect:
			flush()
			sc.DrawImageRect(o.image, o.srcRect, o.dstRect, o.quality, o.cacheKey)
		case opPath:
			flush()
			sc.DrawPath(o.path, o.paint)

		// Platform view ops: no-op on SkiaCanvas, skip entirely
		case opEmbedPlatformView:
		case opOccludePlatformViews:

		default:
			// Unknown op type: flush and fall back to execute()
			flush()
			op.execute(sc)
		}
	}

	flush()
}
