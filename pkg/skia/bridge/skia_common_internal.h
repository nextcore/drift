// Backend-provided functions for the shared Skia bridge.
// Each backend (skia_metal.mm, skia_vk.cc) defines these with external linkage.
// skia_common.cc calls them to access platform-specific font management.

#ifndef DRIFT_SKIA_COMMON_INTERNAL_H
#define DRIFT_SKIA_COMMON_INTERNAL_H

#include "core/SkFontMgr.h"

// Returns the platform font manager (Core Text on Apple, Android NDK on Android).
sk_sp<SkFontMgr> drift_get_font_manager();

// Returns the platform fallback font name ("SF Pro Text" on Apple, "sans-serif" on Android).
const char* drift_platform_fallback_font();

#endif  // DRIFT_SKIA_COMMON_INTERNAL_H
