// Package mcu implements the Mackie Control Universal protocol codec.
//
// It parses raw MIDI bytes into typed messages and encodes typed commands
// back to MIDI bytes. This package has no platform dependencies — everything
// is pure []byte ↔ typed value conversion.
package mcu

import "github.com/loov/logic-push3/push3"

// MessageKind identifies the type of an MCU message.
type MessageKind uint8

const (
	MsgUnknown     MessageKind = iota
	MsgButton                  // Note On/Off → button press/release
	MsgFader                   // Pitch Bend → fader position
	MsgVPot                    // CC 16-23 → encoder rotation
	MsgSysEx                   // SysEx → LCD, handshake, meters, etc.
	MsgChannelPressure         // Channel pressure → meter level
)

// Message is a parsed MCU MIDI message.
type Message struct {
	Kind MessageKind

	// Button fields (Kind == MsgButton)
	Button  push3.MCUButton
	Pressed bool

	// Fader fields (Kind == MsgFader)
	FaderChannel uint8  // 0-7 (channel), 8 (master)
	FaderValue   uint16 // 0-16383

	// VPot fields (Kind == MsgVPot)
	VPotChannel uint8 // 0-7
	VPotDelta   int   // positive = clockwise, negative = counter-clockwise

	// SysEx fields (Kind == MsgSysEx)
	SysExData []byte // raw payload between F0 and F7

	// Meter fields (Kind == MsgChannelPressure)
	MeterChannel uint8 // 0-7
	MeterLevel   uint8 // 0-15
}

// Parse converts raw MIDI bytes into an MCU Message.
// Returns MsgUnknown if the message type is not recognized.
func Parse(data []byte) Message {
	if len(data) == 0 {
		return Message{Kind: MsgUnknown}
	}

	status := data[0]

	switch {
	// Note On (0x90) — button press/release
	case status == 0x90 && len(data) >= 3:
		return Message{
			Kind:    MsgButton,
			Button:  push3.MCUButton(data[1]),
			Pressed: data[2] > 0,
		}

	// Note Off (0x80) — button release
	case status == 0x80 && len(data) >= 3:
		return Message{
			Kind:    MsgButton,
			Button:  push3.MCUButton(data[1]),
			Pressed: false,
		}

	// Pitch Bend (0xE0-0xE8) — fader
	case status >= 0xE0 && status <= 0xE8 && len(data) >= 3:
		ch := status - 0xE0
		value := uint16(data[1]) | uint16(data[2])<<7
		return Message{
			Kind:         MsgFader,
			FaderChannel: ch,
			FaderValue:   value,
		}

	// Control Change (0xB0) — V-Pot rotation
	case status == 0xB0 && len(data) >= 3:
		cc := data[1]
		val := data[2]
		if cc >= 16 && cc <= 23 {
			return Message{
				Kind:        MsgVPot,
				VPotChannel: cc - 16,
				VPotDelta:   push3.DecodeRelative(val),
			}
		}
		return Message{Kind: MsgUnknown}

	// Channel Pressure (0xD0) — meter level
	case status == 0xD0 && len(data) >= 2:
		// MCU meters: high nibble = channel, low nibble = level
		ch := data[1] >> 4
		level := data[1] & 0x0F
		return Message{
			Kind:         MsgChannelPressure,
			MeterChannel: ch,
			MeterLevel:   level,
		}

	// SysEx (0xF0)
	case status == 0xF0:
		// Find the payload between F0 and F7
		end := len(data)
		if data[end-1] == 0xF7 {
			end--
		}
		payload := make([]byte, end-1)
		copy(payload, data[1:end])
		return Message{
			Kind:      MsgSysEx,
			SysExData: payload,
		}
	}

	return Message{Kind: MsgUnknown}
}

// EncodeButtonPress creates a Note On message for the given MCU button.
func EncodeButtonPress(button push3.MCUButton) []byte {
	return []byte{0x90, byte(button), 0x7F}
}

// EncodeButtonRelease creates a Note On with velocity 0 for the given MCU button.
func EncodeButtonRelease(button push3.MCUButton) []byte {
	return []byte{0x90, byte(button), 0x00}
}

// EncodeButtonTap sends a press immediately followed by a release.
func EncodeButtonTap(button push3.MCUButton) [][]byte {
	return [][]byte{
		EncodeButtonPress(button),
		EncodeButtonRelease(button),
	}
}

// EncodeFader creates a Pitch Bend message for an MCU fader.
// channel: 0-7 (channel strips), 8 (master)
// value: 0-16383 (14-bit)
func EncodeFader(channel uint8, value uint16) []byte {
	return []byte{
		0xE0 | (channel & 0x0F),
		byte(value & 0x7F),        // LSB
		byte((value >> 7) & 0x7F), // MSB
	}
}

// EncodeVPot creates a CC message for an MCU V-Pot rotation.
// channel: 0-7, delta: positive = CW, negative = CCW
func EncodeVPot(channel uint8, delta int) []byte {
	var val byte
	if delta >= 0 {
		val = byte(delta & 0x3F)
	} else {
		val = byte(128 + delta)
	}
	return []byte{0xB0, 16 + (channel & 0x07), val}
}
