/**
 * DriftContainer.kt
 * Custom FrameLayout that enables scrolling through unfocused text inputs.
 *
 * Problem: Platform views (EditText) sit on top of DriftSurfaceView. When an unfocused
 * EditText receives a touch, it consumes it. This prevents scroll gestures that start
 * on a text field from working.
 *
 * Solution: Delay the interception decision until we can distinguish tap vs scroll.
 * - Let EditText see the initial DOWN (so it can handle taps with proper cursor placement)
 * - Track movement; if it exceeds touch slop, intercept and forward to DriftSurfaceView
 * - If touch ends without exceeding slop, EditText handles it as a normal tap
 */
package {{.PackageName}}

import android.content.Context
import android.os.SystemClock
import android.view.MotionEvent
import android.view.View
import android.view.ViewConfiguration
import android.view.ViewGroup
import android.widget.EditText
import android.widget.FrameLayout
import kotlin.math.abs

/**
 * Container that hosts DriftSurfaceView and platform views.
 * Handles touch forwarding for scroll-in-text-field support.
 */
class DriftContainer(context: Context) : FrameLayout(context) {

    private var surfaceView: View? = null
    private val touchSlop = ViewConfiguration.get(context).scaledTouchSlop

    // Touch tracking state
    private var trackedEditText: EditText? = null
    private var touchStartX = 0f
    private var touchStartY = 0f
    private var isForwardingToDrift = false
    private var pendingDownTime = 0L

    /**
     * Set the DriftSurfaceView reference for touch forwarding.
     */
    fun setSurfaceView(view: View) {
        surfaceView = view
    }

    override fun onInterceptTouchEvent(ev: MotionEvent): Boolean {
        when (ev.actionMasked) {
            MotionEvent.ACTION_DOWN -> {
                val hitView = findEditTextAtPosition(ev.x, ev.y)
                if (hitView != null && !hitView.hasFocus()) {
                    // Start tracking, but don't intercept yet - let EditText see the DOWN
                    trackedEditText = hitView
                    touchStartX = ev.x
                    touchStartY = ev.y
                    isForwardingToDrift = false
                    pendingDownTime = ev.downTime
                    return false
                }
                trackedEditText = null
            }

            MotionEvent.ACTION_MOVE -> {
                if (trackedEditText != null && !isForwardingToDrift) {
                    val dx = abs(ev.x - touchStartX)
                    val dy = abs(ev.y - touchStartY)
                    if (dx > touchSlop || dy > touchSlop) {
                        // Movement exceeded slop - this is a scroll, not a tap
                        isForwardingToDrift = true

                        // Send ACTION_CANCEL to the EditText to clean up its state
                        trackedEditText?.let { editText ->
                            // Convert to EditText's local coordinate space
                            val localX = ev.x - editText.left
                            val localY = ev.y - editText.top
                            val cancel = MotionEvent.obtain(
                                pendingDownTime,
                                SystemClock.uptimeMillis(),
                                MotionEvent.ACTION_CANCEL,
                                localX, localY, 0
                            )
                            editText.dispatchTouchEvent(cancel)
                            cancel.recycle()
                        }

                        // Send DOWN to Drift at original position (in surfaceView coordinates)
                        surfaceView?.let { surface ->
                            val downX = touchStartX - surface.left
                            val downY = touchStartY - surface.top
                            val down = MotionEvent.obtain(
                                pendingDownTime,
                                pendingDownTime,
                                MotionEvent.ACTION_DOWN,
                                downX, downY, 0
                            )
                            surface.dispatchTouchEvent(down)
                            down.recycle()
                        }

                        return true // Intercept remaining events
                    }
                }
            }

            MotionEvent.ACTION_UP, MotionEvent.ACTION_CANCEL -> {
                // Touch ended without exceeding slop - it's a tap
                // EditText handles it normally (focus + cursor positioning)
                cleanupTouchState()
            }
        }
        return super.onInterceptTouchEvent(ev)
    }

    override fun onTouchEvent(event: MotionEvent): Boolean {
        if (isForwardingToDrift) {
            // Forward to DriftSurfaceView (adjust coordinates)
            surfaceView?.let { surface ->
                val adjustedX = event.x - surface.left
                val adjustedY = event.y - surface.top
                val adjusted = MotionEvent.obtain(
                    event.downTime,
                    event.eventTime,
                    event.actionMasked,
                    adjustedX, adjustedY,
                    event.metaState
                )
                surface.dispatchTouchEvent(adjusted)
                adjusted.recycle()
            }

            if (event.actionMasked == MotionEvent.ACTION_UP ||
                event.actionMasked == MotionEvent.ACTION_CANCEL) {
                cleanupTouchState()
            }
            return true // Keep receiving events
        }
        return super.onTouchEvent(event)
    }

    private fun cleanupTouchState() {
        trackedEditText = null
        isForwardingToDrift = false
    }

    /**
     * Find an EditText at the given coordinates, or null if none.
     */
    private fun findEditTextAtPosition(x: Float, y: Float): EditText? {
        return findEditTextIn(this, x, y)
    }

    private fun findEditTextIn(parent: ViewGroup, x: Float, y: Float): EditText? {
        // Iterate children in reverse (topmost first) to match touch dispatch order
        for (i in parent.childCount - 1 downTo 0) {
            val child = parent.getChildAt(i)
            if (child.visibility != View.VISIBLE) continue
            if (child === surfaceView) continue // Skip the surface view

            val childX = x - child.left
            val childY = y - child.top
            if (childX >= 0 && childX < child.width && childY >= 0 && childY < child.height) {
                // Found a view at this position
                if (child is EditText) {
                    return child
                }
                // Recurse into ViewGroups
                if (child is ViewGroup) {
                    val found = findEditTextIn(child, childX, childY)
                    if (found != null) return found
                }
            }
        }
        return null
    }
}
