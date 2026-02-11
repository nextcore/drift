/// PermissionHandler.swift
/// Handles runtime permission requests for the Drift platform channel.

import UIKit
import AVFoundation
import Photos
import CoreLocation
import Contacts
import EventKit
import UserNotifications

enum PermissionHandler {
    private static var locationManager: CLLocationManager?
    private static var locationDelegate: LocationPermissionDelegate?

    private static func getLocationManager() -> CLLocationManager {
        if let manager = locationManager {
            return manager
        }

        if Thread.isMainThread {
            let manager = CLLocationManager()
            locationManager = manager
            return manager
        }

        var manager: CLLocationManager?
        DispatchQueue.main.sync {
            let created = CLLocationManager()
            locationManager = created
            manager = created
        }
        return manager ?? CLLocationManager()
    }

    static func handle(method: String, args: Any?) -> (Any?, Error?) {
        switch method {
        case "check":
            return check(args: args)
        case "request":
            return request(args: args)
        case "requestMultiple":
            return requestMultiple(args: args)
        case "openSettings":
            return openSettings()
        case "shouldShowRationale":
            // iOS doesn't have this concept
            return (["shouldShow": false], nil)
        default:
            return (nil, NSError(domain: "Permissions", code: 404, userInfo: [NSLocalizedDescriptionKey: "Unknown method: \(method)"]))
        }
    }

    private static func check(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let permission = dict["permission"] as? String else {
            return (nil, NSError(domain: "Permissions", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing permission"]))
        }

        let status = checkPermissionStatus(permission)
        return (["status": status], nil)
    }

    private static func request(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let permission = dict["permission"] as? String else {
            return (nil, NSError(domain: "Permissions", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing permission"]))
        }

        let currentStatus = checkPermissionStatus(permission)
        requestPermission(permission, args: dict)
        return (["status": currentStatus], nil)
    }

    private static func requestMultiple(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let permissions = dict["permissions"] as? [String] else {
            return (nil, NSError(domain: "Permissions", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing permissions"]))
        }

        var results: [String: String] = [:]
        for permission in permissions {
            results[permission] = checkPermissionStatus(permission)
            requestPermission(permission)
        }

        return (["results": results], nil)
    }

    private static func openSettings() -> (Any?, Error?) {
        DispatchQueue.main.async {
            if let url = URL(string: UIApplication.openSettingsURLString) {
                UIApplication.shared.open(url)
            }
        }
        return (nil, nil)
    }

    private static func checkPermissionStatus(_ permission: String) -> String {
        switch permission {
        case "camera":
            return cameraStatus()
        case "microphone":
            return microphoneStatus()
        case "photos":
            return photosStatus()
        case "location":
            return locationStatus(always: false)
        case "location_always":
            return locationStatus(always: true)
        case "contacts":
            return contactsStatus()
        case "calendar":
            return calendarStatus()
        case "notifications":
            return notificationsStatus()
        default:
            return "unknown"
        }
    }

    private static func requestPermission(_ permission: String, args: [String: Any]? = nil) {
        switch permission {
        case "camera":
            requestCamera()
        case "microphone":
            requestMicrophone()
        case "photos":
            requestPhotos()
        case "location":
            requestLocation(always: false)
        case "location_always":
            requestLocation(always: true)
        case "contacts":
            requestContacts()
        case "calendar":
            requestCalendar()
        case "notifications":
            requestNotifications(args: args)
        default:
            break
        }
    }

    // MARK: - Camera

    private static func cameraStatus() -> String {
        switch AVCaptureDevice.authorizationStatus(for: .video) {
        case .authorized:
            return "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestCamera() {
        AVCaptureDevice.requestAccess(for: .video) { granted in
            // Use permanently_denied for consistency with cameraStatus()
            let status = granted ? "granted" : "permanently_denied"
            sendPermissionChange("camera", status: status)
        }
    }

    // MARK: - Microphone

    private static func microphoneStatus() -> String {
        switch AVCaptureDevice.authorizationStatus(for: .audio) {
        case .authorized:
            return "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestMicrophone() {
        AVCaptureDevice.requestAccess(for: .audio) { granted in
            // Use permanently_denied for consistency with microphoneStatus()
            let status = granted ? "granted" : "permanently_denied"
            sendPermissionChange("microphone", status: status)
        }
    }

    // MARK: - Photos

    private static func photosStatus() -> String {
        let status = PHPhotoLibrary.authorizationStatus(for: .readWrite)

        switch status {
        case .authorized:
            return "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        case .limited:
            return "limited"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestPhotos() {
        PHPhotoLibrary.requestAuthorization(for: .readWrite) { status in
            let statusStr: String
            switch status {
            case .authorized:
                statusStr = "granted"
            case .limited:
                statusStr = "limited"
            case .restricted:
                statusStr = "restricted"
            default:
                // Use permanently_denied for consistency with photosStatus()
                statusStr = "permanently_denied"
            }
            sendPermissionChange("photos", status: statusStr)
        }
    }

    // MARK: - Location

    private static func locationStatus(always: Bool) -> String {
        let status = getLocationManager().authorizationStatus

        switch status {
        case .authorizedAlways:
            return "granted"
        case .authorizedWhenInUse:
            return always ? "denied" : "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestLocation(always: Bool) {
        let manager = getLocationManager()
        locationDelegate = LocationPermissionDelegate(always: always)
        manager.delegate = locationDelegate

        if always {
            manager.requestAlwaysAuthorization()
        } else {
            manager.requestWhenInUseAuthorization()
        }
    }

    // MARK: - Contacts

    private static func contactsStatus() -> String {
        switch CNContactStore.authorizationStatus(for: .contacts) {
        case .authorized:
            return "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        case .limited:
            return "limited"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestContacts() {
        CNContactStore().requestAccess(for: .contacts) { granted, _ in
            // Use permanently_denied for consistency with contactsStatus()
            let status = granted ? "granted" : "permanently_denied"
            sendPermissionChange("contacts", status: status)
        }
    }

    // MARK: - Calendar

    private static func calendarStatus() -> String {
        switch EKEventStore.authorizationStatus(for: .event) {
        case .authorized:
            return "granted"
        case .denied:
            return "permanently_denied"
        case .restricted:
            return "restricted"
        case .notDetermined:
            return "not_determined"
        case .fullAccess:
            return "granted"
        case .writeOnly:
            return "limited"
        @unknown default:
            return "unknown"
        }
    }

    private static func requestCalendar() {
        if #available(iOS 17.0, *) {
            EKEventStore().requestFullAccessToEvents { granted, _ in
                // Use permanently_denied for consistency with calendarStatus()
                let status = granted ? "granted" : "permanently_denied"
                sendPermissionChange("calendar", status: status)
            }
        } else {
            EKEventStore().requestAccess(to: .event) { granted, _ in
                // Use permanently_denied for consistency with calendarStatus()
                let status = granted ? "granted" : "permanently_denied"
                sendPermissionChange("calendar", status: status)
            }
        }
    }

    // MARK: - Notifications

    private static func notificationsStatus() -> String {
        var status = "unknown"
        let semaphore = DispatchSemaphore(value: 0)

        UNUserNotificationCenter.current().getNotificationSettings { settings in
            switch settings.authorizationStatus {
            case .authorized:
                status = "granted"
            case .denied:
                status = "permanently_denied"
            case .notDetermined:
                status = "not_determined"
            case .provisional:
                status = "provisional"
            case .ephemeral:
                status = "provisional"
            @unknown default:
                status = "unknown"
            }
            semaphore.signal()
        }

        semaphore.wait()
        return status
    }

    private static func requestNotifications(args: [String: Any]? = nil) {
        var authOptions: UNAuthorizationOptions = []

        // Parse notification options from args
        let alertEnabled = args?["alert"] as? Bool ?? true
        let soundEnabled = args?["sound"] as? Bool ?? true
        let badgeEnabled = args?["badge"] as? Bool ?? true
        let provisionalEnabled = args?["provisional"] as? Bool ?? false

        if alertEnabled {
            authOptions.insert(.alert)
        }
        if soundEnabled {
            authOptions.insert(.sound)
        }
        if badgeEnabled {
            authOptions.insert(.badge)
        }
        if provisionalEnabled {
            authOptions.insert(.provisional)
        }

        UNUserNotificationCenter.current().requestAuthorization(options: authOptions) { granted, _ in
            // Get the actual status after authorization
            UNUserNotificationCenter.current().getNotificationSettings { settings in
                let status: String
                switch settings.authorizationStatus {
                case .authorized:
                    status = "granted"
                case .provisional:
                    status = "provisional"
                case .denied:
                    // Use permanently_denied for consistency with notificationsStatus()
                    status = "permanently_denied"
                default:
                    status = granted ? "granted" : "permanently_denied"
                }
                sendPermissionChange("notifications", status: status)
            }
        }
    }

    // MARK: - Helpers

    private static func sendPermissionChange(_ permission: String, status: String) {
        DispatchQueue.main.async {
            PlatformChannelManager.shared.sendEvent(channel: "drift/permissions/changes", data: [
                "permission": permission,
                "status": status
            ])
        }
    }
}

// MARK: - Location Permission Delegate

private class LocationPermissionDelegate: NSObject, CLLocationManagerDelegate {
    private let always: Bool

    init(always: Bool) {
        self.always = always
    }

    func locationManagerDidChangeAuthorization(_ manager: CLLocationManager) {
        handleAuthorizationStatus(manager.authorizationStatus)
    }

    func locationManager(_ manager: CLLocationManager, didChangeAuthorization status: CLAuthorizationStatus) {
        handleAuthorizationStatus(status)
    }

    private func handleAuthorizationStatus(_ status: CLAuthorizationStatus) {
        let permission = always ? "location_always" : "location"
        let statusText: String

        switch status {
        case .authorizedAlways:
            statusText = "granted"
        case .authorizedWhenInUse:
            statusText = always ? "denied" : "granted"
        case .denied:
            statusText = "permanently_denied"
        case .restricted:
            statusText = "restricted"
        case .notDetermined:
            return
        @unknown default:
            statusText = "unknown"
        }

        DispatchQueue.main.async {
            PlatformChannelManager.shared.sendEvent(channel: "drift/permissions/changes", data: [
                "permission": permission,
                "status": statusText
            ])
        }
    }
}
