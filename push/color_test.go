package push

import (
	"testing"
)

func TestEncodePaletteColor(t *testing.T) {
	tests := []struct {
		name    string
		value   uint8
		wantLSB uint8
		wantMSB uint8
	}{
		{"zero", 0, 0, 0},
		{"max 7-bit", 127, 127, 0},
		{"bit 7 set", 128, 0, 1},
		{"all bits", 255, 127, 1},
		{"mid value", 100, 100, 0},
		{"mid value high", 200, 72, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsb, msb := encodePaletteColor(tt.value)
			if lsb != tt.wantLSB || msb != tt.wantMSB {
				t.Errorf("encodePaletteColor(%d) = (%d, %d), want (%d, %d)",
					tt.value, lsb, msb, tt.wantLSB, tt.wantMSB)
			}
			// Verify round-trip: reconstructed value matches input.
			got := lsb | (msb << 7)
			if got != tt.value {
				t.Errorf("round-trip: got %d, want %d", got, tt.value)
			}
		})
	}
}

func TestEncodeTouchStripConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  TouchStripConfig
		want uint8
	}{
		{"all false", TouchStripConfig{}, 0},
		{"host control only", TouchStripConfig{HostControl: true}, 0x01},
		{"point + autoreturn center", TouchStripConfig{
			PointMode:      true,
			AutoReturn:     true,
			ReturnToCenter: true,
		}, 0x68},
		{"all true", TouchStripConfig{
			HostControl:    true,
			HostSendsSysEx: true,
			ModWheel:       true,
			PointMode:      true,
			BarFromCenter:  true,
			AutoReturn:     true,
			ReturnToCenter: true,
		}, 0x7F},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeTouchStripConfig(tt.cfg)
			if got != tt.want {
				t.Errorf("encodeTouchStripConfig() = 0x%02X, want 0x%02X", got, tt.want)
			}
		})
	}
}

func TestTouchStripLEDPacking(t *testing.T) {
	// Verify that SetTouchStripLEDs produces correctly packed bytes.
	// We can't call SetTouchStripLEDs without an output port, so test
	// the packing logic directly.
	tests := []struct {
		name string
		leds [31]uint8
		want [16]byte // expected data bytes (after command byte)
	}{
		{
			"all zero",
			[31]uint8{},
			[16]byte{},
		},
		{
			"first LED only",
			func() [31]uint8 {
				var l [31]uint8
				l[0] = 7
				return l
			}(),
			func() [16]byte {
				var b [16]byte
				b[0] = 7 // lo=7, hi=0
				return b
			}(),
		},
		{
			"second LED only",
			func() [31]uint8 {
				var l [31]uint8
				l[1] = 5
				return l
			}(),
			func() [16]byte {
				var b [16]byte
				b[0] = 5 << 3 // lo=0, hi=5
				return b
			}(),
		},
		{
			"last LED (index 30)",
			func() [31]uint8 {
				var l [31]uint8
				l[30] = 3
				return l
			}(),
			func() [16]byte {
				var b [16]byte
				b[15] = 3 // lo=3, hi=0 (only one LED in last byte)
				return b
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the packing logic from SetTouchStripLEDs.
			var got [16]byte
			for i := range 16 {
				var lo, hi uint8
				idx := i * 2
				if idx < 31 {
					lo = tt.leds[idx] & 0x07
				}
				if idx+1 < 31 {
					hi = tt.leds[idx+1] & 0x07
				}
				got[i] = (hi << 3) | lo
			}
			if got != tt.want {
				t.Errorf("packing mismatch:\n got %v\nwant %v", got, tt.want)
			}
		})
	}
}
