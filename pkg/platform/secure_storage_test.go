package platform

import (
	"testing"
)

func TestSecureStorageOptionsToArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     *SecureStorageOptions
		wantNil  bool
		wantKeys []string
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantNil: true,
		},
		{
			name:     "empty options",
			opts:     &SecureStorageOptions{},
			wantNil:  false,
			wantKeys: []string{},
		},
		{
			name: "with accessibility",
			opts: &SecureStorageOptions{
				KeychainAccessibility: AccessibleWhenUnlocked,
			},
			wantKeys: []string{"accessibility"},
		},
		{
			name: "with biometric",
			opts: &SecureStorageOptions{
				RequireBiometric: true,
				BiometricPrompt:  "Authenticate to access",
			},
			wantKeys: []string{"requireBiometric", "biometricPrompt"},
		},
		{
			name: "with service",
			opts: &SecureStorageOptions{
				Service: "com.myapp.storage",
			},
			wantKeys: []string{"service"},
		},
		{
			name: "all options",
			opts: &SecureStorageOptions{
				KeychainAccessibility:    AccessibleAfterFirstUnlock,
				RequireBiometric: true,
				BiometricPrompt:  "Please authenticate",
				Service:          "myservice",
			},
			wantKeys: []string{"accessibility", "requireBiometric", "biometricPrompt", "service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.opts.toArgs()

			if tt.wantNil {
				if args != nil {
					t.Errorf("expected nil, got %v", args)
				}
				return
			}

			if args == nil {
				t.Fatal("expected non-nil args")
			}

			for _, key := range tt.wantKeys {
				if _, ok := args[key]; !ok {
					t.Errorf("expected key %q in args", key)
				}
			}
		})
	}
}

func TestKeychainAccessibilityConstants(t *testing.T) {
	tests := []struct {
		constant KeychainAccessibility
		expected string
	}{
		{AccessibleWhenUnlocked, "when_unlocked"},
		{AccessibleAfterFirstUnlock, "after_first_unlock"},
		{AccessibleWhenUnlockedThisDeviceOnly, "when_unlocked_this_device_only"},
		{AccessibleAfterFirstUnlockThisDeviceOnly, "after_first_unlock_this_device_only"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestBiometricTypeConstants(t *testing.T) {
	tests := []struct {
		constant BiometricType
		expected string
	}{
		{BiometricTypeNone, "none"},
		{BiometricTypeTouchID, "touch_id"},
		{BiometricTypeFaceID, "face_id"},
		{BiometricTypeFingerprint, "fingerprint"},
		{BiometricTypeFace, "face"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestSecureStorageErrorFormat(t *testing.T) {
	err := &SecureStorageError{
		Code:    SecureStorageErrorItemNotFound,
		Message: "Key not found in storage",
	}

	expected := "item_not_found: Key not found in storage"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestSecureStorageWrapError(t *testing.T) {
	s := &SecureStorageService{
		channel: NewMethodChannel("test/secure_storage"),
		events:  NewEventChannel("test/secure_storage/events"),
	}

	tests := []struct {
		name     string
		err      error
		wantType string
		wantCode string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantType: "",
		},
		{
			name: "channel error with known code",
			err: &ChannelError{
				Code:    SecureStorageErrorItemNotFound,
				Message: "Not found",
			},
			wantType: "*platform.SecureStorageError",
			wantCode: SecureStorageErrorItemNotFound,
		},
		{
			name: "channel error with auth cancelled",
			err: &ChannelError{
				Code:    SecureStorageErrorAuthCancelled,
				Message: "User cancelled",
			},
			wantType: "*platform.SecureStorageError",
			wantCode: SecureStorageErrorAuthCancelled,
		},
		{
			name: "channel error with unknown code",
			err: &ChannelError{
				Code:    "unknown_error",
				Message: "Something happened",
			},
			wantType: "*platform.ChannelError",
		},
		{
			name: "already secure storage error",
			err: &SecureStorageError{
				Code:    "custom",
				Message: "Custom error",
			},
			wantType: "*platform.SecureStorageError",
			wantCode: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.wrapError(tt.err)

			if tt.wantType == "" {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil error")
			}

			typeName := getTypeName(result)
			if typeName != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, typeName)
			}

			if tt.wantCode != "" {
				if sse, ok := result.(*SecureStorageError); ok {
					if sse.Code != tt.wantCode {
						t.Errorf("expected code %q, got %q", tt.wantCode, sse.Code)
					}
				}
			}
		})
	}
}

func getTypeName(v any) string {
	if v == nil {
		return ""
	}
	return "*platform." + getStructName(v)
}

func getStructName(v any) string {
	switch v.(type) {
	case *SecureStorageError:
		return "SecureStorageError"
	case *ChannelError:
		return "ChannelError"
	default:
		return "unknown"
	}
}

func TestSecureStorageServiceInitialization(t *testing.T) {
	// Verify the global SecureStorage service is properly initialized
	if SecureStorage == nil {
		t.Fatal("SecureStorage service is nil")
	}

	if SecureStorage.channel == nil {
		t.Error("SecureStorage.channel is nil")
	}

	if SecureStorage.events == nil {
		t.Error("SecureStorage.events is nil")
	}
}
