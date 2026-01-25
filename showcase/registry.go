package main

import (
	"github.com/go-drift/drift/pkg/core"
)

// Demo represents a showcase demo page.
type Demo struct {
	Route    string
	Title    string
	Subtitle string
	Category string
	Builder  func(ctx core.BuildContext) core.Widget
}

// Category constants for demo organization.
const (
	CategoryWidgets  = "widgets"
	CategoryPlatform = "platform"
)

// demos is the registry of all showcase demo pages.
// Add new demos here to automatically update navigation and routing.
var demos = []Demo{
	// Widget demos
	{"/buttons", "Buttons", "Tappable buttons with haptics", CategoryWidgets, buildButtonsPage},
	{"/forms", "Forms", "Text input and form handling", CategoryWidgets, buildFormsPage},
	{"/layouts", "Layouts", "Row, Column, Stack composition", CategoryWidgets, buildLayoutsPage},
	{"/decorations", "Decorations", "Rounded corners and borders", CategoryWidgets, buildDecorationsPage},
	{"/images", "Images", "PNG/JPG image rendering", CategoryWidgets, buildImagesPage},
	{"/tabs", "Tabs", "Bottom tab bar navigation", CategoryWidgets, buildTabsPage},
	{"/scroll", "Scrolling", "Scrollable lists with physics", CategoryWidgets, buildScrollPage},
	{"/gestures", "Gestures", "Drag gestures with axis locking", CategoryWidgets, buildGesturesPage},
	{"/animations", "Animations", "Implicit animations for smooth UI", CategoryWidgets, buildAnimationsPage},
	{"/error-boundaries", "Error Boundaries", "Graceful error handling in widgets", CategoryWidgets, buildErrorBoundariesPage},

	// Platform demos
	{"/webview", "WebView", "Embedded native browser view", CategoryPlatform, buildWebViewPage},
	{"/notifications", "Notifications", "Permissions and local alerts", CategoryPlatform, buildNotificationsPage},
	{"/secure-storage", "Secure Storage", "Keychain and encrypted storage", CategoryPlatform, buildSecureStoragePage},
	{"/clipboard", "Clipboard", "Copy and paste text data", CategoryPlatform, buildClipboardPage},
	{"/share", "Share", "Share content with other apps", CategoryPlatform, buildSharePage},
	{"/permissions", "Permissions", "Runtime permission management", CategoryPlatform, buildPermissionsPage},
	{"/location", "Location", "GPS and location services", CategoryPlatform, buildLocationPage},
	{"/camera", "Camera", "Photo capture and gallery", CategoryPlatform, buildCameraPage},
	{"/storage", "Storage", "File picker and app directories", CategoryPlatform, buildStoragePage},
}

// demosWithTheming returns demos plus theming page (which needs isDark state).
// This is separate because theming has a different builder signature.
func themingDemo() Demo {
	return Demo{"/theming", "Theming", "Colors, typography, dark mode", CategoryWidgets, nil}
}
