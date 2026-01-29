package main

import (
	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/drift"
	"github.com/go-drift/drift/pkg/platform"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

// buildSecureStoragePage creates a stateful widget for secure storage demos.
func buildSecureStoragePage(ctx core.BuildContext) core.Widget {
	return secureStoragePage{}
}

type secureStoragePage struct{}

func (s secureStoragePage) CreateElement() core.Element {
	return core.NewStatefulElement(s, nil)
}

func (s secureStoragePage) Key() any {
	return nil
}

func (s secureStoragePage) CreateState() core.State {
	return &secureStorageState{}
}

type secureStorageState struct {
	core.StateBase
	statusText      *core.ManagedState[string]
	biometricStatus *core.ManagedState[string]
	storedValue     *core.ManagedState[string]
	keyController   *platform.TextEditingController
	valueController *platform.TextEditingController
}

func (s *secureStorageState) InitState() {
	s.statusText = core.NewManagedState(&s.StateBase, "Enter a key and value to store securely.")
	s.biometricStatus = core.NewManagedState(&s.StateBase, "Checking biometric availability...")
	s.storedValue = core.NewManagedState(&s.StateBase, "")
	s.keyController = platform.NewTextEditingController("demo_key")
	s.valueController = platform.NewTextEditingController("secret_value_123")

	// Check biometric availability on init
	go func() {
		available, err := platform.SecureStorage.IsBiometricAvailable()
		if err != nil {
			drift.Dispatch(func() {
				s.biometricStatus.Set("Error checking biometrics: " + err.Error())
			})
			return
		}

		biometricType, _ := platform.SecureStorage.GetBiometricType()
		var message string
		if available {
			message = "Biometric available: " + string(biometricType)
		} else {
			message = "Biometric not available"
		}
		drift.Dispatch(func() {
			s.biometricStatus.Set(message)
		})
	}()

	// Listen for async biometric auth results
	go func() {
		for event := range platform.SecureStorage.Listen() {
			drift.Dispatch(func() {
				if event.Success {
					if event.Value != "" {
						s.storedValue.Set(event.Value)
						s.statusText.Set("Retrieved value with biometric auth")
					} else {
						s.statusText.Set("Operation completed with biometric auth")
					}
				} else {
					s.statusText.Set("Auth failed: " + event.Error)
				}
			})
		}
	}()
}

func (s *secureStorageState) Build(ctx core.BuildContext) core.Widget {
	_, colors, _ := theme.UseTheme(ctx)

	return demoPage(ctx, "Secure Storage",
		// Biometric status section
		sectionTitle("Biometric Status", colors),
		widgets.VSpace(12),
		statusCard(s.biometricStatus.Get(), colors),
		widgets.VSpace(24),

		// Input section
		sectionTitle("Store Value", colors),
		widgets.VSpace(12),
		widgets.TextField{
			Label:        "Key",
			Controller:   s.keyController,
			Placeholder:  "Enter key name",
			KeyboardType: platform.KeyboardTypeText,
			BorderRadius: 8,
		},
		widgets.VSpace(12),
		widgets.TextField{
			Label:        "Value",
			Controller:   s.valueController,
			Placeholder:  "Enter secret value",
			KeyboardType: platform.KeyboardTypeText,
			BorderRadius: 8,
		},
		widgets.VSpace(16),

		// Action buttons
		widgets.Button{
			Label: "Save to Secure Storage",
			OnTap: func() {
				s.saveValue()
			},
			Color:     colors.Primary,
			TextColor: colors.OnPrimary,
			Haptic:    true,
		},
		widgets.VSpace(12),

		widgets.Button{
			Label: "Save with Biometric",
			OnTap: func() {
				s.saveWithBiometric()
			},
			Color:     colors.Secondary,
			TextColor: colors.OnSecondary,
			Haptic:    true,
		},
		widgets.VSpace(16),

		statusCard(s.statusText.Get(), colors),
		widgets.VSpace(24),

		// Retrieve section
		sectionTitle("Retrieve Value", colors),
		widgets.VSpace(12),
		widgets.Button{
			Label: "Get from Secure Storage",
			OnTap: func() {
				s.getValue()
			},
			Color:     colors.Primary,
			TextColor: colors.OnPrimary,
			Haptic:    true,
		},
		widgets.VSpace(12),

		widgets.Button{
			Label: "Check if Key Exists",
			OnTap: func() {
				s.checkExists()
			},
			Color:     colors.Tertiary,
			TextColor: colors.OnTertiary,
			Haptic:    true,
		},
		widgets.VSpace(12),

		retrievedValueCard(s.storedValue.Get(), colors),
		widgets.VSpace(24),

		// Management section
		sectionTitle("Manage Storage", colors),
		widgets.VSpace(12),
		widgets.Button{
			Label: "Delete Key",
			OnTap: func() {
				s.deleteValue()
			},
			Color:     colors.Error,
			TextColor: colors.OnError,
			Haptic:    true,
		},
		widgets.VSpace(12),

		widgets.Button{
			Label: "List All Keys",
			OnTap: func() {
				s.listKeys()
			},
			Color:     colors.SurfaceContainerHigh,
			TextColor: colors.OnSurface,
			Haptic:    true,
		},
		widgets.VSpace(40),
	)
}

func (s *secureStorageState) saveValue() {
	key := s.keyController.Text()
	value := s.valueController.Text()

	if key == "" || value == "" {
		s.statusText.Set("Please enter both key and value")
		return
	}

	err := platform.SecureStorage.Set(key, value, nil)
	if err != nil {
		s.statusText.Set("Error saving: " + err.Error())
		return
	}

	s.statusText.Set("Value saved securely for key: " + key)
}

func (s *secureStorageState) saveWithBiometric() {
	key := s.keyController.Text()
	value := s.valueController.Text()

	if key == "" || value == "" {
		s.statusText.Set("Please enter both key and value")
		return
	}

	err := platform.SecureStorage.Set(key, value, &platform.SecureStorageOptions{
		RequireBiometric: true,
		BiometricPrompt:  "Authenticate to save securely",
	})
	if err == platform.ErrAuthPending {
		s.statusText.Set("Authenticating...")
		return
	}
	if err != nil {
		s.statusText.Set("Error saving: " + err.Error())
		return
	}

	s.statusText.Set("Value saved with biometric protection for key: " + key)
}

func (s *secureStorageState) getValue() {
	key := s.keyController.Text()

	if key == "" {
		s.statusText.Set("Please enter a key name")
		return
	}

	value, err := platform.SecureStorage.Get(key, nil)
	if err == platform.ErrAuthPending {
		s.statusText.Set("Authenticating...")
		return
	}
	if err != nil {
		s.statusText.Set("Error retrieving: " + err.Error())
		s.storedValue.Set("")
		return
	}

	if value == "" {
		s.statusText.Set("No value found for key: " + key)
		s.storedValue.Set("")
		return
	}

	s.storedValue.Set(value)
	s.statusText.Set("Retrieved value for key: " + key)
}

func (s *secureStorageState) checkExists() {
	key := s.keyController.Text()

	if key == "" {
		s.statusText.Set("Please enter a key name")
		return
	}

	exists, err := platform.SecureStorage.Contains(key, nil)
	if err != nil {
		s.statusText.Set("Error checking: " + err.Error())
		return
	}

	if exists {
		s.statusText.Set("Key exists: " + key)
	} else {
		s.statusText.Set("Key does not exist: " + key)
	}
}

func (s *secureStorageState) deleteValue() {
	key := s.keyController.Text()

	if key == "" {
		s.statusText.Set("Please enter a key name")
		return
	}

	err := platform.SecureStorage.Delete(key, nil)
	if err != nil {
		s.statusText.Set("Error deleting: " + err.Error())
		return
	}

	s.storedValue.Set("")
	s.statusText.Set("Deleted key: " + key)
}

func (s *secureStorageState) listKeys() {
	keys, err := platform.SecureStorage.GetAllKeys(nil)
	if err != nil {
		s.statusText.Set("Error listing keys: " + err.Error())
		return
	}

	if len(keys) == 0 {
		s.statusText.Set("No keys stored")
		return
	}

	message := "Stored keys: "
	for i, key := range keys {
		if i > 0 {
			message += ", "
		}
		message += key
	}
	s.statusText.Set(message)
}

func retrievedValueCard(value string, colors theme.ColorScheme) core.Widget {
	displayValue := value
	if displayValue == "" {
		displayValue = "(no value retrieved)"
	}

	return widgets.Container{
		Color: colors.SurfaceVariant,
		ChildWidget: widgets.PaddingAll(12,
			widgets.ColumnOf(
				widgets.MainAxisAlignmentStart,
				widgets.CrossAxisAlignmentStart,
				widgets.MainAxisSizeMin,
				widgets.Text{Content: "Retrieved Value:", Style: graphics.TextStyle{
					Color:      colors.OnSurfaceVariant,
					FontSize:   12,
					FontWeight: graphics.FontWeightBold,
				}},
				widgets.VSpace(4),
				widgets.Text{Content: displayValue, Style: graphics.TextStyle{
					Color:    colors.OnSurface,
					FontSize: 16,
				}},
			),
		),
	}
}
