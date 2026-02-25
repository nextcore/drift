package graphics

import (
	"math"
	"testing"
	"unsafe"
)

func TestCommandBufferSave(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeSave()
	buf.writeRestore()

	if len(buf.data) != 2 {
		t.Fatalf("expected 2 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdSave {
		t.Errorf("expected cmdSave=%v, got %v", cmdSave, buf.data[0])
	}
	if buf.data[1] != cmdRestore {
		t.Errorf("expected cmdRestore=%v, got %v", cmdRestore, buf.data[1])
	}
}

func TestCommandBufferTranslate(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeTranslate(10.5, 20.5)

	if len(buf.data) != 3 {
		t.Fatalf("expected 3 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdTranslate {
		t.Errorf("expected cmdTranslate, got %v", buf.data[0])
	}
	if buf.data[1] != 10.5 {
		t.Errorf("expected dx=10.5, got %v", buf.data[1])
	}
	if buf.data[2] != 20.5 {
		t.Errorf("expected dy=20.5, got %v", buf.data[2])
	}
}

func TestCommandBufferClipRect(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeClipRect(Rect{Left: 1, Top: 2, Right: 3, Bottom: 4})

	if len(buf.data) != 5 {
		t.Fatalf("expected 5 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdClipRect {
		t.Errorf("expected cmdClipRect")
	}
	if buf.data[1] != 1 || buf.data[2] != 2 || buf.data[3] != 3 || buf.data[4] != 4 {
		t.Errorf("unexpected rect values: %v", buf.data[1:5])
	}
}

func TestCommandBufferClipRRect(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	rrect := RRect{
		Rect:        Rect{Left: 0, Top: 0, Right: 100, Bottom: 50},
		TopLeft:     Radius{X: 5, Y: 5},
		TopRight:    Radius{X: 10, Y: 10},
		BottomRight: Radius{X: 15, Y: 15},
		BottomLeft:  Radius{X: 20, Y: 20},
	}
	buf.writeClipRRect(rrect)

	// opcode + 4 rect + 8 radii = 13
	if len(buf.data) != 13 {
		t.Fatalf("expected 13 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdClipRRect {
		t.Errorf("expected cmdClipRRect")
	}
}

func TestCommandBufferClear(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeClear(ColorRed)

	if len(buf.data) != 2 {
		t.Fatalf("expected 2 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdClear {
		t.Errorf("expected cmdClear")
	}
	// Verify color encoding
	decoded := Color(math.Float32bits(buf.data[1]))
	if decoded != ColorRed {
		t.Errorf("expected ColorRed=%#x, got %#x", ColorRed, decoded)
	}
}

func TestCommandBufferDrawRect(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := Paint{
		Color:     ColorBlue,
		Style:     PaintStyleFill,
		BlendMode: BlendModeSrcOver,
		Alpha:     1.0,
	}
	rect := Rect{Left: 10, Top: 20, Right: 110, Bottom: 120}
	buf.writeDrawRect(rect, paint)

	if buf.data[0] != cmdDrawRect {
		t.Errorf("expected cmdDrawRect opcode")
	}
	// Verify rect geometry follows opcode
	if buf.data[1] != 10 || buf.data[2] != 20 || buf.data[3] != 110 || buf.data[4] != 120 {
		t.Errorf("unexpected rect: %v", buf.data[1:5])
	}
}

func TestCommandBufferDrawRectWithGradient(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := Paint{
		Color:     ColorWhite,
		Style:     PaintStyleFill,
		BlendMode: BlendModeSrcOver,
		Alpha:     1.0,
		Gradient: NewLinearGradient(
			AlignCenterLeft, AlignCenterRight,
			[]GradientStop{
				{Position: 0, Color: ColorRed},
				{Position: 1, Color: ColorBlue},
			},
		),
	}
	rect := Rect{Left: 0, Top: 0, Right: 100, Bottom: 100}
	buf.writeDrawRect(rect, paint)

	if buf.data[0] != cmdDrawRect {
		t.Errorf("expected cmdDrawRect opcode")
	}
	// The buffer should be significantly longer with gradient data
	if len(buf.data) < 20 {
		t.Errorf("expected gradient data, buffer only has %d floats", len(buf.data))
	}
}

func TestCommandBufferDrawRectWithDash(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := Paint{
		Color:       ColorBlack,
		Style:       PaintStyleStroke,
		StrokeWidth: 2,
		BlendMode:   BlendModeSrcOver,
		Alpha:       1.0,
		Dash:        &DashPattern{Intervals: []float64{10, 5}, Phase: 2},
	}
	rect := Rect{Left: 0, Top: 0, Right: 50, Bottom: 50}
	buf.writeDrawRect(rect, paint)

	if buf.data[0] != cmdDrawRect {
		t.Errorf("expected cmdDrawRect opcode")
	}
	// Should include dash data
	if len(buf.data) < 15 {
		t.Errorf("expected dash data, buffer only has %d floats", len(buf.data))
	}
}

func TestCommandBufferSaveLayerAlpha(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeSaveLayerAlpha(Rect{Left: 0, Top: 0, Right: 200, Bottom: 200}, 0.5)

	// opcode + 4 bounds + 1 alpha = 6
	if len(buf.data) != 6 {
		t.Fatalf("expected 6 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdSaveLayerAlpha {
		t.Errorf("expected cmdSaveLayerAlpha")
	}
	if buf.data[5] != 0.5 {
		t.Errorf("expected alpha=0.5, got %v", buf.data[5])
	}
}

func TestCommandBufferSaveLayerBlur(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	buf.writeSaveLayerBlur(Rect{Left: 0, Top: 0, Right: 100, Bottom: 100}, 5.0, 10.0)

	// opcode + 4 bounds + 2 sigma = 7
	if len(buf.data) != 7 {
		t.Fatalf("expected 7 floats, got %d", len(buf.data))
	}
	if buf.data[0] != cmdSaveLayerBlur {
		t.Errorf("expected cmdSaveLayerBlur")
	}
	if buf.data[5] != 5.0 || buf.data[6] != 10.0 {
		t.Errorf("expected sigmas 5.0, 10.0, got %v, %v", buf.data[5], buf.data[6])
	}
}

func TestCommandBufferDrawCircle(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := Paint{
		Color:     ColorGreen,
		Style:     PaintStyleFill,
		BlendMode: BlendModeSrcOver,
		Alpha:     1.0,
	}
	buf.writeDrawCircle(Offset{X: 50, Y: 50}, 25, paint)

	if buf.data[0] != cmdDrawCircle {
		t.Errorf("expected cmdDrawCircle opcode")
	}
	if buf.data[1] != 50 || buf.data[2] != 50 || buf.data[3] != 25 {
		t.Errorf("unexpected circle geometry: cx=%v cy=%v r=%v", buf.data[1], buf.data[2], buf.data[3])
	}
}

func TestCommandBufferDrawLine(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := Paint{
		Color:       ColorBlack,
		Style:       PaintStyleStroke,
		StrokeWidth: 1,
		BlendMode:   BlendModeSrcOver,
		Alpha:       1.0,
	}
	buf.writeDrawLine(Offset{X: 0, Y: 0}, Offset{X: 100, Y: 100}, paint)

	if buf.data[0] != cmdDrawLine {
		t.Errorf("expected cmdDrawLine opcode")
	}
	if buf.data[1] != 0 || buf.data[2] != 0 || buf.data[3] != 100 || buf.data[4] != 100 {
		t.Errorf("unexpected line geometry")
	}
}

func TestCommandBufferDrawRectShadow(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	shadow := BoxShadow{
		Color:      ColorBlack,
		Offset:     Offset{X: 2, Y: 4},
		BlurRadius: 10,
		Spread:     1,
		BlurStyle:  BlurStyleOuter,
	}
	buf.writeDrawRectShadow(Rect{Left: 10, Top: 10, Right: 90, Bottom: 90}, shadow)

	if buf.data[0] != cmdDrawRectShadow {
		t.Errorf("expected cmdDrawRectShadow opcode")
	}
	// opcode + 4 rect + color + sigma + dx + dy + spread + blurStyle = 11
	if len(buf.data) != 11 {
		t.Fatalf("expected 11 floats, got %d", len(buf.data))
	}
}

func TestCommandBufferPointerEncoding(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	// Create a test pointer value
	var dummy int
	ptr := unsafe.Pointer(&dummy)
	expected := uintptr(ptr)

	buf.writePtr(ptr)

	if len(buf.data) != 2 {
		t.Fatalf("expected 2 floats for pointer, got %d", len(buf.data))
	}

	lo := math.Float32bits(buf.data[0])
	hi := math.Float32bits(buf.data[1])
	decoded := uintptr(lo) | (uintptr(hi) << 32)
	if decoded != expected {
		t.Errorf("pointer round-trip failed: expected %#x, got %#x", expected, decoded)
	}
}

func TestCommandBufferPoolReuse(t *testing.T) {
	buf1 := getCommandBuffer()
	buf1.writeSave()
	buf1.writeRestore()
	putCommandBuffer(buf1)

	buf2 := getCommandBuffer()
	// Buffer should be reset
	if len(buf2.data) != 0 {
		t.Errorf("expected empty buffer from pool, got %d items", len(buf2.data))
	}
	putCommandBuffer(buf2)
}

func TestCommandBufferSaveLayer(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	paint := &Paint{
		BlendMode: BlendModeSrcOver,
		Alpha:     0.8,
	}
	buf.writeSaveLayer(Rect{Left: 0, Top: 0, Right: 100, Bottom: 100}, paint)

	if buf.data[0] != cmdSaveLayer {
		t.Errorf("expected cmdSaveLayer opcode")
	}
	// opcode + 4 bounds + blend + alpha + cf_len(0) + if_len(0) = 9
	if len(buf.data) != 9 {
		t.Fatalf("expected 9 floats for SaveLayer without filters, got %d", len(buf.data))
	}
}

func TestCommandBufferSaveLayerWithFilters(t *testing.T) {
	buf := getCommandBuffer()
	defer putCommandBuffer(buf)

	cf := ColorFilterGrayscale()
	paint := &Paint{
		BlendMode:   BlendModeSrcOver,
		Alpha:       1.0,
		ColorFilter: &cf,
	}
	buf.writeSaveLayer(Rect{Left: 0, Top: 0, Right: 100, Bottom: 100}, paint)

	if buf.data[0] != cmdSaveLayer {
		t.Errorf("expected cmdSaveLayer opcode")
	}
	// Should be longer than base 9 due to color filter data
	if len(buf.data) <= 9 {
		t.Errorf("expected filter data, buffer only has %d floats", len(buf.data))
	}
}
