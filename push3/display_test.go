package push3

import "testing"

func TestRGB565(t *testing.T) {
	// Push 3 uses BGR565LE: LSB = GGGRRRRR, MSB = BBBBBGGG.
	tests := []struct {
		name    string
		r, g, b uint8
		wantLo  byte
		wantHi  byte
	}{
		{"black", 0, 0, 0, 0x00, 0x00},
		{"white", 255, 255, 255, 0xFF, 0xFF},
		{"red", 255, 0, 0, 0x1F, 0x00},
		{"green", 0, 255, 0, 0xE0, 0x07},
		{"blue", 0, 0, 255, 0x00, 0xF8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi := rgb565(tt.r, tt.g, tt.b)
			if lo != tt.wantLo || hi != tt.wantHi {
				t.Errorf("rgb565(%d,%d,%d) = (0x%02X, 0x%02X), want (0x%02X, 0x%02X)",
					tt.r, tt.g, tt.b, lo, hi, tt.wantLo, tt.wantHi)
			}
		})
	}
}

func TestXOREncode(t *testing.T) {
	// Verify the XOR mask applied to a known sequence.
	input := []byte{0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}
	want := []byte{0xE7, 0xF3, 0xE7, 0xFF, 0x18, 0x0C, 0x18, 0x00}

	got := make([]byte, len(input))
	for i := range input {
		got[i] = input[i] ^ xorMask[i&3]
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("byte %d: got 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestFrameLayout(t *testing.T) {
	t.Run("sizes", func(t *testing.T) {
		if frameSize != 327696 {
			t.Errorf("frameSize = %d, want 327696", frameSize)
		}
		if frameBufSize != 307200 {
			t.Errorf("frameBufSize = %d, want 307200", frameBufSize)
		}
		if bytesPerLine != 2048 {
			t.Errorf("bytesPerLine = %d, want 2048", bytesPerLine)
		}
	})

	t.Run("header", func(t *testing.T) {
		want := [16]byte{0xFF, 0xCC, 0xAA, 0x88}
		if frameHeader != want {
			t.Errorf("frameHeader = %X, want %X", frameHeader, want)
		}
	})
}

func TestSetPixel(t *testing.T) {
	d := &Display{
		frame: make([]byte, frameBufSize),
	}

	// Set top-left corner to red → BGR565: lo=0x1F, hi=0x00.
	d.SetPixel(0, 0, 255, 0, 0)
	if d.frame[0] != 0x1F || d.frame[1] != 0x00 {
		t.Errorf("(0,0) red = [0x%02X, 0x%02X], want [0x1F, 0x00]", d.frame[0], d.frame[1])
	}

	// Set bottom-right corner to blue → BGR565: lo=0x00, hi=0xF8.
	d.SetPixel(959, 159, 0, 0, 255)
	off := (159*960 + 959) * 2
	if d.frame[off] != 0x00 || d.frame[off+1] != 0xF8 {
		t.Errorf("(959,159) blue = [0x%02X, 0x%02X], want [0x00, 0xF8]", d.frame[off], d.frame[off+1])
	}
}
