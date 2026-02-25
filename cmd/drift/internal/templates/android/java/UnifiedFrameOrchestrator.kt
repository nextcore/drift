/**
 * UnifiedFrameOrchestrator provides deterministic frame-perfect synchronization
 * between Skia content and platform views.
 *
 * All frame operations run synchronously in a single Choreographer callback
 * during the ANIMATION phase (before Android's TRAVERSAL phase). This ensures:
 *
 * 1. Engine step (dispatch, layout, paint) runs on UI thread
 * 2. Platform view geometry is applied immediately (no async posting)
 * 3. Skia rendering submits GPU work from same context
 * 4. Both Skia and platform views hit the same vsync
 */
package {{.PackageName}}

import android.graphics.Path
import android.os.Handler
import android.os.Looper
import android.view.Choreographer
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.util.concurrent.atomic.AtomicBoolean

class UnifiedFrameOrchestrator(
    private val skiaHost: DriftSkiaHost,
    private val overlayController: InputOverlayController
) : Choreographer.FrameCallback {

    @Volatile
    private var active = false

    private val frameScheduled = AtomicBoolean(false)
    private val mainHandler = Handler(Looper.getMainLooper())

    private val postFrameRunnable = Runnable {
        if (active) {
            Choreographer.getInstance().postFrameCallback(this)
        } else {
            frameScheduled.set(false)
        }
    }

    override fun doFrame(frameTimeNanos: Long) {
        frameScheduled.set(false)
        if (!active || !skiaHost.engineReady) return

        val w = skiaHost.surfaceWidth
        val h = skiaHost.surfaceHeight
        if (w <= 0 || h <= 0) return

        // 1. Step engine pipeline, get geometry snapshot
        val snapshotBytes = NativeBridge.stepAndSnapshot(w, h)

        // 2. Apply platform view geometry synchronously on UI thread
        if (snapshotBytes != null) {
            val snapshot = parseSnapshot(snapshotBytes)
            if (snapshot != null) {
                overlayController.applySnapshot(snapshot)
            }
        }

        // 3. Render Skia into HardwareBuffer + present
        skiaHost.renderFrame()

        // 4. Continue animation if needed
        if (NativeBridge.needsFrame() != 0) {
            scheduleFrame()
        }
    }

    private fun parseSnapshot(data: ByteArray): FrameSnapshot? {
        return try {
            val buf = ByteBuffer.wrap(data).order(ByteOrder.LITTLE_ENDIAN)

            // Header: version (uint32) + viewCount (uint32)
            if (buf.remaining() < 8) return null
            val version = buf.getInt().toUInt()
            if (version != 1u) return null
            val viewCount = buf.getInt().toUInt().toInt()

            val views = ArrayList<ViewSnapshot>(viewCount)
            for (i in 0 until viewCount) {
                // Per-view fixed part: 60 bytes
                if (buf.remaining() < 60) return null

                val viewId = buf.getLong()
                val x = buf.getFloat()
                val y = buf.getFloat()
                val width = buf.getFloat()
                val height = buf.getFloat()
                val clipLeft = buf.getFloat()
                val clipTop = buf.getFloat()
                val clipRight = buf.getFloat()
                val clipBottom = buf.getFloat()
                val visibleLeft = buf.getFloat()
                val visibleTop = buf.getFloat()
                val visibleRight = buf.getFloat()
                val visibleBottom = buf.getFloat()
                val flags = buf.get().toInt() and 0xFF
                buf.get() // reserved
                val pathCount = buf.getShort().toInt() and 0xFFFF

                val hasClip = (flags and 1) != 0
                val visible = (flags and 2) != 0

                val occPaths = if (pathCount > 0) {
                    ArrayList<Path>(pathCount).also { list ->
                        for (j in 0 until pathCount) {
                            if (buf.remaining() < 2) return null
                            val cmdCount = buf.getShort().toInt() and 0xFFFF
                            val path = Path()
                            for (k in 0 until cmdCount) {
                                if (buf.remaining() < 2) return null
                                val op = buf.get().toInt() and 0xFF
                                val argCount = buf.get().toInt() and 0xFF
                                if (buf.remaining() < argCount * 4) return null
                                when (op) {
                                    0 -> { // MoveTo
                                        val ax = buf.getFloat()
                                        val ay = buf.getFloat()
                                        path.moveTo(ax, ay)
                                    }
                                    1 -> { // LineTo
                                        val ax = buf.getFloat()
                                        val ay = buf.getFloat()
                                        path.lineTo(ax, ay)
                                    }
                                    2 -> { // QuadTo
                                        val x1 = buf.getFloat()
                                        val y1 = buf.getFloat()
                                        val x2 = buf.getFloat()
                                        val y2 = buf.getFloat()
                                        path.quadTo(x1, y1, x2, y2)
                                    }
                                    3 -> { // CubicTo
                                        val x1 = buf.getFloat()
                                        val y1 = buf.getFloat()
                                        val x2 = buf.getFloat()
                                        val y2 = buf.getFloat()
                                        val x3 = buf.getFloat()
                                        val y3 = buf.getFloat()
                                        path.cubicTo(x1, y1, x2, y2, x3, y3)
                                    }
                                    4 -> { // Close
                                        path.close()
                                    }
                                    else -> {
                                        // Unknown op: skip argCount * 4 bytes
                                        buf.position(buf.position() + argCount * 4)
                                    }
                                }
                            }
                            list.add(path)
                        }
                    }
                } else {
                    emptyList()
                }

                views.add(ViewSnapshot(
                    viewId = viewId,
                    x = x,
                    y = y,
                    width = width,
                    height = height,
                    clipLeft = if (hasClip) clipLeft else null,
                    clipTop = if (hasClip) clipTop else null,
                    clipRight = if (hasClip) clipRight else null,
                    clipBottom = if (hasClip) clipBottom else null,
                    visible = visible,
                    visibleLeft = visibleLeft,
                    visibleTop = visibleTop,
                    visibleRight = visibleRight,
                    visibleBottom = visibleBottom,
                    occlusionPaths = occPaths
                ))
            }
            FrameSnapshot(views)
        } catch (e: Exception) {
            null
        }
    }

    fun scheduleFrame() {
        if (active && frameScheduled.compareAndSet(false, true)) {
            mainHandler.post(postFrameRunnable)
        }
    }

    fun start() {
        active = true
        scheduleFrame()
    }

    fun stop() {
        active = false
        mainHandler.removeCallbacks(postFrameRunnable)
        Choreographer.getInstance().removeFrameCallback(this)
        frameScheduled.set(false)
    }
}

// Data classes for frame snapshot

data class FrameSnapshot(
    val views: List<ViewSnapshot>
)

data class ViewSnapshot(
    val viewId: Long,
    val x: Float,
    val y: Float,
    val width: Float,
    val height: Float,
    val clipLeft: Float?,
    val clipTop: Float?,
    val clipRight: Float?,
    val clipBottom: Float?,
    val visible: Boolean,
    val visibleLeft: Float,
    val visibleTop: Float,
    val visibleRight: Float,
    val visibleBottom: Float,
    val occlusionPaths: List<Path>
)
