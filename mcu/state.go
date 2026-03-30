package mcu

import (
	"fmt"
	"sync"
)

// State accumulates MCU state from incoming messages.
// All methods are safe for concurrent use.
type State struct {
	mu sync.Mutex

	Transport     TransportState
	Tracks        [8]TrackState
	LCD           [2]LCDRow
	VPotRing      [8]uint8
	SelectedTrack int // -1 if no track selected
	AssignMode    MCUAssignMode

	// Modifier state from host LED feedback.
	Flip                              bool
	Zoom                              bool
	Scrub                             bool
	ModShift, ModCtrl, ModOpt, ModAlt bool
}

// NewState creates a new MCU state with default values.
func NewState() *State {
	s := &State{
		SelectedTrack: -1,
	}
	// Initialize LCD rows with spaces.
	for i := range s.LCD {
		for j := range s.LCD[i] {
			s.LCD[i][j] = ' '
		}
	}
	return s
}

// Snapshot is an immutable copy of the MCU state for reading without locks.
type Snapshot struct {
	Transport     TransportState
	Tracks        [8]TrackState
	LCD           [2]LCDRow
	VPotRing      [8]uint8
	SelectedTrack int
	AssignMode    MCUAssignMode
	Flip          bool
	Zoom          bool
	Scrub         bool
}

// Snapshot returns an immutable copy of the current state.
func (s *State) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Snapshot{
		Transport:     s.Transport,
		Tracks:        s.Tracks,
		LCD:           s.LCD,
		VPotRing:      s.VPotRing,
		SelectedTrack: s.SelectedTrack,
		AssignMode:    s.AssignMode,
		Flip:          s.Flip,
		Zoom:          s.Zoom,
		Scrub:         s.Scrub,
	}
}

// Handle processes a parsed MCU message and updates the state.
// Returns a human-readable description of the change, or "" if the message was not handled.
func (s *State) Handle(msg Message) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch msg.Kind {
	case MsgButton:
		return s.handleButton(msg)
	case MsgFader:
		return s.handleFader(msg)
	case MsgVPot:
		return s.handleVPot(msg)
	case MsgChannelPressure:
		return s.handleMeter(msg)
	case MsgSysEx:
		return s.handleSysEx(msg)
	}
	return ""
}

func (s *State) handleButton(msg Message) string {
	note := msg.Button
	pressed := msg.Pressed

	switch {
	// Transport
	case note == MCUPlay:
		s.Transport.Play = pressed
		s.Transport.Stop = !pressed
		return fmt.Sprintf("transport: play=%v", pressed)
	case note == MCUStop:
		if pressed {
			s.Transport.Play = false
			s.Transport.Stop = true
		}
		return fmt.Sprintf("transport: stop=%v", pressed)
	case note == MCURecord:
		s.Transport.Record = pressed
		return fmt.Sprintf("transport: record=%v", pressed)
	case note == MCUFastFwd:
		s.Transport.FFwd = pressed
		return fmt.Sprintf("transport: ffwd=%v", pressed)
	case note == MCURewind:
		s.Transport.Rew = pressed
		return fmt.Sprintf("transport: rew=%v", pressed)

	// Channel strip: Rec arm (0-7)
	case note >= MCURecArm0 && note < MCURecArm0+8:
		ch := int(note - MCURecArm0)
		s.Tracks[ch].RecArm = pressed
		return fmt.Sprintf("track[%d]: rec_arm=%v", ch, pressed)

	// Channel strip: Solo (8-15)
	case note >= MCUSolo0 && note < MCUSolo0+8:
		ch := int(note - MCUSolo0)
		s.Tracks[ch].Solo = pressed
		return fmt.Sprintf("track[%d]: solo=%v", ch, pressed)

	// Channel strip: Mute (16-23)
	case note >= MCUMute0 && note < MCUMute0+8:
		ch := int(note - MCUMute0)
		s.Tracks[ch].Mute = pressed
		return fmt.Sprintf("track[%d]: mute=%v", ch, pressed)

	// Channel strip: Select (24-31)
	case note >= MCUSelect0 && note < MCUSelect0+8:
		ch := int(note - MCUSelect0)
		if pressed {
			for i := range s.Tracks {
				s.Tracks[i].Selected = false
			}
			s.Tracks[ch].Selected = true
			s.SelectedTrack = ch
		}
		return fmt.Sprintf("track[%d]: selected=%v", ch, pressed)

	// Modifier LEDs from host
	case note == MCUModShift:
		s.ModShift = pressed
	case note == MCUModCtrl:
		s.ModCtrl = pressed
	case note == MCUModOption:
		s.ModOpt = pressed
	case note == MCUModAlt:
		s.ModAlt = pressed

	// Flip LED
	case note == MCUFlip:
		s.Flip = pressed
		return fmt.Sprintf("flip=%v", pressed)

	// Zoom/Scrub LEDs
	case note == MCUZoom:
		s.Zoom = pressed
	case note == MCUScrub:
		s.Scrub = pressed

	// Assign buttons — detect mode
	case note >= MCUAssignTrack && note <= MCUAssignInstrument:
		if pressed {
			s.AssignMode = DetectModeFromAssign(byte(note))
			return fmt.Sprintf("assign_mode=%s", s.AssignMode)
		}
	}
	return ""
}

func (s *State) handleFader(msg Message) string {
	ch := msg.FaderChannel
	if ch < 8 {
		s.Tracks[ch].FaderLevel = msg.FaderValue
		return fmt.Sprintf("track[%d]: fader=%d", ch, msg.FaderValue)
	}
	return ""
}

func (s *State) handleVPot(msg Message) string {
	// V-Pot rotation updates are informational; the ring display comes via SysEx.
	return fmt.Sprintf("vpot[%d]: delta=%d", msg.VPotChannel, msg.VPotDelta)
}

func (s *State) handleMeter(_ Message) string {
	// Meter level updates for the channel strip.
	return ""
}

func (s *State) handleSysEx(msg Message) string {
	payload := msg.SysExData
	if !IsMCUSysEx(payload) {
		return ""
	}

	switch ClassifySysEx(payload) {
	case SysExLCD:
		pos, text, ok := ParseLCD(payload)
		if !ok {
			return ""
		}
		ApplyLCD(&s.LCD, pos, text)

		// Update track names from top LCD row.
		for i := range 8 {
			name := s.LCD[0].Cell(i)
			if name != s.Tracks[i].Name {
				s.Tracks[i].Name = name
			}
		}

		// Try to detect mode from LCD content.
		detected := DetectMode(s.LCD)
		if detected != MCUAssignModeUnknown {
			s.AssignMode = detected
		}

		return fmt.Sprintf("lcd: pos=%d text=%q", pos, text)

	case SysExVPotRing:
		if len(payload) >= 7 {
			ch := payload[5] & 0x07
			ring := payload[6] & 0x0F
			s.VPotRing[ch] = ring
			return fmt.Sprintf("vpot_ring[%d]=%d", ch, ring)
		}
	}

	return ""
}
