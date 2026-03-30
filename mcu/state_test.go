package mcu

import (
	"testing"

	"github.com/loov/push3/push3"
)

func TestStateTransport(t *testing.T) {
	s := NewState()

	// Play press
	s.Handle(Message{Kind: MsgButton, Button: push3.MCUPlay, Pressed: true})
	snap := s.Snapshot()
	if !snap.Transport.Play {
		t.Error("expected play=true after play press")
	}
	if snap.Transport.Stop {
		t.Error("expected stop=false after play press")
	}

	// Stop press
	s.Handle(Message{Kind: MsgButton, Button: push3.MCUStop, Pressed: true})
	snap = s.Snapshot()
	if snap.Transport.Play {
		t.Error("expected play=false after stop press")
	}
	if !snap.Transport.Stop {
		t.Error("expected stop=true after stop press")
	}

	// Record press
	s.Handle(Message{Kind: MsgButton, Button: push3.MCURecord, Pressed: true})
	snap = s.Snapshot()
	if !snap.Transport.Record {
		t.Error("expected record=true")
	}
}

func TestStateTrackSelect(t *testing.T) {
	s := NewState()

	// Select track 3
	s.Handle(Message{Kind: MsgButton, Button: push3.MCUSelect(3), Pressed: true})
	snap := s.Snapshot()
	if snap.SelectedTrack != 3 {
		t.Errorf("SelectedTrack = %d, want 3", snap.SelectedTrack)
	}
	if !snap.Tracks[3].Selected {
		t.Error("track 3 should be selected")
	}

	// Select track 5 — track 3 should deselect
	s.Handle(Message{Kind: MsgButton, Button: push3.MCUSelect(5), Pressed: true})
	snap = s.Snapshot()
	if snap.SelectedTrack != 5 {
		t.Errorf("SelectedTrack = %d, want 5", snap.SelectedTrack)
	}
	if snap.Tracks[3].Selected {
		t.Error("track 3 should no longer be selected")
	}
}

func TestStateMuteSolo(t *testing.T) {
	s := NewState()

	s.Handle(Message{Kind: MsgButton, Button: push3.MCUMute(2), Pressed: true})
	if !s.Snapshot().Tracks[2].Mute {
		t.Error("expected track 2 mute=true")
	}

	s.Handle(Message{Kind: MsgButton, Button: push3.MCUSolo(4), Pressed: true})
	if !s.Snapshot().Tracks[4].Solo {
		t.Error("expected track 4 solo=true")
	}
}

func TestStateFader(t *testing.T) {
	s := NewState()

	s.Handle(Message{Kind: MsgFader, FaderChannel: 0, FaderValue: 12000})
	if s.Snapshot().Tracks[0].FaderLevel != 12000 {
		t.Errorf("fader level = %d, want 12000", s.Snapshot().Tracks[0].FaderLevel)
	}
}

func TestStateLCD(t *testing.T) {
	s := NewState()

	// Simulate LCD SysEx for track names
	payload := []byte{0x00, 0x00, 0x66, 0x14, 0x12, 0x00}
	payload = append(payload, []byte("Track1 Track2 Bass   Drums  Keys   Pad    Synth  Master ")...)

	s.Handle(Message{Kind: MsgSysEx, SysExData: payload})
	snap := s.Snapshot()

	if snap.Tracks[0].Name != "Track1" {
		t.Errorf("track 0 name = %q, want %q", snap.Tracks[0].Name, "Track1")
	}
	if snap.Tracks[2].Name != "Bass" {
		t.Errorf("track 2 name = %q, want %q", snap.Tracks[2].Name, "Bass")
	}
	if snap.Tracks[7].Name != "Master" {
		t.Errorf("track 7 name = %q, want %q", snap.Tracks[7].Name, "Master")
	}
}

func TestStateInitialLCD(t *testing.T) {
	s := NewState()
	snap := s.Snapshot()
	// LCD should be initialized to spaces
	for i := range snap.LCD {
		for j := range snap.LCD[i] {
			if snap.LCD[i][j] != ' ' {
				t.Fatalf("LCD[%d][%d] = %d, want space", i, j, snap.LCD[i][j])
			}
		}
	}
}
