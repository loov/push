package mcu

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Message
	}{
		{
			name: "note on - button press",
			data: []byte{0x90, 94, 127}, // PLAY press
			want: Message{Kind: MsgButton, Button: Play, Pressed: true},
		},
		{
			name: "note on vel 0 - button release",
			data: []byte{0x90, 94, 0}, // PLAY release
			want: Message{Kind: MsgButton, Button: Play, Pressed: false},
		},
		{
			name: "note off - button release",
			data: []byte{0x80, 93, 0}, // STOP release
			want: Message{Kind: MsgButton, Button: Stop, Pressed: false},
		},
		{
			name: "pitch bend - fader",
			data: []byte{0xE0, 0x00, 0x40}, // channel 0, value 8192
			want: Message{Kind: MsgFader, FaderChannel: 0, FaderValue: 0x40 << 7},
		},
		{
			name: "pitch bend - fader channel 3",
			data: []byte{0xE3, 0x7F, 0x7F}, // channel 3, max value
			want: Message{Kind: MsgFader, FaderChannel: 3, FaderValue: 16383},
		},
		{
			name: "CC vpot clockwise",
			data: []byte{0xB0, 16, 3}, // channel 0, delta +3
			want: Message{Kind: MsgVPot, VPotChannel: 0, VPotDelta: 3},
		},
		{
			name: "CC vpot counter-clockwise",
			data: []byte{0xB0, 17, 65}, // channel 1, 0b01000001 = sign-magnitude -1
			want: Message{Kind: MsgVPot, VPotChannel: 1, VPotDelta: -1},
		},
		{
			name: "channel pressure - meter",
			data: []byte{0xD0, 0x35}, // channel 3, level 5
			want: Message{Kind: MsgChannelPressure, MeterChannel: 3, MeterLevel: 5},
		},
		{
			name: "sysex",
			data: []byte{0xF0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xF7},
			want: Message{Kind: MsgSysEx, SysExData: []byte{0x00, 0x00, 0x66, 0x14, 0x00}},
		},
		{
			name: "empty data",
			data: []byte{},
			want: Message{Kind: MsgUnknown},
		},
		{
			name: "CC vpot ring",
			data: []byte{0xB0, 50, 0x45}, // CC 50 = channel 2, mode=0, value=5, center=true
			want: Message{Kind: MsgVPotRing, VPotRingChannel: 2, VPotRingMode: VPotRingSingle, VPotRingValue: 5, VPotRingCenter: true},
		},
		{
			name: "unknown CC",
			data: []byte{0xB0, 80, 64},
			want: Message{Kind: MsgUnknown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.data)
			if got.Kind != tt.want.Kind {
				t.Errorf("Kind = %v, want %v", got.Kind, tt.want.Kind)
			}
			switch tt.want.Kind {
			case MsgButton:
				if got.Button != tt.want.Button || got.Pressed != tt.want.Pressed {
					t.Errorf("Button = %v/%v, want %v/%v", got.Button, got.Pressed, tt.want.Button, tt.want.Pressed)
				}
			case MsgFader:
				if got.FaderChannel != tt.want.FaderChannel || got.FaderValue != tt.want.FaderValue {
					t.Errorf("Fader = ch%d/%d, want ch%d/%d", got.FaderChannel, got.FaderValue, tt.want.FaderChannel, tt.want.FaderValue)
				}
			case MsgVPot:
				if got.VPotChannel != tt.want.VPotChannel || got.VPotDelta != tt.want.VPotDelta {
					t.Errorf("VPot = ch%d/%d, want ch%d/%d", got.VPotChannel, got.VPotDelta, tt.want.VPotChannel, tt.want.VPotDelta)
				}
			case MsgVPotRing:
				if got.VPotRingChannel != tt.want.VPotRingChannel || got.VPotRingMode != tt.want.VPotRingMode ||
					got.VPotRingValue != tt.want.VPotRingValue || got.VPotRingCenter != tt.want.VPotRingCenter {
					t.Errorf("VPotRing = ch%d/mode%d/val%d/center%v, want ch%d/mode%d/val%d/center%v",
						got.VPotRingChannel, got.VPotRingMode, got.VPotRingValue, got.VPotRingCenter,
						tt.want.VPotRingChannel, tt.want.VPotRingMode, tt.want.VPotRingValue, tt.want.VPotRingCenter)
				}
			case MsgChannelPressure:
				if got.MeterChannel != tt.want.MeterChannel || got.MeterLevel != tt.want.MeterLevel {
					t.Errorf("Meter = ch%d/%d, want ch%d/%d", got.MeterChannel, got.MeterLevel, tt.want.MeterChannel, tt.want.MeterLevel)
				}
			}
		})
	}
}

func TestEncodeButtonPress(t *testing.T) {
	got := EncodeButtonPress(Play)
	want := []byte{0x90, 94, 0x7F}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("byte[%d] = 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestEncodeButtonRelease(t *testing.T) {
	got := EncodeButtonRelease(Stop)
	want := []byte{0x90, 93, 0x00}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("byte[%d] = 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestEncodeFader(t *testing.T) {
	tests := []struct {
		name    string
		channel uint8
		value   uint16
		want    []byte
	}{
		{"ch0 zero", 0, 0, []byte{0xE0, 0x00, 0x00}},
		{"ch0 max", 0, 16383, []byte{0xE0, 0x7F, 0x7F}},
		{"ch3 mid", 3, 8192, []byte{0xE3, 0x00, 0x40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeFader(tt.channel, tt.value)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("byte[%d] = 0x%02X, want 0x%02X", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEncodeVPot(t *testing.T) {
	tests := []struct {
		name    string
		channel uint8
		delta   int
		wantCC  byte
		wantVal byte
	}{
		{"ch0 cw 1", 0, 1, 16, 1},
		{"ch0 cw 10", 0, 10, 16, 10},
		{"ch0 ccw -1", 0, -1, 16, 65},  // 0x40 | 1 = 65
		{"ch0 ccw -10", 0, -10, 16, 74}, // 0x40 | 10 = 74
		{"ch3 cw 5", 3, 5, 19, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeVPot(tt.channel, tt.delta)
			if got[1] != tt.wantCC {
				t.Errorf("CC = %d, want %d", got[1], tt.wantCC)
			}
			if got[2] != tt.wantVal {
				t.Errorf("val = %d, want %d", got[2], tt.wantVal)
			}
		})
	}
}
