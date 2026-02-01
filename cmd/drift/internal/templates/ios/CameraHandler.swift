/// CameraHandler.swift
/// Handles camera capture and photo library selection.

import UIKit
import PhotosUI

final class CameraHandler: NSObject {
    static let shared = CameraHandler()

    // Track current request ID for correlation
    private var currentRequestId: String?

    static func handle(method: String, args: Any?) -> (Any?, Error?) {
        let dict = args as? [String: Any] ?? [:]
        switch method {
        case "capturePhoto":
            let useFront = dict["useFrontCamera"] as? Bool ?? false
            let requestId = dict["requestId"] as? String
            DispatchQueue.main.async { shared.openCamera(front: useFront, requestId: requestId) }
            return (nil, nil)
        case "pickFromGallery":
            let multi = dict["allowMultiple"] as? Bool ?? false
            let requestId = dict["requestId"] as? String
            DispatchQueue.main.async { shared.openGallery(multi: multi, requestId: requestId) }
            return (nil, nil)
        default:
            return (nil, NSError(domain: "Camera", code: 404))
        }
    }

    // MARK: - Pickers

    private func openCamera(front: Bool, requestId: String?) {
        guard UIImagePickerController.isSourceTypeAvailable(.camera) else {
            sendResult(type: "capture", image: nil, requestId: requestId, error: "Camera not available")
            return
        }

        currentRequestId = requestId

        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.cameraDevice = front ? .front : .rear
        picker.delegate = self
        present(picker)
    }

    private func openGallery(multi: Bool, requestId: String?) {
        // Note: multi-select is not currently supported on iOS; only the first image is returned.
        _ = multi
        currentRequestId = requestId

        var config = PHPickerConfiguration()
        config.selectionLimit = 1
        config.filter = .images
        let picker = PHPickerViewController(configuration: config)
        picker.delegate = self
        present(picker)
    }

    // MARK: - Helpers

    private func present(_ vc: UIViewController) {
        guard let scene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let root = scene.windows.first?.rootViewController else { return }
        var top = root
        while let presented = top.presentedViewController { top = presented }
        top.present(vc, animated: true)
    }

    private func saveToTemp(_ image: UIImage) -> String? {
        guard let data = image.jpegData(compressionQuality: 0.9) else { return nil }
        let filename = "photo_\(Int(Date().timeIntervalSince1970 * 1000)).jpg"
        let url = FileManager.default.temporaryDirectory.appendingPathComponent(filename)
        do {
            try data.write(to: url)
            return url.path
        } catch {
            return nil
        }
    }

    private func sendResult(type: String, image: UIImage?, requestId: String?, cancelled: Bool = false, error: String? = nil) {
        var payload: [String: Any] = ["type": type, "cancelled": cancelled]

        if let requestId = requestId {
            payload["requestId"] = requestId
        }

        if let error = error {
            payload["error"] = error
        } else if let image = image, let path = saveToTemp(image) {
            payload["media"] = [
                "path": path,
                "width": Int(image.size.width),
                "height": Int(image.size.height),
                "mimeType": "image/jpeg"
            ]
        }

        PlatformChannelManager.shared.sendEvent(channel: "drift/camera/result", data: payload)
    }
}

// MARK: - UIImagePickerControllerDelegate

extension CameraHandler: UIImagePickerControllerDelegate, UINavigationControllerDelegate {
    func imagePickerController(_ picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
        let image = info[.originalImage] as? UIImage
        let requestId = currentRequestId
        currentRequestId = nil
        picker.dismiss(animated: true) { [weak self] in
            if let image = image {
                self?.sendResult(type: "capture", image: image, requestId: requestId)
            } else {
                self?.sendResult(type: "capture", image: nil, requestId: requestId, error: "No image captured")
            }
        }
    }

    func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
        let requestId = currentRequestId
        currentRequestId = nil
        picker.dismiss(animated: true) { [weak self] in
            self?.sendResult(type: "capture", image: nil, requestId: requestId, cancelled: true)
        }
    }
}

// MARK: - PHPickerViewControllerDelegate

extension CameraHandler: PHPickerViewControllerDelegate {
    func picker(_ picker: PHPickerViewController, didFinishPicking results: [PHPickerResult]) {
        let requestId = currentRequestId
        currentRequestId = nil
        picker.dismiss(animated: true)

        guard let provider = results.first?.itemProvider, provider.canLoadObject(ofClass: UIImage.self) else {
            sendResult(type: "gallery", image: nil, requestId: requestId, cancelled: results.isEmpty)
            return
        }

        provider.loadObject(ofClass: UIImage.self) { [weak self] object, error in
            DispatchQueue.main.async {
                if let image = object as? UIImage {
                    self?.sendResult(type: "gallery", image: image, requestId: requestId)
                } else {
                    self?.sendResult(type: "gallery", image: nil, requestId: requestId, error: error?.localizedDescription ?? "Failed to load image")
                }
            }
        }
    }
}
