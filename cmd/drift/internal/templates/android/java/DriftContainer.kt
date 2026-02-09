/**
 * DriftContainer.kt
 * FrameLayout that hosts DriftSurfaceView and platform views.
 *
 * Touch interception for platform views is handled by TouchInterceptorView,
 * which wraps each platform view individually.
 */
package {{.PackageName}}

import android.content.Context
import android.widget.FrameLayout

/**
 * Container that hosts DriftSurfaceView and interceptor-wrapped platform views.
 */
class DriftContainer(context: Context) : FrameLayout(context)
