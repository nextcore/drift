package graphics

// paintParams extracts extended paint parameters with defaults applied.
func paintParams(paint Paint) (cap, join int32, miter float32, dash []float32, dashPhase float32, blend int32, alpha float32) {
	cap = int32(paint.StrokeCap)
	join = int32(paint.StrokeJoin)
	miter = float32(paint.MiterLimit)
	if miter == 0 {
		miter = 4.0
	}
	// Validate dash pattern: must have even count >= 2 with finite positive intervals
	if paint.Dash != nil && len(paint.Dash.Intervals) >= 2 && len(paint.Dash.Intervals)%2 == 0 {
		valid := true
		for _, v := range paint.Dash.Intervals {
			if !(v > 0) { // false for NaN, zero, and negative
				valid = false
				break
			}
		}
		if valid {
			dash = make([]float32, len(paint.Dash.Intervals))
			for i, v := range paint.Dash.Intervals {
				dash[i] = float32(v)
			}
			dashPhase = float32(paint.Dash.Phase)
		}
	}
	// Clamp BlendMode to valid range, invalid values default to SrcOver
	blend = int32(paint.BlendMode)
	if blend < 0 || blend > int32(BlendModeLuminosity) {
		blend = int32(BlendModeSrcOver)
	}
	// Clamp Alpha to [0,1]; invalid values (negative, >1, NaN) default to 1.0
	alpha = float32(paint.Alpha)
	if !(alpha >= 0 && alpha <= 1) {
		alpha = 1.0
	}
	return
}
