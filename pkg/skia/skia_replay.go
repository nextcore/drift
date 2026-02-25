//go:build android || darwin || ios

package skia

/*
#include "skia_bridge.h"
*/
import "C"

import "unsafe"

// ReplayCommandBuffer replays a batch of encoded drawing commands onto a Skia canvas
// in a single CGO call. The data slice contains sequential opcodes and their payloads.
func ReplayCommandBuffer(canvas unsafe.Pointer, data []float32) {
	if len(data) == 0 || canvas == nil {
		return
	}
	C.drift_skia_replay_command_buffer(
		C.DriftSkiaCanvas(canvas),
		(*C.float)(unsafe.Pointer(&data[0])),
		C.int(len(data)),
	)
}
