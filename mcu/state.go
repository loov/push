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
	VPotRing      [8]VPotRingState
	Timecode      [10]DisplayDigit // 10-digit timecode display (right-to-left)
	Assignment    [2]DisplayDigit  // 2-digit assignment display
	SelectedTrack int              // -1 if no track selected
	AssignMode    AssignMode

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
	VPotRing      [8]VPotRingState
	Timecode      [10]DisplayDigit
	Assignment    [2]DisplayDigit
	SelectedTrack int
	AssignMode    AssignMode
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
		Timecode:      s.Timecode,
		Assignment:    s.Assignment,
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
	case MsgVPotRing:
		return s.handleVPotRing(msg)
	case MsgJogWheel:
		return fmt.Sprintf("jog: delta=%d", msg.JogDelta)
	case MsgTimecode:
		return s.handleTimecode(msg)
	case MsgAssignment:
		return s.handleAssignment(msg)
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
	case note == Play:
		s.Transport.Play = pressed
		s.Transport.Stop = !pressed
		return fmt.Sprintf("transport: play=%v", pressed)
	case note == Stop:
		if pressed {
			s.Transport.Play = false
			s.Transport.Stop = true
		}
		return fmt.Sprintf("transport: stop=%v", pressed)
	case note == Record:
		s.Transport.Record = pressed
		return fmt.Sprintf("transport: record=%v", pressed)
	case note == FastFwd:
		s.Transport.FFwd = pressed
		return fmt.Sprintf("transport: ffwd=%v", pressed)
	case note == Rewind:
		s.Transport.Rew = pressed
		return fmt.Sprintf("transport: rew=%v", pressed)

	// Channel strip: Rec arm (0-7)
	case note >= RecArm0 && note < RecArm0+8:
		ch := int(note - RecArm0)
		s.Tracks[ch].RecArm = pressed
		return fmt.Sprintf("track[%d]: rec_arm=%v", ch, pressed)

	// Channel strip: Solo (8-15)
	case note >= Solo0 && note < Solo0+8:
		ch := int(note - Solo0)
		s.Tracks[ch].Solo = pressed
		return fmt.Sprintf("track[%d]: solo=%v", ch, pressed)

	// Channel strip: Mute (16-23)
	case note >= Mute0 && note < Mute0+8:
		ch := int(note - Mute0)
		s.Tracks[ch].Mute = pressed
		return fmt.Sprintf("track[%d]: mute=%v", ch, pressed)

	// Channel strip: Select (24-31)
	case note >= Select0 && note < Select0+8:
		ch := int(note - Select0)
		if pressed {
			for i := range s.Tracks {
				s.Tracks[i].Selected = false
			}
			s.Tracks[ch].Selected = true
			s.SelectedTrack = ch
		}
		return fmt.Sprintf("track[%d]: selected=%v", ch, pressed)

	// Modifier LEDs from host
	case note == ModShift:
		s.ModShift = pressed
	case note == ModCtrl:
		s.ModCtrl = pressed
	case note == ModOption:
		s.ModOpt = pressed
	case note == ModAlt:
		s.ModAlt = pressed

	// Flip LED
	case note == Flip:
		s.Flip = pressed
		return fmt.Sprintf("flip=%v", pressed)

	// Zoom/Scrub LEDs
	case note == Zoom:
		s.Zoom = pressed
	case note == Scrub:
		s.Scrub = pressed

	// Assign buttons — detect mode
	case note >= AssignTrack && note <= AssignInstrument:
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
	// V-Pot rotation updates are informational; the ring display comes via CC 48-55.
	return fmt.Sprintf("vpot[%d]: delta=%d", msg.VPotChannel, msg.VPotDelta)
}

func (s *State) handleVPotRing(msg Message) string {
	ch := msg.VPotRingChannel
	if ch >= 8 {
		return ""
	}
	s.VPotRing[ch] = VPotRingState{
		Mode:   msg.VPotRingMode,
		Value:  msg.VPotRingValue,
		Center: msg.VPotRingCenter,
	}
	return fmt.Sprintf("vpot_ring[%d]: mode=%s value=%d center=%v", ch, msg.VPotRingMode, msg.VPotRingValue, msg.VPotRingCenter)
}

func (s *State) handleTimecode(msg Message) string {
	pos := msg.DisplayPosition
	if pos >= 10 {
		return ""
	}
	s.Timecode[pos] = DisplayDigit{Char: msg.DisplayChar, Dot: msg.DisplayDot}
	return fmt.Sprintf("timecode[%d]: char=0x%02X dot=%v", pos, msg.DisplayChar, msg.DisplayDot)
}

func (s *State) handleAssignment(msg Message) string {
	pos := msg.DisplayPosition
	if pos >= 2 {
		return ""
	}
	s.Assignment[pos] = DisplayDigit{Char: msg.DisplayChar, Dot: msg.DisplayDot}
	return fmt.Sprintf("assignment[%d]: char=0x%02X dot=%v", pos, msg.DisplayChar, msg.DisplayDot)
}

func (s *State) handleMeter(msg Message) string {
	ch := msg.MeterChannel
	if ch >= 8 {
		return ""
	}
	switch msg.MeterLevel {
	case MeterSetOverload:
		s.Tracks[ch].Overload = true
		return fmt.Sprintf("track[%d]: overload=set", ch)
	case MeterClearOverload:
		s.Tracks[ch].Overload = false
		return fmt.Sprintf("track[%d]: overload=clear", ch)
	default:
		s.Tracks[ch].MeterLevel = msg.MeterLevel
		return fmt.Sprintf("track[%d]: meter=%d", ch, msg.MeterLevel)
	}
}

func (s *State) handleSysEx(msg Message) string {
	payload := msg.SysExData
	if !IsSysEx(payload) {
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
		if detected != AssignModeUnknown {
			s.AssignMode = detected
		}

		return fmt.Sprintf("lcd: pos=%d text=%q", pos, text)

	case SysExMeterMode:
		// Logic Pro repurposes SysEx 0x20 as V-Pot ring echo.
		if len(payload) >= 7 {
			ch := payload[5] & 0x07
			val := payload[6]
			s.VPotRing[ch] = VPotRingState{
				Mode:   VPotRingLEDMode((val >> 4) & 0x03),
				Value:  val & 0x0F,
				Center: val&0x40 != 0,
			}
			return fmt.Sprintf("vpot_ring[%d]: value=%d", ch, val&0x0F)
		}
	}

	return ""
}
