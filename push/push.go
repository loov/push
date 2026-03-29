// Package push provides Push 3 hardware communication over MIDI.
//
// It connects to the Push 3's MIDI ports, dispatches incoming events
// (buttons, pads, encoders) to callbacks, and provides methods to
// set LED colors on buttons and pads.
package push

import (
	"fmt"

	"github.com/loov/logic-push3/midi"
	"github.com/loov/logic-push3/push3"
)

// mpeCCSlide is the CC number for MPE "Slide" (vertical finger position).
const mpeCCSlide = 74

// Push3 represents a connected Push 3 device.
type Push3 struct {
	input  *midi.InputPort
	output *midi.OutputPort

	// MPE channel → pad position tracking.
	// Each pad press gets a unique MIDI channel; we track it so
	// aftertouch and slide events can be routed to the correct pad.
	activePads [16]push3.PadPosition // indexed by MIDI channel
	padActive  [16]bool             // whether channel has an active pad

	// Event callbacks. Set these before calling Connect.
	OnButton          func(id push3.ButtonID, pressed bool)
	OnPad             func(pos push3.PadPosition, velocity uint8, pressed bool)
	OnPadPressure     func(pos push3.PadPosition, pressure uint8)   // Aftertouch (channel pressure per MPE channel)
	OnPadSlide        func(pos push3.PadPosition, value uint8)      // CC 74 — vertical finger position
	OnPadPitchBend    func(pos push3.PadPosition, value uint16)     // MPE pitch bend (0-16383, center 8192)
	OnEncoder         func(id push3.EncoderID, delta int)
	OnEncoderTouch    func(id push3.EncoderID, touched bool)
	OnTouchStrip      func(value uint16)                            // Position 0-16383
	OnTouchStripTouch func(touched bool)                            // Finger on/off
	OnRawMIDI         func(data []byte)
}

// Push 3 MIDI port name patterns.
const (
	SourceName = "Ableton Push 3 Live Port"
	DestName   = "Ableton Push 3 Live Port"
)

// Connect finds the Push 3 MIDI ports and starts listening for events.
func Connect(client *midi.Client, sourceName, destName string) (*Push3, error) {
	source, err := midi.FindSource(sourceName)
	if err != nil {
		return nil, fmt.Errorf("push: finding source: %w", err)
	}
	dest, err := midi.FindDestination(destName)
	if err != nil {
		return nil, fmt.Errorf("push: finding destination: %w", err)
	}

	output, err := client.OpenOutput("push3-out", dest)
	if err != nil {
		return nil, fmt.Errorf("push: opening output: %w", err)
	}

	p := &Push3{output: output}

	input, err := client.OpenInput("push3-in", source, p.handleMIDI)
	if err != nil {
		return nil, fmt.Errorf("push: opening input: %w", err)
	}
	p.input = input

	return p, nil
}

func (p *Push3) handleMIDI(data []byte) {
	if p.OnRawMIDI != nil {
		p.OnRawMIDI(data)
	}
	if len(data) < 2 {
		return
	}

	status := data[0] & 0xF0
	ch := data[0] & 0x0F

	switch status {
	case 0x90: // Note On
		if len(data) < 3 {
			return
		}
		note := data[1]
		velocity := data[2]
		pressed := velocity > 0

		// Check if it's a pad note (36-99). Pads use MPE channels.
		if pos, ok := push3.PadPositionFromNote(note); ok {
			if pressed {
				p.activePads[ch] = pos
				p.padActive[ch] = true
			}
			if p.OnPad != nil {
				p.OnPad(pos, velocity, pressed)
			}
			return
		}

		// Touch strip touch.
		if note == push3.TouchTouchStrip {
			if p.OnTouchStripTouch != nil {
				p.OnTouchStripTouch(pressed)
			}
			return
		}

		// Check if it's an encoder touch note.
		if enc, ok := encoderFromTouchNote(note); ok {
			if p.OnEncoderTouch != nil {
				p.OnEncoderTouch(enc, pressed)
			}
			return
		}

	case 0x80: // Note Off
		if len(data) < 3 {
			return
		}
		note := data[1]

		// Pad release.
		if pos, ok := push3.PadPositionFromNote(note); ok {
			p.padActive[ch] = false
			if p.OnPad != nil {
				p.OnPad(pos, 0, false)
			}
			return
		}

		// Touch strip release.
		if note == push3.TouchTouchStrip {
			if p.OnTouchStripTouch != nil {
				p.OnTouchStripTouch(false)
			}
			return
		}

		// Encoder touch release.
		if enc, ok := encoderFromTouchNote(note); ok {
			if p.OnEncoderTouch != nil {
				p.OnEncoderTouch(enc, false)
			}
			return
		}

	case 0xD0: // Channel Pressure — pad aftertouch (MPE)
		if len(data) < 2 {
			return
		}
		if p.padActive[ch] && p.OnPadPressure != nil {
			p.OnPadPressure(p.activePads[ch], data[1])
		}
		return

	case 0xE0: // Pitch Bend
		if len(data) < 3 {
			return
		}
		value := uint16(data[1]) | uint16(data[2])<<7
		if ch == 0 {
			// Channel 0: touch strip.
			if p.OnTouchStrip != nil {
				p.OnTouchStrip(value)
			}
		} else if p.padActive[ch] && p.OnPadPitchBend != nil {
			// MPE channels: per-pad pitch bend.
			p.OnPadPitchBend(p.activePads[ch], value)
		}
		return

	case 0xB0: // Control Change
		if len(data) < 3 {
			return
		}
		cc := data[1]
		value := data[2]

		// MPE Slide (CC 74) — pad vertical finger position.
		if cc == mpeCCSlide && p.padActive[ch] {
			if p.OnPadSlide != nil {
				p.OnPadSlide(p.activePads[ch], value)
			}
			return
		}

		// Swing/Tempo encoder click (CC 15 sends val=127 press, val=0 release).
		if cc == byte(push3.ButtonSwingTempoPress) {
			if p.OnButton != nil {
				p.OnButton(push3.ButtonSwingTempoPress, value > 0)
			}
			return
		}

		// Encoder rotation.
		if enc, ok := push3.EncoderFromCC(cc); ok {
			if p.OnEncoder != nil {
				p.OnEncoder(enc, push3.DecodeRelative(value))
			}
			return
		}

		// Button press/release (CC-based).
		if isButtonCC(cc) {
			if p.OnButton != nil {
				p.OnButton(push3.ButtonID(cc), value > 0)
			}
			return
		}
	}
}

// Send sends raw MIDI data to the Push 3.
func (p *Push3) Send(data []byte) error {
	return p.output.Send(data)
}

// SendSysEx sends a Push 3 SysEx message.
// prefix is automatically added; data is the payload after the prefix.
func (p *Push3) SendSysEx(data []byte) error {
	msg := []byte{0xF0, 0x00, 0x21, 0x1D, 0x01, 0x01}
	msg = append(msg, data...)
	msg = append(msg, 0xF7)
	return p.output.Send(msg)
}

// SetPadColor sets the LED color of a pad using a palette velocity index.
func (p *Push3) SetPadColor(pos push3.PadPosition, paletteIndex uint8) error {
	note := pos.PadNote()
	return p.output.Send([]byte{0x90, note, paletteIndex})
}

// SetButtonColor sets the LED color of a button using a palette velocity index.
func (p *Push3) SetButtonColor(button push3.ButtonID, paletteIndex uint8) error {
	return p.output.Send([]byte{0xB0, byte(button), paletteIndex})
}

// SetAllPadsColor sets all 64 pads to the same palette color.
func (p *Push3) SetAllPadsColor(paletteIndex uint8) error {
	for row := range uint8(8) {
		for col := range uint8(8) {
			pos := push3.PadPosition{Row: row, Col: col}
			if err := p.SetPadColor(pos, paletteIndex); err != nil {
				return fmt.Errorf("push: setting pad (%d,%d): %w", row, col, err)
			}
		}
	}
	return nil
}

// ClearPads turns off all pad LEDs.
func (p *Push3) ClearPads() error {
	return p.SetAllPadsColor(push3.PaletteBlack)
}

