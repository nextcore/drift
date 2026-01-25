/// SecureStorageHandler.swift
/// Handles secure storage using iOS Keychain with optional biometric authentication.

import Foundation
import Security
import LocalAuthentication

enum SecureStorageHandler {
    // MARK: - Error Codes

    private static let errorItemNotFound = "item_not_found"
    private static let errorAuthFailed = "auth_failed"
    private static let errorAuthCancelled = "auth_cancelled"
    private static let errorBiometricNotAvailable = "biometric_not_available"
    private static let errorBiometricNotEnrolled = "biometric_not_enrolled"
    private static let errorDuplicateItem = "duplicate_item"

    // MARK: - Public Interface

    static func handle(method: String, args: Any?) -> (Any?, Error?) {
        switch method {
        case "set":
            return set(args: args)
        case "get":
            return get(args: args)
        case "delete":
            return delete(args: args)
        case "contains":
            return contains(args: args)
        case "getAllKeys":
            return getAllKeys(args: args)
        case "deleteAll":
            return deleteAll(args: args)
        case "isBiometricAvailable":
            return isBiometricAvailable()
        case "getBiometricType":
            return getBiometricType()
        default:
            return (nil, NSError(domain: "SecureStorage", code: 404, userInfo: [NSLocalizedDescriptionKey: "Unknown method: \(method)"]))
        }
    }

    // MARK: - CRUD Operations

    private static func set(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let key = dict["key"] as? String,
              let value = dict["value"] as? String else {
            return (nil, NSError(domain: "SecureStorage", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing key or value"]))
        }

        guard let data = value.data(using: .utf8) else {
            return (nil, NSError(domain: "SecureStorage", code: 400, userInfo: [NSLocalizedDescriptionKey: "Invalid value encoding"]))
        }

        let service = dict["service"] as? String ?? defaultService()
        let accessibility = parseAccessibility(dict["accessibility"] as? String)
        let requireBiometric = dict["requireBiometric"] as? Bool ?? false
        let biometricPrompt = dict["biometricPrompt"] as? String

        // Build base query
        var query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]

        // Add access control for biometric protection
        if requireBiometric {
            var accessError: Unmanaged<CFError>?
            guard let accessControl = SecAccessControlCreateWithFlags(
                nil,
                accessibility,
                .biometryCurrentSet,
                &accessError
            ) else {
                let error = accessError?.takeRetainedValue()
                return (nil, NSError(domain: "SecureStorage", code: -1, userInfo: [
                    NSLocalizedDescriptionKey: error?.localizedDescription ?? "Failed to create access control"
                ]))
            }
            query[kSecAttrAccessControl as String] = accessControl

            // Add context with prompt if provided
            if let prompt = biometricPrompt {
                let context = LAContext()
                context.localizedReason = prompt
                query[kSecUseAuthenticationContext as String] = context
            }
        } else {
            query[kSecAttrAccessible as String] = accessibility
        }

        // Try to delete existing item first
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(deleteQuery as CFDictionary)

        // Add new item
        let status = SecItemAdd(query as CFDictionary, nil)

        if status != errSecSuccess {
            return (nil, mapKeychainError(status))
        }

        return (nil, nil)
    }

    private static func get(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let key = dict["key"] as? String else {
            return (nil, NSError(domain: "SecureStorage", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing key"]))
        }

        let service = dict["service"] as? String ?? defaultService()
        let biometricPrompt = dict["biometricPrompt"] as? String

        var query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        // Always add authentication context with prompt for biometric-protected items
        // Use provided prompt or a default one
        let context = LAContext()
        context.localizedReason = biometricPrompt ?? "Authenticate to access secure data"
        query[kSecUseAuthenticationContext as String] = context

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        if status == errSecItemNotFound {
            return (["value": NSNull()], nil)
        }

        if status != errSecSuccess {
            return (nil, mapKeychainError(status))
        }

        guard let data = result as? Data,
              let value = String(data: data, encoding: .utf8) else {
            return (["value": NSNull()], nil)
        }

        return (["value": value], nil)
    }

    private static func delete(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let key = dict["key"] as? String else {
            return (nil, NSError(domain: "SecureStorage", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing key"]))
        }

        let service = dict["service"] as? String ?? defaultService()

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)

        // Treat "item not found" as success for delete
        if status != errSecSuccess && status != errSecItemNotFound {
            return (nil, mapKeychainError(status))
        }

        return (nil, nil)
    }

    private static func contains(args: Any?) -> (Any?, Error?) {
        guard let dict = args as? [String: Any],
              let key = dict["key"] as? String else {
            return (nil, NSError(domain: "SecureStorage", code: 400, userInfo: [NSLocalizedDescriptionKey: "Missing key"]))
        }

        let service = dict["service"] as? String ?? defaultService()

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnAttributes as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        return (["exists": status == errSecSuccess], nil)
    }

    private static func getAllKeys(args: Any?) -> (Any?, Error?) {
        let dict = args as? [String: Any]
        let service = dict?["service"] as? String ?? defaultService()

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecReturnAttributes as String: true,
            kSecMatchLimit as String: kSecMatchLimitAll
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        if status == errSecItemNotFound {
            return (["keys": [String]()] as [String: Any], nil)
        }

        if status != errSecSuccess {
            return (nil, mapKeychainError(status))
        }

        guard let items = result as? [[String: Any]] else {
            return (["keys": [String]()] as [String: Any], nil)
        }

        let keys = items.compactMap { $0[kSecAttrAccount as String] as? String }
        return (["keys": keys], nil)
    }

    private static func deleteAll(args: Any?) -> (Any?, Error?) {
        let dict = args as? [String: Any]
        let service = dict?["service"] as? String ?? defaultService()

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service
        ]

        let status = SecItemDelete(query as CFDictionary)

        // Treat "item not found" as success for delete
        if status != errSecSuccess && status != errSecItemNotFound {
            return (nil, mapKeychainError(status))
        }

        return (nil, nil)
    }

    // MARK: - Biometric Methods

    private static func isBiometricAvailable() -> (Any?, Error?) {
        let context = LAContext()
        var error: NSError?
        let available = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)
        return (["available": available], nil)
    }

    private static func getBiometricType() -> (Any?, Error?) {
        let context = LAContext()
        var error: NSError?

        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            if let laError = error as? LAError {
                if laError.code == .biometryNotEnrolled {
                    return (["type": "none", "reason": "not_enrolled"], nil)
                }
                if laError.code == .biometryNotAvailable {
                    return (["type": "none", "reason": "not_available"], nil)
                }
            }
            return (["type": "none"], nil)
        }

        switch context.biometryType {
        case .faceID:
            return (["type": "face_id"], nil)
        case .touchID:
            return (["type": "touch_id"], nil)
        case .opticID:
            return (["type": "optic_id"], nil)
        case .none:
            return (["type": "none"], nil)
        @unknown default:
            return (["type": "none"], nil)
        }
    }

    // MARK: - Helpers

    private static func defaultService() -> String {
        return Bundle.main.bundleIdentifier ?? "com.drift.securestorage"
    }

    private static func parseAccessibility(_ value: String?) -> CFString {
        switch value {
        case "when_unlocked":
            return kSecAttrAccessibleWhenUnlocked
        case "after_first_unlock":
            return kSecAttrAccessibleAfterFirstUnlock
        case "when_unlocked_this_device_only":
            return kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        case "after_first_unlock_this_device_only":
            return kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        default:
            return kSecAttrAccessibleWhenUnlocked
        }
    }

    private static func mapKeychainError(_ status: OSStatus) -> Error {
        let code: String
        let message: String

        switch status {
        case errSecItemNotFound:
            code = errorItemNotFound
            message = "Item not found in keychain"
        case errSecUserCanceled:
            code = errorAuthCancelled
            message = "User cancelled authentication"
        case errSecAuthFailed:
            code = errorAuthFailed
            message = "Authentication failed"
        case errSecDuplicateItem:
            code = errorDuplicateItem
            message = "Item already exists"
        case errSecInteractionNotAllowed:
            code = errorAuthFailed
            message = "Interaction not allowed - device may be locked"
        default:
            code = "keychain_error"
            message = "Keychain error: \(status)"
        }

        return NSError(domain: code, code: Int(status), userInfo: [
            NSLocalizedDescriptionKey: message
        ])
    }
}
