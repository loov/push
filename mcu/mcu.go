// Package mcu implements the Mackie Control Universal protocol codec.
//
// It parses raw MIDI bytes into typed messages and encodes typed commands
// back to MIDI bytes. This package has no platform dependencies — everything
// is pure []byte ↔ typed value conversion.
package mcu

// MessageKind identifies the type of an MCU message.
type MessageKind uint8

const (
	MsgUnknown         MessageKind = iota
	MsgButton                      // Note On/Off → button press/release
	MsgFader                       // Pitch Bend → fader position
	MsgVPot                        // CC 16-23 → encoder rotation
	MsgSysEx                       // SysEx → LCD, handshake, meters, etc.
	MsgVPotRing                    // CC 48-55 → V-Pot LED ring state
	MsgJogWheel                    // CC 60 → jog wheel rotation
	MsgTimecode                    // CC 64-73 → 7-segment timecode digit
	MsgAssignment                  // CC 74-75 → 2-digit assignment display
	MsgChannelPressure             // Channel pressure → meter level
)

// String returns a human-readable name for the message kind.
func (k MessageKind) String() string {
	switch k {
	case MsgButton:
		return "Button"
	case MsgFader:
		return "Fader"
	case MsgVPot:
		return "VPot"
	case MsgSysEx:
		return "SysEx"
	case MsgVPotRing:
		return "VPotRing"
	case MsgJogWheel:
		return "JogWheel"
	case MsgTimecode:
		return "Timecode"
	case MsgAssignment:
		return "Assignment"
	case MsgChannelPressure:
		return "ChannelPressure"
	default:
		return "Unknown"
	}
}

// Message is a parsed MCU MIDI message.
type Message struct {
	Kind MessageKind

	// Button fields (Kind == MsgButton)
	Button   Button
	Pressed  bool  // true if velocity > 0 (button down or LED on/blink)
	Velocity uint8 // raw velocity: 0x00=off, odd=blink, 0x7F=on

	// Fader fields (Kind == MsgFader)
	FaderChannel uint8  // 0-7 (channel), 8 (master)
	FaderValue   uint16 // 0-16383

	// VPot fields (Kind == MsgVPot)
	VPotChannel uint8 // 0-7
	VPotDelta   int   // positive = clockwise, negative = counter-clockwise

	// VPotRing fields (Kind == MsgVPotRing)
	VPotRingChannel uint8           // 0-7
	VPotRingMode    VPotRingLEDMode // display mode
	VPotRingValue   uint8           // 0-11
	VPotRingCenter  bool            // center LED on/off

	// JogWheel fields (Kind == MsgJogWheel)
	JogDelta int // positive = clockwise, negative = counter-clockwise

	// Display fields (Kind == MsgTimecode or MsgAssignment)
	DisplayPosition uint8 // digit position (0-9 for timecode, 0-1 for assignment)
	DisplayChar     byte  // 7-segment character code (ASCII 0x30-0x5F range)
	DisplayDot      bool  // dot/decimal point LED

	// SysEx fields (Kind == MsgSysEx)
	SysExData []byte // raw payload between F0 and F7

	// Meter fields (Kind == MsgChannelPressure)
	MeterChannel uint8 // 0-7
	MeterLevel   uint8 // 0-13 (signal level) or MeterSetOverload/MeterClearOverload
}

// LED returns the three-state LED interpretation of the button velocity.
func (m Message) LED() LEDState { return LEDStateFromVelocity(m.Velocity) }

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
			Kind:     MsgButton,
			Button:   Button(data[1]),
			Pressed:  data[2] > 0,
			Velocity: data[2],
		}

	// Note Off (0x80) — button release
	case status == 0x80 && len(data) >= 3:
		return Message{
			Kind:    MsgButton,
			Button:  Button(data[1]),
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

	// Control Change (0xB0 or 0xBF) — timecode display can arrive on channel 15
	case (status == 0xB0 || status == 0xBF) && len(data) >= 3:
		cc := data[1]
		val := data[2]
		switch {
		// CC 16-23: V-Pot rotation (encoder turn) — channel 0 only
		case status == 0xB0 && cc >= 16 && cc <= 23:
			return Message{
				Kind:        MsgVPot,
				VPotChannel: cc - 16,
				VPotDelta:   DecodeRelative(val),
			}
		// CC 48-55: V-Pot LED ring (host → device) — channel 0 only
		case status == 0xB0 && cc >= 48 && cc <= 55:
			return Message{
				Kind:            MsgVPotRing,
				VPotRingChannel: cc - 48,
				VPotRingMode:    VPotRingLEDMode((val >> 4) & 0x03),
				VPotRingValue:   val & 0x0F,
				VPotRingCenter:  val&0x40 != 0,
			}
		// CC 60: Jog wheel rotation — channel 0 only
		case status == 0xB0 && cc == 60:
			return Message{
				Kind:     MsgJogWheel,
				JogDelta: DecodeRelative(val),
			}
		// CC 64-73: Timecode display (7-segment, right-to-left) — channel 0 or 15
		case cc >= 64 && cc <= 73:
			return Message{
				Kind:            MsgTimecode,
				DisplayPosition: cc - 64,
				DisplayChar:     val & 0x3F,
				DisplayDot:      val&0x40 != 0,
			}
		// CC 74-75: Assignment display (2-digit mode indicator) — channel 0 or 15
		case cc >= 74 && cc <= 75:
			return Message{
				Kind:            MsgAssignment,
				DisplayPosition: cc - 74,
				DisplayChar:     val & 0x3F,
				DisplayDot:      val&0x40 != 0,
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

// DecodeRelative converts a sign-magnitude CC value to a signed delta.
// Bit 6 is the sign (0=positive, 1=negative), bits 0-5 are the magnitude.
// Values 1-63 = clockwise, 65-127 = counter-clockwise.
func DecodeRelative(value uint8) int {
	if value < 64 {
		return int(value)
	}
	return -int(value & 0x3F)
}

// EncodeButtonPress creates a Note On message for the given button.
func EncodeButtonPress(button Button) []byte {
	return []byte{0x90, byte(button), 0x7F}
}

// EncodeButtonRelease creates a Note On with velocity 0 for the given button.
func EncodeButtonRelease(button Button) []byte {
	return []byte{0x90, byte(button), 0x00}
}

// EncodeButtonTap sends a press immediately followed by a release.
func EncodeButtonTap(button Button) [][]byte {
	return [][]byte{
		EncodeButtonPress(button),
		EncodeButtonRelease(button),
	}
}

// EncodeFader creates a Pitch Bend message for a fader.
// channel: 0-7 (channel strips), 8 (master)
// value: 0-16383 (14-bit)
func EncodeFader(channel uint8, value uint16) []byte {
	return []byte{
		0xE0 | (channel & 0x0F),
		byte(value & 0x7F),        // LSB
		byte((value >> 7) & 0x7F), // MSB
	}
}

// EncodeVPot creates a CC message for a V-Pot rotation.
// channel: 0-7, delta: positive = CW, negative = CCW
func EncodeVPot(channel uint8, delta int) []byte {
	var val byte
	if delta >= 0 {
		val = byte(delta & 0x3F)
	} else {
		val = 0x40 | byte(-delta&0x3F)
	}
	return []byte{0xB0, 16 + (channel & 0x07), val}
}

// EncodeVPotRing creates a CC 48-55 message for V-Pot LED ring display.
func EncodeVPotRing(channel uint8, mode VPotRingLEDMode, value uint8, center bool) []byte {
	val := byte(mode&0x03)<<4 | byte(value&0x0F)
	if center {
		val |= 0x40
	}
	return []byte{0xB0, 48 + (channel & 0x07), val}
}

// EncodeTimecodeDigit creates a CC 64-73 message for a 7-segment timecode digit.
// position: 0-9 (right-to-left), char: character code (0x00-0x3F), dot: decimal point.
func EncodeTimecodeDigit(position uint8, char byte, dot bool) []byte {
	val := char & 0x3F
	if dot {
		val |= 0x40
	}
	return []byte{0xB0, 64 + (position & 0x0F), val}
}

// EncodeAssignmentDigit creates a CC 74-75 message for the 2-digit assignment display.
// position: 0-1, char: character code (0x00-0x3F), dot: decimal point.
func EncodeAssignmentDigit(position uint8, char byte, dot bool) []byte {
	val := char & 0x3F
	if dot {
		val |= 0x40
	}
	return []byte{0xB0, 74 + (position & 0x01), val}
}

// EncodeLCD creates a SysEx message to update LCD text at the given position.
// modelID: device model ID, position: 0-111, text: ASCII characters to write.
func EncodeLCD(modelID byte, position uint8, text string) []byte {
	msg := []byte{0xF0}
	msg = append(msg, SysExPrefix[:]...)
	msg = append(msg, modelID, cmdLCD, position)
	msg = append(msg, []byte(text)...)
	msg = append(msg, 0xF7)
	return msg
}

// EncodeMeterLevel creates a Channel Pressure message for meter display.
// channel: 0-7, level: 0x0-0xD (signal level) or MeterSetOverload/MeterClearOverload.
func EncodeMeterLevel(channel uint8, level uint8) []byte {
	return []byte{0xD0, (channel&0x07)<<4 | (level & 0x0F)}
}

// EncodeFaderTouch creates a Note On message for a fader touch event.
// channel: 0-7 (strips), 8 (master). touched: true=finger on, false=finger off.
func EncodeFaderTouch(channel uint8, touched bool) []byte {
	vel := byte(0x00)
	if touched {
		vel = 0x7F
	}
	return []byte{0x90, byte(FaderTouch0) + (channel & 0x0F), vel}
}
