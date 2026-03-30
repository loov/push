package push

import (
	"testing"

	"github.com/loov/push3/push3"
)

func TestIsButtonCC(t *testing.T) {
	tests := []struct {
		name string
		cc   uint8
		want bool
	}{
		{"play", byte(push3.ButtonPlay), true},
		{"record", byte(push3.ButtonRecord), true},
		{"stop", byte(push3.ButtonStopClip), true},
		{"shift", byte(push3.ButtonShift), true},
		{"upper1", byte(push3.ButtonUpper1), true},
		{"lower1", byte(push3.ButtonLower1), true},
		{"mute", byte(push3.ButtonMute), true},
		{"solo", byte(push3.ButtonSolo), true},
		{"div 1/4", byte(push3.ButtonDiv1_4), true},

		// Encoder CCs should NOT be buttons
		{"encoder CC 71", 71, false},
		{"encoder CC 79", 79, false},
		{"encoder CC 14", 14, false},
		{"encoder CC 15 (swing/tempo click)", 15, false},

		// Random CCs
		{"cc 0", 0, false},
		{"cc 127", 127, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isButtonCC(tt.cc); got != tt.want {
				t.Errorf("isButtonCC(%d) = %v, want %v", tt.cc, got, tt.want)
			}
		})
	}
}

func TestEncoderFromTouchNote(t *testing.T) {
	tests := []struct {
		name   string
		note   uint8
		wantID push3.EncoderID
		wantOK bool
	}{
		{"track 1", 0, push3.EncoderTrack1, true},
		{"track 8", 7, push3.EncoderTrack8, true},
		{"volume", 8, push3.EncoderVolume, true},
		{"tempo", 10, push3.EncoderSwingTempo, true},
		{"jog", 11, push3.EncoderJog, true},
		{"unused 9", 9, 0, false},
		{"out of range", 12, 0, false},
		{"pad note", 36, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := encoderFromTouchNote(tt.note)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && id != tt.wantID {
				t.Errorf("id = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestHandleMIDI_Pads(t *testing.T) {
	var gotPos push3.PadPosition
	var gotVel uint8
	var gotPressed bool
	called := false

	p := &Push3{
		OnPad: func(pos push3.PadPosition, velocity uint8, pressed bool) {
			gotPos = pos
			gotVel = velocity
			gotPressed = pressed
			called = true
		},
	}

	// Note On for pad at (0, 0) = note 92, velocity 100
	p.handleMIDI([]byte{0x90, 92, 100})
	if !called {
		t.Fatal("OnPad not called for note 92")
	}
	if gotPos.Row != 0 || gotPos.Col != 0 {
		t.Errorf("position = (%d,%d), want (0,0)", gotPos.Row, gotPos.Col)
	}
	if gotVel != 100 {
		t.Errorf("velocity = %d, want 100", gotVel)
	}
	if !gotPressed {
		t.Error("pressed should be true")
	}

	// Note Off for same pad
	called = false
	p.handleMIDI([]byte{0x80, 92, 0})
	if !called {
		t.Fatal("OnPad not called for note off")
	}
	if gotPressed {
		t.Error("pressed should be false on note off")
	}
}

func TestHandleMIDI_Buttons(t *testing.T) {
	var gotID push3.ButtonID
	var gotPressed bool
	called := false

	p := &Push3{
		OnButton: func(id push3.ButtonID, pressed bool) {
			gotID = id
			gotPressed = pressed
			called = true
		},
	}

	// CC for Play button, value 127 = press
	p.handleMIDI([]byte{0xB0, byte(push3.ButtonPlay), 127})
	if !called {
		t.Fatal("OnButton not called")
	}
	if gotID != push3.ButtonPlay {
		t.Errorf("button = %d, want %d", gotID, push3.ButtonPlay)
	}
	if !gotPressed {
		t.Error("pressed should be true")
	}

	// CC for Play button, value 0 = release
	called = false
	p.handleMIDI([]byte{0xB0, byte(push3.ButtonPlay), 0})
	if !called {
		t.Fatal("OnButton not called for release")
	}
	if gotPressed {
		t.Error("pressed should be false on release")
	}
}

func TestHandleMIDI_Encoders(t *testing.T) {
	var gotID push3.EncoderID
	var gotDelta int
	called := false

	p := &Push3{
		OnEncoder: func(id push3.EncoderID, delta int) {
			gotID = id
			gotDelta = delta
			called = true
		},
	}

	// CC 71 (track 1 encoder), value 3 = clockwise 3
	p.handleMIDI([]byte{0xB0, 71, 3})
	if !called {
		t.Fatal("OnEncoder not called")
	}
	if gotID != push3.EncoderTrack1 {
		t.Errorf("encoder = %v, want %v", gotID, push3.EncoderTrack1)
	}
	if gotDelta != 3 {
		t.Errorf("delta = %d, want 3", gotDelta)
	}

	// CC 71, value 65 = counter-clockwise 1
	called = false
	p.handleMIDI([]byte{0xB0, 71, 65})
	if !called {
		t.Fatal("OnEncoder not called for CCW")
	}
	if gotDelta != -63 {
		t.Errorf("delta = %d, want -63", gotDelta)
	}
}

func TestHandleMIDI_EncoderTouch(t *testing.T) {
	var gotID push3.EncoderID
	var gotTouched bool
	called := false

	p := &Push3{
		OnEncoderTouch: func(id push3.EncoderID, touched bool) {
			gotID = id
			gotTouched = touched
			called = true
		},
	}

	// Note On for encoder touch, note 0 = Track 1
	p.handleMIDI([]byte{0x90, 0, 127})
	if !called {
		t.Fatal("OnEncoderTouch not called")
	}
	if gotID != push3.EncoderTrack1 {
		t.Errorf("encoder = %v, want %v", gotID, push3.EncoderTrack1)
	}
	if !gotTouched {
		t.Error("touched should be true")
	}

	// Note Off for encoder touch
	called = false
	p.handleMIDI([]byte{0x80, 0, 0})
	if !called {
		t.Fatal("OnEncoderTouch not called for release")
	}
	if gotTouched {
		t.Error("touched should be false")
	}
}
