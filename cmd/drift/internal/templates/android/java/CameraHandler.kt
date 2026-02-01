/**
 * CameraHandler.kt
 * Handles camera capture and photo library selection for the Drift platform channel.
 */
package {{.PackageName}}

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.graphics.BitmapFactory
import android.net.Uri
import android.provider.MediaStore
import androidx.core.content.FileProvider
import java.io.File
import java.io.FileOutputStream

object CameraHandler {
    private const val REQUEST_CAMERA = 1001
    private const val REQUEST_GALLERY = 1002
    private const val REQUEST_GALLERY_MULTI = 1003

    private var pendingCameraFile: File? = null
    private var pendingType: String = "capture"

    fun handle(context: Context, method: String, args: Any?): Pair<Any?, Exception?> {
        return when (method) {
            "capturePhoto" -> capturePhoto(context, args)
            "pickFromGallery" -> pickFromGallery(context, args)
            else -> Pair(null, IllegalArgumentException("Unknown method: $method"))
        }
    }

    private fun capturePhoto(context: Context, args: Any?): Pair<Any?, Exception?> {
        val activity = PlatformChannelManager.currentActivity()
            ?: return Pair(null, IllegalStateException("No active activity"))

        val argsMap = args as? Map<*, *> ?: emptyMap<String, Any>()
        val useFrontCamera = argsMap["useFrontCamera"] as? Boolean ?: false

        val intent = Intent(MediaStore.ACTION_IMAGE_CAPTURE)

        // Create temp file for camera output
        val photoFile = createTempImageFile(context)
        pendingCameraFile = photoFile
        pendingType = "capture"

        val photoUri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            photoFile
        )
        intent.putExtra(MediaStore.EXTRA_OUTPUT, photoUri)

        // Set camera facing if supported
        if (useFrontCamera) {
            intent.putExtra("android.intent.extras.CAMERA_FACING", 1) // Front camera
            intent.putExtra("android.intent.extras.LENS_FACING_FRONT", 1)
            intent.putExtra("android.intent.extra.USE_FRONT_CAMERA", true)
        }

        intent.addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION)
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)

        try {
            @Suppress("DEPRECATION")
            activity.startActivityForResult(intent, REQUEST_CAMERA)
        } catch (e: Exception) {
            return Pair(null, e)
        }

        return Pair(null, null)
    }

    private fun pickFromGallery(context: Context, args: Any?): Pair<Any?, Exception?> {
        val activity = PlatformChannelManager.currentActivity()
            ?: return Pair(null, IllegalStateException("No active activity"))

        val argsMap = args as? Map<*, *> ?: emptyMap<String, Any>()
        val allowMultiple = argsMap["allowMultiple"] as? Boolean ?: false

        pendingType = "gallery"

        val intent = Intent(Intent.ACTION_GET_CONTENT).apply {
            type = "image/*"
            putExtra(Intent.EXTRA_ALLOW_MULTIPLE, allowMultiple)
            addCategory(Intent.CATEGORY_OPENABLE)
        }

        try {
            val requestCode = if (allowMultiple) REQUEST_GALLERY_MULTI else REQUEST_GALLERY
            @Suppress("DEPRECATION")
            activity.startActivityForResult(Intent.createChooser(intent, "Select Image"), requestCode)
        } catch (e: Exception) {
            return Pair(null, e)
        }

        return Pair(null, null)
    }

    fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?, context: Context) {
        when (requestCode) {
            REQUEST_CAMERA -> handleCameraResult(resultCode)
            REQUEST_GALLERY, REQUEST_GALLERY_MULTI -> handleGalleryResult(resultCode, data, context, requestCode == REQUEST_GALLERY_MULTI)
        }
    }

    private fun handleCameraResult(resultCode: Int) {
        val file = pendingCameraFile
        pendingCameraFile = null

        if (resultCode != Activity.RESULT_OK) {
            file?.delete()
            sendResult(type = "capture", cancelled = true)
            return
        }

        if (file == null || !file.exists()) {
            sendResult(type = "capture", error = "No camera output file")
            return
        }

        try {
            val mediaInfo = getMediaInfoFromFile(file)
            sendResult(type = "capture", media = mediaInfo)
        } catch (e: Exception) {
            file.delete()
            sendResult(type = "capture", error = e.message ?: "Failed to process image")
        }
    }

    private fun handleGalleryResult(resultCode: Int, data: Intent?, context: Context, allowMultiple: Boolean) {
        if (resultCode != Activity.RESULT_OK || data == null) {
            sendResult(type = "gallery", cancelled = true)
            return
        }

        try {
            val mediaList = mutableListOf<Map<String, Any?>>()

            // Check for multiple selection
            val clipData = data.clipData
            if (clipData != null && allowMultiple) {
                for (i in 0 until clipData.itemCount) {
                    val uri = clipData.getItemAt(i).uri
                    val mediaInfo = getMediaInfoFromUri(context, uri)
                    mediaList.add(mediaInfo)
                }
            } else {
                // Single selection
                val uri = data.data
                if (uri != null) {
                    val mediaInfo = getMediaInfoFromUri(context, uri)
                    mediaList.add(mediaInfo)
                }
            }

            if (mediaList.isEmpty()) {
                sendResult(type = "gallery", cancelled = true)
            } else if (mediaList.size == 1) {
                sendResult(type = "gallery", media = mediaList.first())
            } else {
                sendResult(type = "gallery", mediaList = mediaList)
            }
        } catch (e: Exception) {
            sendResult(type = "gallery", error = e.message ?: "Failed to process image")
        }
    }

    private fun getMediaInfoFromUri(context: Context, uri: Uri): Map<String, Any?> {
        // Copy to temp file for reliable access (content:// URIs may be temporary)
        val tempFile = createTempImageFile(context)
        val input = context.contentResolver.openInputStream(uri)
            ?: throw IllegalStateException("Failed to open input stream for URI")

        input.use { stream ->
            FileOutputStream(tempFile).use { output ->
                stream.copyTo(output)
            }
        }

        if (tempFile.length() == 0L) {
            tempFile.delete()
            throw IllegalStateException("Failed to copy image data")
        }

        val mimeType = context.contentResolver.getType(uri) ?: "image/jpeg"
        return getMediaInfoFromFile(tempFile, mimeType)
    }

    private fun getMediaInfoFromFile(file: File, mimeType: String = "image/jpeg"): Map<String, Any?> {
        val path = file.absolutePath
        val size = file.length()

        // Get image dimensions without loading full bitmap
        val options = BitmapFactory.Options().apply {
            inJustDecodeBounds = true
        }
        BitmapFactory.decodeFile(path, options)

        return mapOf(
            "path" to path,
            "mimeType" to mimeType,
            "width" to options.outWidth,
            "height" to options.outHeight,
            "size" to size
        )
    }

    private fun createTempImageFile(context: Context): File {
        val timeStamp = System.currentTimeMillis()
        val fileName = "drift_photo_$timeStamp.jpg"
        val storageDir = context.getExternalFilesDir(null) ?: context.cacheDir
        return File(storageDir, fileName)
    }

    private fun sendResult(
        type: String,
        media: Map<String, Any?>? = null,
        mediaList: List<Map<String, Any?>>? = null,
        cancelled: Boolean = false,
        error: String? = null
    ) {
        val payload = mutableMapOf<String, Any?>(
            "type" to type,
            "cancelled" to cancelled
        )
        if (media != null) {
            payload["media"] = media
        }
        if (mediaList != null && mediaList.isNotEmpty()) {
            payload["mediaList"] = mediaList
        }
        if (error != null) {
            payload["error"] = error
        }
        PlatformChannelManager.sendEvent("drift/camera/result", payload)
    }
}
