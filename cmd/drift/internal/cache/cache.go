// Package cache provides centralized cache directory resolution for Drift.
//
// Priority order: --cache-dir flag > DRIFT_CACHE_DIR env > ~/.drift default.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var global struct {
	version  string
	cacheDir string
}

// SetGlobal initializes the cache resolver with the CLI version.
// This should be called at startup from root.go.
func SetGlobal(version string) {
	global.version = normalizeVersion(version)
}

// normalizeVersion strips prefixes and suffixes to get a clean version
// suitable for cache paths. This ensures the CLI version matches the
// version used by fetch_skia_release.sh.
//
// Note: Dev builds (e.g., "0.1.0-dev") normalize to the release version
// ("v0.1.0") since prebuilt artifacts are only published for releases.
//
// Examples:
//
//	"drift-v0.1.0" -> "v0.1.0"
//	"0.1.0-dev"    -> "v0.1.0"
//	"v0.1.0"       -> "v0.1.0"
//	"0.1.0"        -> "v0.1.0"
func normalizeVersion(version string) string {
	// Strip "drift-" prefix (matches fetch_skia_release.sh behavior)
	version = strings.TrimPrefix(version, "drift-")

	// Strip "-dev" suffix for development builds
	version = strings.TrimSuffix(version, "-dev")

	// Ensure "v" prefix for consistency
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return version
}

// SetCacheDir sets an override for the cache directory.
// This is typically called when parsing the --cache-dir flag.
func SetCacheDir(dir string) {
	global.cacheDir = dir
}

// Root returns the cache root directory.
// Priority: --cache-dir flag > DRIFT_CACHE_DIR env > ~/.drift default.
func Root() (string, error) {
	if global.cacheDir != "" {
		return global.cacheDir, nil
	}

	if envDir := os.Getenv("DRIFT_CACHE_DIR"); envDir != "" {
		return envDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}

	return filepath.Join(home, ".drift"), nil
}

// BuildRoot returns the build cache directory for a module.
// Returns: <cache_root>/build/<module_slug>
func BuildRoot(modulePath string) (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}

	moduleSlug := strings.ReplaceAll(modulePath, "/", "_")
	return filepath.Join(root, "build", moduleSlug), nil
}

// LibDir returns the versioned library directory for prebuilt binaries.
// Returns: <cache_root>/lib/<version>/<platform>/<arch>
func LibDir(platform, arch string) (string, error) {
	if global.version == "" {
		return "", fmt.Errorf("cache version not initialized")
	}

	root, err := Root()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, "lib", global.version, platform, arch), nil
}
