package platform

// CapturedMedia represents a captured or selected media file.
type CapturedMedia struct {
	// Path is the absolute file path to the media file in the app's temp directory.
	Path string

	// MimeType is the media type (e.g., "image/jpeg").
	MimeType string

	// Width is the image width in pixels.
	Width int

	// Height is the image height in pixels.
	Height int

	// Size is the file size in bytes.
	Size int64
}

// CameraResult represents the result of a camera or gallery operation.
// Results are delivered asynchronously via CameraResults().
type CameraResult struct {
	// Type indicates the operation: "capture" for camera, "gallery" for picker.
	Type string

	// Media contains the captured/selected media for single-selection operations.
	Media *CapturedMedia

	// MediaList contains multiple selected media when AllowMultiple is true.
	MediaList []CapturedMedia

	// Cancelled is true if the user dismissed the camera/picker without selecting.
	Cancelled bool

	// Error contains the error message if the operation failed.
	Error string
}

// CapturePhotoOptions configures photo capture behavior.
type CapturePhotoOptions struct {
	// UseFrontCamera opens the front-facing camera when true.
	// Note: On Android, this is a hint that not all camera apps honor.
	UseFrontCamera bool
}

// PickFromGalleryOptions configures gallery picker behavior.
type PickFromGalleryOptions struct {
	// AllowMultiple enables selecting multiple images from the gallery.
	// When true on Android, results are returned in CameraResult.MediaList.
	// Note: iOS does not support multi-select; only a single image is returned.
	AllowMultiple bool
}

var cameraService = newCameraService()

type cameraServiceState struct {
	channel  *MethodChannel
	events   *EventChannel
	resultCh chan CameraResult
}

func newCameraService() *cameraServiceState {
	service := &cameraServiceState{
		channel:  NewMethodChannel("drift/camera"),
		events:   NewEventChannel("drift/camera/result"),
		resultCh: make(chan CameraResult, 8),
	}

	service.events.Listen(EventHandler{
		OnEvent: func(data any) {
			result := parseCameraResult(data)
			select {
			case service.resultCh <- result:
			default:
				// Buffer full, drop oldest
				<-service.resultCh
				service.resultCh <- result
			}
		},
	})

	return service
}

// CapturePhoto opens the native camera to capture a photo.
// The operation is asynchronous - listen on CameraResults() for the result.
// The captured image is saved to the app's temp directory as a JPEG.
func CapturePhoto(opts CapturePhotoOptions) error {
	_, err := cameraService.channel.Invoke("capturePhoto", map[string]any{
		"useFrontCamera": opts.UseFrontCamera,
	})
	return err
}

// PickFromGallery opens the native photo picker to select images from the library.
// The operation is asynchronous - listen on CameraResults() for the result.
// Selected images are copied to the app's temp directory for reliable access.
func PickFromGallery(opts PickFromGalleryOptions) error {
	_, err := cameraService.channel.Invoke("pickFromGallery", map[string]any{
		"allowMultiple": opts.AllowMultiple,
	})
	return err
}

// CameraResults returns a channel that receives camera and gallery operation results.
// Listen on this channel from a goroutine and use drift.Dispatch to update UI.
// The channel is buffered (capacity 8) and drops oldest results if full.
func CameraResults() <-chan CameraResult {
	return cameraService.resultCh
}

func parseCameraResult(data any) CameraResult {
	m, ok := data.(map[string]any)
	if !ok {
		return CameraResult{Error: "invalid result format"}
	}

	result := CameraResult{
		Type:      parseString(m["type"]),
		Cancelled: parseBool(m["cancelled"]),
		Error:     parseString(m["error"]),
	}

	if mediaData, ok := m["media"].(map[string]any); ok {
		result.Media = parseCapturedMedia(mediaData)
	}

	if mediaList, ok := m["mediaList"].([]any); ok {
		for _, item := range mediaList {
			if itemMap, ok := item.(map[string]any); ok {
				if media := parseCapturedMedia(itemMap); media != nil {
					result.MediaList = append(result.MediaList, *media)
				}
			}
		}
	}

	return result
}

func parseCapturedMedia(m map[string]any) *CapturedMedia {
	path := parseString(m["path"])
	if path == "" {
		return nil
	}
	return &CapturedMedia{
		Path:     path,
		MimeType: parseString(m["mimeType"]),
		Width:    parseInt(m["width"]),
		Height:   parseInt(m["height"]),
		Size:     parseInt64(m["size"]),
	}
}

func parseInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}
