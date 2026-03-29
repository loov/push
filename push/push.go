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

// Push3 represents a connected Push 3 device.
type Push3 struct {
	input  *midi.InputPort
	output *midi.OutputPort

	// Event callbacks. Set these before calling Connect.
	OnButton       func(id push3.ButtonID, pressed bool)
	OnPad          func(pos push3.PadPosition, velocity uint8, pressed bool)
	OnEncoder      func(id push3.EncoderID, delta int)
	OnEncoderTouch func(id push3.EncoderID, touched bool)
	OnRawMIDI      func(data []byte)
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
	switch status {
	case 0x90: // Note On
		if len(data) < 3 {
			return
		}
		note := data[1]
		velocity := data[2]
		pressed := velocity > 0

		// Check if it's a pad note (36-99).
		if pos, ok := push3.PadPositionFromNote(note); ok {
			if p.OnPad != nil {
				p.OnPad(pos, velocity, pressed)
			}
			return
		}

		// Check if it's an encoder touch note (0-10).
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
			if p.OnPad != nil {
				p.OnPad(pos, 0, false)
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

	case 0xB0: // Control Change
		if len(data) < 3 {
			return
		}
		cc := data[1]
		value := data[2]

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

